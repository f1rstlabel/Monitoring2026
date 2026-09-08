package handler

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/monitoring"

	"github.com/gin-gonic/gin"
)

type publicMonitorInput struct {
	Name            string                   `json:"name"`
	MonitorType     domain.PublicMonitorType `json:"monitorType"`
	TargetURL       string                   `json:"targetUrl"`
	TargetHost      string                   `json:"targetHost"`
	TargetPort      int                      `json:"targetPort"`
	DNSRecordType   string                   `json:"dnsRecordType"`
	Keyword         string                   `json:"keyword"`
	JSONPath        string                   `json:"jsonPath"`
	ExpectedValue   string                   `json:"expectedValue"`
	GroupID         string                   `json:"groupId"`
	GroupName       string                   `json:"groupName"`
	IntervalSeconds int                      `json:"intervalSeconds"`
	TimeoutSeconds  int                      `json:"timeoutSeconds"`
	RetryCount      int                      `json:"retryCount"`
	Enabled         *bool                    `json:"enabled"`
	NotifyOnFailure *bool                    `json:"notifyOnFailure"`
}

type publicMonitorUpdateInput struct {
	Name            *string                   `json:"name"`
	MonitorType     *domain.PublicMonitorType `json:"monitorType"`
	TargetURL       *string                   `json:"targetUrl"`
	TargetHost      *string                   `json:"targetHost"`
	TargetPort      *int                      `json:"targetPort"`
	DNSRecordType   *string                   `json:"dnsRecordType"`
	Keyword         *string                   `json:"keyword"`
	JSONPath        *string                   `json:"jsonPath"`
	ExpectedValue   *string                   `json:"expectedValue"`
	GroupID         *string                   `json:"groupId"`
	GroupName       *string                   `json:"groupName"`
	IntervalSeconds *int                      `json:"intervalSeconds"`
	TimeoutSeconds  *int                      `json:"timeoutSeconds"`
	RetryCount      *int                      `json:"retryCount"`
	Enabled         *bool                     `json:"enabled"`
	NotifyOnFailure *bool                     `json:"notifyOnFailure"`
}

type publicMonitorDeleteInput struct {
	Reason string `json:"reason"`
}

type publicMonitorPurgeInput struct {
	Reason       string `json:"reason"`
	Confirmation string `json:"confirmation"`
}

func normalizePublicMonitorInput(input publicMonitorInput) (publicMonitorInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.TargetURL = strings.TrimSpace(input.TargetURL)
	input.TargetHost = strings.TrimSpace(strings.Trim(input.TargetHost, "[]"))
	input.DNSRecordType = strings.ToUpper(strings.TrimSpace(input.DNSRecordType))
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.JSONPath = strings.TrimSpace(input.JSONPath)
	input.ExpectedValue = strings.TrimSpace(input.ExpectedValue)
	input.GroupID = strings.TrimSpace(input.GroupID)
	input.GroupName = strings.TrimSpace(input.GroupName)
	if input.MonitorType == "" {
		input.MonitorType = domain.PublicMonitorHTTP
	}
	if input.GroupName == "" {
		input.GroupName = "Ungrouped"
	}
	if input.IntervalSeconds == 0 {
		input.IntervalSeconds = 60
	}
	if input.TimeoutSeconds == 0 {
		input.TimeoutSeconds = 10
	}
	if input.IntervalSeconds < 15 || input.IntervalSeconds > 86400 {
		return input, errors.New("interval must be between 15 seconds and 24 hours")
	}
	if input.TimeoutSeconds < 1 || input.TimeoutSeconds > 120 {
		return input, errors.New("timeout must be between 1 and 120 seconds")
	}
	if input.RetryCount < 0 || input.RetryCount > 5 {
		return input, errors.New("retry count must be between 0 and 5")
	}

	switch input.MonitorType {
	case domain.PublicMonitorHTTP, domain.PublicMonitorHTTPKeyword, domain.PublicMonitorHTTPJSON:
		if input.TargetURL == "" {
			return input, errors.New("target URL is required for HTTP monitors")
		}
		if err := monitoring.ValidateTargetURL(input.TargetURL); err != nil {
			return input, err
		}
		if input.MonitorType == domain.PublicMonitorHTTPKeyword && input.Keyword == "" {
			return input, errors.New("keyword is required for HTTP Keyword monitors")
		}
		if input.MonitorType == domain.PublicMonitorHTTPJSON && (input.JSONPath == "" || input.ExpectedValue == "") {
			return input, errors.New("JSON path and expected value are required for HTTP JSON monitors")
		}
		input.TargetHost, input.TargetPort, input.DNSRecordType = "", 0, ""
		if input.MonitorType != domain.PublicMonitorHTTPKeyword {
			input.Keyword = ""
		}
		if input.MonitorType != domain.PublicMonitorHTTPJSON {
			input.JSONPath, input.ExpectedValue = "", ""
		}
	case domain.PublicMonitorTCP:
		if input.TargetHost == "" {
			return input, errors.New("target host is required for TCP monitors")
		}
		if input.TargetPort < 1 || input.TargetPort > 65535 {
			return input, errors.New("TCP port must be between 1 and 65535")
		}
		if err := monitoring.ValidatePublicHost(input.TargetHost); err != nil {
			return input, err
		}
		input.TargetURL = "tcp://" + formatHostPort(input.TargetHost, input.TargetPort)
		input.DNSRecordType, input.Keyword, input.JSONPath, input.ExpectedValue = "", "", "", ""
	case domain.PublicMonitorPing:
		if input.TargetHost == "" {
			return input, errors.New("target host is required for Ping monitors")
		}
		if err := monitoring.ValidatePublicHost(input.TargetHost); err != nil {
			return input, err
		}
		input.TargetURL = "ping://" + input.TargetHost
		input.TargetPort, input.DNSRecordType = 0, ""
		input.Keyword, input.JSONPath, input.ExpectedValue = "", "", ""
	case domain.PublicMonitorDNS:
		if input.TargetHost == "" {
			return input, errors.New("target host is required for DNS monitors")
		}
		if err := monitoring.ValidatePublicHost(input.TargetHost); err != nil {
			return input, err
		}
		if input.DNSRecordType == "" {
			input.DNSRecordType = "A"
		}
		if input.DNSRecordType != "A" && input.DNSRecordType != "AAAA" && input.DNSRecordType != "CNAME" {
			return input, errors.New("DNS record type must be A, AAAA, or CNAME")
		}
		input.TargetURL = "dns://" + input.TargetHost + "/" + input.DNSRecordType
		input.TargetPort = 0
		input.Keyword, input.JSONPath, input.ExpectedValue = "", "", ""
	default:
		return input, errors.New("unsupported monitor type; use HTTP, HTTP Keyword, HTTP JSON, TCP, Ping, or DNS")
	}
	return input, nil
}

