package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"sanoc/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type NotificationItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "INCIDENT_NEW", "INCIDENT_RESOLVED", "FLAP_ALERT", "GATEWAY_DISCONNECTED"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	TargetURL string    `json:"targetUrl"`
	IsUnread  bool      `json:"isUnread"`
	Timestamp time.Time `json:"timestamp"`
}

type NotificationsHandler struct {
	incidentRepo repository.IncidentRepository
	publicMonitorIncidentRepo repository.PublicMonitorIncidentRepository
	notifLogRepo repository.NotificationLogRepository
	mu           sync.RWMutex
	liveNotifs   []NotificationItem
}

func NewNotificationsHandler(incidentRepo repository.IncidentRepository) *NotificationsHandler {
	return &NotificationsHandler{
		incidentRepo: incidentRepo,
		liveNotifs:   make([]NotificationItem, 0),
	}
}

func (h *NotificationsHandler) SetIncidentRepo(repo repository.IncidentRepository) {
	h.incidentRepo = repo
}

func (h *NotificationsHandler) SetNotifLogRepo(repo repository.NotificationLogRepository) {
	h.notifLogRepo = repo
}

func (h *NotificationsHandler) SetPublicMonitorIncidentRepo(repo repository.PublicMonitorIncidentRepository) {
	h.publicMonitorIncidentRepo = repo
}

func (h *NotificationsHandler) AddRealtimeNotification(item NotificationItem) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.liveNotifs = append([]NotificationItem{item}, h.liveNotifs...)
	if len(h.liveNotifs) > 50 {
		h.liveNotifs = h.liveNotifs[:50]
	}
}

// GET /api/v1/notifications
func (h *NotificationsHandler) GetNotifications(c *gin.Context) {
	items := make([]NotificationItem, 0)
	seen := make(map[string]bool)

	if h.incidentRepo != nil {
		incidents, err := h.incidentRepo.GetAll()
		if err == nil {
			for _, inc := range incidents {
				if seen[inc.ID] {
					continue
				}
				seen[inc.ID] = true

				notifType := "INCIDENT_NEW"
				title := fmt.Sprintf("Incident: %s is DOWN", inc.DeviceName)
				msg := fmt.Sprintf("%s (%s) unreachable via ICMP probe", inc.DeviceName, inc.DeviceIP)
				tTime := time.Now()
				if inc.Status == "RESOLVED" {
					notifType = "INCIDENT_RESOLVED"
					title = fmt.Sprintf("Recovered: %s is UP", inc.DeviceName)
					msg = fmt.Sprintf("%s (%s) ping restored (duration: %s)", inc.DeviceName, inc.DeviceIP, inc.Duration)
					if inc.ResolvedAt != "" {
						if parsed, err := time.Parse("2006-01-02 15:04:05", inc.ResolvedAt); err == nil {
							tTime = parsed
						}
					}
				}

				items = append(items, NotificationItem{
					ID:        "notif-" + inc.ID,
					Type:      notifType,
					Title:     title,
					Message:   msg,
					TargetURL: "/devices/" + inc.DeviceID,
					IsUnread:  inc.Status == "ACTIVE",
					Timestamp: tTime,
				})
			}
		}
	}

	if h.publicMonitorIncidentRepo != nil {
		publicIncidents, _, err := h.publicMonitorIncidentRepo.GetAll("", "", "", time.Now().UTC().Add(-30*24*time.Hour), time.Now().UTC(), 1, 50)
		if err == nil {
			for _, inc := range publicIncidents {
				notifType := "PUBLIC_MONITOR_INCIDENT"
				title := fmt.Sprintf("Public monitor down: %s", inc.MonitorName)
				message := inc.LastError
				if message == "" {
					message = "Endpoint tidak merespons sesuai health check"
				}
				if inc.Status == "RESOLVED" {
					notifType = "PUBLIC_MONITOR_RECOVERED"
					title = fmt.Sprintf("Public monitor recovered: %s", inc.MonitorName)
					message = fmt.Sprintf("Endpoint kembali UP setelah %s", formatPublicIncidentDuration(inc.DurationSeconds))
				}
				timestamp := inc.StartedAt
				if inc.Status == "RESOLVED" && inc.ResolvedAt != "" {
					timestamp = inc.ResolvedAt
				}
				tTime, parseErr := time.Parse(time.RFC3339, timestamp)
				if parseErr != nil {
					tTime = time.Now()
				}
				items = append(items, NotificationItem{
					ID: "public-notif-" + inc.ID, Type: notifType, Title: title, Message: message,
					TargetURL: "/incidents/" + inc.ID + "?source=PUBLIC_MONITOR", IsUnread: inc.Status == "ACTIVE", Timestamp: tTime,
				})
			}
		}
	}

	h.mu.RLock()
	for _, liveItem := range h.liveNotifs {
		if !seen[liveItem.ID] {
			seen[liveItem.ID] = true
			items = append(items, liveItem)
		}
	}
	h.mu.RUnlock()

	unreadCount := 0
	formatted := make([]gin.H, len(items))
	for i, item := range items {
		if item.IsUnread {
			unreadCount++
		}
		formatted[i] = gin.H{
			"id":        item.ID,
			"type":      item.Type,
			"title":     item.Title,
			"message":   item.Message,
			"targetUrl": item.TargetURL,
			"isUnread":  item.IsUnread,
			"timestamp": item.Timestamp.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": formatted,
		"unreadCount":   unreadCount,
	})
}

func formatPublicIncidentDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm", seconds/60)
}

// PATCH /api/v1/notifications/read-all
func (h *NotificationsHandler) MarkAllAsRead(c *gin.Context) {
	h.mu.Lock()
	for i := range h.liveNotifs {
		h.liveNotifs[i].IsUnread = false
	}
	h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"success": true, "unreadCount": 0})
}

// DELETE /api/v1/notifications
func (h *NotificationsHandler) ClearNotifications(c *gin.Context) {
	h.mu.Lock()
	h.liveNotifs = make([]NotificationItem, 0)
	h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All notifications cleared"})
}

// GET /api/v1/notifications/logs
func (h *NotificationsHandler) GetNotificationLogs(c *gin.Context) {
	incidentID := c.Query("incidentId")
	if h.notifLogRepo == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	logs, err := h.notifLogRepo.GetByIncidentID(incidentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification logs"})
		return
	}
	paginateSlice(c, logs)
}
