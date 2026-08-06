package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"govmonitor-it/backend/internal/repository"

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