func formatHostPort(host string, port int) string {
	if strings.Contains(host, ":") {
		return "[" + host + "]:" + strconv.Itoa(port)
	}
	return host + ":" + strconv.Itoa(port)
}

func (h *Handler) validatePublicMonitorGroup(input *publicMonitorInput) error {
	if input.GroupID == "" || h.publicMonitorGroupRepo == nil {
		return nil
	}
	group, err := h.publicMonitorGroupRepo.GetByID(input.GroupID)
	if errors.Is(err, sql.ErrNoRows) || group == nil {
		return errors.New("public monitor group not found")
	}
	if err != nil {
		return errors.New("failed to validate public monitor group")
	}
	if !group.Enabled {
		return errors.New("public monitor group is disabled")
	}
	input.GroupName = group.Name
	return nil
}

func (h *Handler) GetPublicMonitors(c *gin.Context) {
	if h.publicMonitorRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitoring is not initialized"})
		return
	}
	page, limit := parsePositiveInt(c.Query("page"), 1), parsePositiveInt(c.Query("page_size"), 25)
	if limit > 100 {
		limit = 100
	}
	items, total, err := h.publicMonitorRepo.GetAll(c.Query("search"), c.Query("status"), c.Query("group"), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitors"})
		return
	}
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "data": items, "total": total, "page": page, "pageSize": limit, "totalPages": totalPages})
}

func (h *Handler) GetArchivedPublicMonitors(c *gin.Context) {
	if h.publicMonitorRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitoring is not initialized"})
		return
	}
	page, limit := parsePositiveInt(c.Query("page"), 1), parsePositiveInt(c.Query("page_size"), 25)
	if limit > 100 {
		limit = 100
	}
	items, total, err := h.publicMonitorRepo.GetArchived(c.Query("search"), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch archived public monitors"})
		return
	}
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "data": items, "total": total, "page": page, "pageSize": limit, "totalPages": totalPages})
}

func (h *Handler) GetPublicMonitorByID(c *gin.Context) {
	monitor, err := h.publicMonitorRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || monitor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor"})
		return
	}
	checks, err := h.publicMonitorRepo.GetChecks(monitor.ID, time.Now().Add(-90*24*time.Hour), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitor history"})
		return
	}
	events, err := h.publicMonitorRepo.GetEvents(monitor.ID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitor events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"monitor": monitor, "checks": checks, "events": events})
}

