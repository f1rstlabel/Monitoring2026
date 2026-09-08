package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sanoc/backend/internal/domain"
)

func (h *Handler) GetPublicMonitorIncidents(c *gin.Context) {
	if h.publicMonitorIncidentRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor incidents are not initialized"})
		return
	}
	from, to, err := publicMonitorWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if status == "ALL" {
		status = ""
	}
	if status != "" && status != string(domain.PublicMonitorIncidentActive) && status != string(domain.PublicMonitorIncidentResolved) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be ACTIVE, RESOLVED, or ALL"})
		return
	}
	page := parsePositiveInt(c.Query("page"), 1)
	limit := parsePositiveInt(c.Query("page_size"), 25)
	if limit > 100 {
		limit = 100
	}
	items, total, err := h.publicMonitorIncidentRepo.GetAll(c.Query("monitorId"), status, c.Query("search"), from, to, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor incidents"})
		return
	}
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "data": items, "total": total, "page": page, "pageSize": limit, "totalPages": totalPages})
}

func (h *Handler) GetPublicMonitorIncidentByID(c *gin.Context) {
	if h.publicMonitorIncidentRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor incidents are not initialized"})
		return
	}
	incident, err := h.publicMonitorIncidentRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || incident == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor incident not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor incident"})
		return
	}
	events, err := h.publicMonitorIncidentRepo.GetEvents(incident.ID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor incident events"})
		return
	}
	notifications, err := h.publicMonitorIncidentRepo.GetNotificationLogs(incident.ID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor notification logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"incident": incident, "events": events, "notifications": notifications})
}

func (h *Handler) GetPublicMonitorReport(c *gin.Context) {
	if h.publicMonitorIncidentRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor reports are not initialized"})
		return
	}
	from, to, err := publicMonitorWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.publicMonitorIncidentRepo.GetReport(c.Query("monitorId"), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate public monitor report"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func publicMonitorWindow(c *gin.Context) (time.Time, time.Time, error) {
	loc := time.FixedZone("WIB", 7*60*60)
	now := time.Now().In(loc)
	period := strings.ToLower(strings.TrimSpace(c.Query("period")))
	if period == "" || period == "monthly" {
		return now.Add(-30 * 24 * time.Hour).UTC(), now.UTC(), nil
	}
	switch period {
	case "daily":
		return now.Add(-24 * time.Hour).UTC(), now.UTC(), nil
	case "weekly":
		return now.Add(-7 * 24 * time.Hour).UTC(), now.UTC(), nil
	case "custom":
		startRaw := strings.TrimSpace(c.Query("startDate"))
		endRaw := strings.TrimSpace(c.Query("endDate"))
		if startRaw == "" || endRaw == "" {
			return time.Time{}, time.Time{}, errors.New("startDate and endDate are required for custom period")
		}
		start, err := time.ParseInLocation("2006-01-02", startRaw, loc)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("startDate must use YYYY-MM-DD")
		}
		end, err := time.ParseInLocation("2006-01-02", endRaw, loc)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("endDate must use YYYY-MM-DD")
		}
		end = end.Add(24*time.Hour - time.Nanosecond)
		if !end.After(start) {
			return time.Time{}, time.Time{}, errors.New("endDate must be on or after startDate")
		}
		return start.UTC(), end.UTC(), nil
	default:
		return time.Time{}, time.Time{}, errors.New("period must be daily, weekly, monthly, or custom")
	}
}