func (h *Handler) GetPublicMonitorChecks(c *gin.Context) {
	if h.publicMonitorRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitoring is not initialized"})
		return
	}
	monitorID := c.Param("id")
	if _, err := h.publicMonitorRepo.GetByID(monitorID); errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor"})
		return
	}
	page, limit := parsePositiveInt(c.Query("page"), 1), parsePositiveInt(c.Query("page_size"), 10)
	if limit > 100 {
		limit = 100
	}
	since := time.Now().UTC().Add(-90 * 24 * time.Hour)
	items, total, err := h.publicMonitorRepo.GetChecksPage(monitorID, since, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitor check history"})
		return
	}
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "data": items, "total": total, "page": page, "pageSize": limit, "totalPages": totalPages})
}

func (h *Handler) CreatePublicMonitor(c *gin.Context) {
	var input publicMonitorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor payload"})
		return
	}
	var err error
	input, err = normalizePublicMonitorInput(input)
	if err == nil {
		err = h.validatePublicMonitorGroup(&input)
	}
	if err != nil || input.Name == "" {
		if err == nil {
			err = errors.New("name is required")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	enabled, notify := true, false
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if input.NotifyOnFailure != nil {
		notify = *input.NotifyOnFailure
	}
	userID, _ := c.Get("userID")
	monitor := &domain.PublicMonitor{Name: input.Name, MonitorType: input.MonitorType, TargetURL: input.TargetURL, TargetHost: input.TargetHost, TargetPort: input.TargetPort, DNSRecordType: input.DNSRecordType, GroupID: input.GroupID, GroupName: input.GroupName, IntervalSeconds: input.IntervalSeconds, TimeoutSeconds: input.TimeoutSeconds, RetryCount: input.RetryCount, Enabled: enabled, NotifyOnFailure: notify, Status: domain.PublicMonitorPaused}
	if id, ok := userID.(string); ok {
		monitor.CreatedByUserID = id
	}
	created, err := h.publicMonitorRepo.Create(monitor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create public monitor"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdatePublicMonitor(c *gin.Context) {
	current, err := h.publicMonitorRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || current == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor"})
		return
	}
	var input publicMonitorUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor payload"})
		return
	}
	merged := publicMonitorInput{Name: current.Name, MonitorType: current.MonitorType, TargetURL: current.TargetURL, TargetHost: current.TargetHost, TargetPort: current.TargetPort, DNSRecordType: current.DNSRecordType, Keyword: current.Keyword, JSONPath: current.JSONPath, ExpectedValue: current.ExpectedValue, GroupID: current.GroupID, GroupName: current.GroupName, IntervalSeconds: current.IntervalSeconds, TimeoutSeconds: current.TimeoutSeconds, RetryCount: current.RetryCount}
	if input.Name != nil {
		merged.Name = strings.TrimSpace(*input.Name)
	}
	if input.MonitorType != nil {
		merged.MonitorType = *input.MonitorType
	}
	if input.TargetURL != nil {
		merged.TargetURL = strings.TrimSpace(*input.TargetURL)
	}
	if input.TargetHost != nil {
		merged.TargetHost = strings.TrimSpace(*input.TargetHost)
	}
	if input.TargetPort != nil {
		merged.TargetPort = *input.TargetPort
	}
	if input.DNSRecordType != nil {
		merged.DNSRecordType = *input.DNSRecordType
	}
	if input.Keyword != nil {
		merged.Keyword = *input.Keyword
	}
	if input.JSONPath != nil {
		merged.JSONPath = *input.JSONPath
	}
	if input.ExpectedValue != nil {
		merged.ExpectedValue = *input.ExpectedValue
	}
	if input.GroupID != nil {
		merged.GroupID = strings.TrimSpace(*input.GroupID)
	}
	if input.GroupName != nil {
		merged.GroupName = strings.TrimSpace(*input.GroupName)
	}
	if input.IntervalSeconds != nil {
		merged.IntervalSeconds = *input.IntervalSeconds
	}
	if input.TimeoutSeconds != nil {
		merged.TimeoutSeconds = *input.TimeoutSeconds
	}
	if input.RetryCount != nil {
		merged.RetryCount = *input.RetryCount
	}
	merged, err = normalizePublicMonitorInput(merged)
	if err == nil {
		err = h.validatePublicMonitorGroup(&merged)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if merged.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	current.Name, current.MonitorType, current.TargetURL = merged.Name, merged.MonitorType, merged.TargetURL
	current.TargetHost, current.TargetPort, current.DNSRecordType = merged.TargetHost, merged.TargetPort, merged.DNSRecordType
	current.Keyword, current.JSONPath, current.ExpectedValue = merged.Keyword, merged.JSONPath, merged.ExpectedValue
	current.GroupID, current.GroupName = merged.GroupID, merged.GroupName
	current.IntervalSeconds, current.TimeoutSeconds, current.RetryCount = merged.IntervalSeconds, merged.TimeoutSeconds, merged.RetryCount
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	if input.NotifyOnFailure != nil {
		current.NotifyOnFailure = *input.NotifyOnFailure
	}
	if !current.Enabled {
		current.Status = domain.PublicMonitorPaused
	}
	if err := h.publicMonitorRepo.Update(current); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update public monitor"})
		return
	}
	if !current.Enabled && h.publicMonitorIncidentRepo != nil {
		if err := h.publicMonitorIncidentRepo.ResolveForPause(current.ID, time.Now().UTC()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Monitor paused but active incident could not be resolved"})
			return
		}
	}
	updated, err := h.publicMonitorRepo.GetByID(current.ID)
	if err != nil {
		c.JSON(http.StatusOK, current)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeletePublicMonitor(c *gin.Context) {
	if h.publicMonitorRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitoring is not initialized"})
		return
	}
	monitor, err := h.publicMonitorRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || monitor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor"})
		return
	}
	var input publicMonitorDeleteInput
	if bindErr := c.ShouldBindJSON(&input); bindErr != nil && !errors.Is(bindErr, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delete payload"})
		return
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "Deleted by user"
	}
	if len(reason) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Archive reason must be 500 characters or fewer"})
		return
	}
	userID, _ := c.Get("userID")
	deletedBy := ""
	if value, ok := userID.(string); ok {
		deletedBy = value
	}
	if err := h.publicMonitorRepo.Archive(monitor.ID, deletedBy, reason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor not found or already archived"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete public monitor"})
		return
	}
	if h.publicMonitorIncidentRepo != nil {
		if err := h.publicMonitorIncidentRepo.ResolveForDeletion(monitor.ID, time.Now().UTC()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Monitor archived but active incident could not be closed"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) PurgePublicMonitor(c *gin.Context) {
	if h.publicMonitorRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitoring is not initialized"})
		return
	}
	monitor, err := h.publicMonitorRepo.GetByIDIncludingArchived(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || monitor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archived public monitor not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch archived public monitor"})
		return
	}
	if monitor.DeletedAt == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "Archive the monitor before removing its configuration"})
		return
	}
	if monitor.PurgedAt != "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monitor configuration is already removed"})
		return
	}
	var input publicMonitorPurgeInput
	if bindErr := c.ShouldBindJSON(&input); bindErr != nil && !errors.Is(bindErr, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid removal payload"})
		return
	}
	if !strings.EqualFold(strings.TrimSpace(input.Confirmation), strings.TrimSpace(monitor.Name)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type the monitor name exactly to remove its configuration"})
		return
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "Monitor configuration removed"
	}
	if len(reason) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Removal reason must be 500 characters or fewer"})
		return
	}
	userID, _ := c.Get("userID")
	purgedBy := ""
	if value, ok := userID.(string); ok {
		purgedBy = value
	}
	if err := h.publicMonitorRepo.Purge(monitor.ID, purgedBy, reason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Monitor configuration not found or already removed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove monitor configuration"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) RestorePublicMonitor(c *gin.Context) {
	if h.publicMonitorRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitoring is not initialized"})
		return
	}
	monitor, err := h.publicMonitorRepo.GetByIDIncludingArchived(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || monitor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archived public monitor not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch archived public monitor"})
		return
	}
	if monitor.DeletedAt == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "Public monitor is already active"})
		return
	}
	if err := h.publicMonitorRepo.Restore(monitor.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore public monitor"})
		return
	}
	updated, err := h.publicMonitorRepo.GetByID(monitor.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) CheckPublicMonitorNow(c *gin.Context) {
	monitor, err := h.publicMonitorRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || monitor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor"})
		return
	}
	if !monitor.Enabled {
		c.JSON(http.StatusConflict, gin.H{"error": "Enable the monitor before running a check"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(monitor.TimeoutSeconds+2)*time.Second)
	defer cancel()
	var result monitoring.CheckResult
	if h.publicMonitorWorker != nil {
		result, err = h.publicMonitorWorker.CheckNow(ctx, *monitor)
	} else {
		result = monitoring.RunPublicMonitorCheck(ctx, *monitor)
		_, _, err = h.publicMonitorRepo.RecordCheck(monitor.ID, result.Status, result.StatusCode, result.LatencyMs, result.Error, time.Now().UTC())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record monitor check", "detail": err.Error(), "result": result})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": result.Status, "statusCode": result.StatusCode, "latencyMs": result.LatencyMs, "error": result.Error})
}

func parsePositiveInt(raw string, fallback int) int {
	value := fallback
	if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
		value = parsed
	}
	return value
}
