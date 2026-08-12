package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sanoc/backend/internal/domain"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const TypeWhatsAppDispatch = "notification:dispatch"

type WhatsAppDispatchPayload struct {
	Message     string    `json:"message"`
	IncidentIDs []string  `json:"incidentIds"`
	CreatedAt   time.Time `json:"createdAt"`
	TaskType    string    `json:"taskType"` // "DOWN" or "RECOVERED"
	Direct      bool      `json:"direct"`
}

// Lua script to atomically read and advance the next allowed send timestamp in Redis
var rateLimitLuaScript = redis.NewScript(`
local next_allowed_key = KEYS[1]
local now_sec = tonumber(ARGV[1])
local spacing_sec = tonumber(ARGV[2])

local current = redis.call('GET', next_allowed_key)
local next_val = now_sec
if current then
    local current_num = tonumber(current)
    if current_num > now_sec then
        next_val = current_num
    end
end

local new_next = next_val + spacing_sec
redis.call('SET', next_allowed_key, tostring(new_next))
return next_val
`)

// GetNextAllowedTime runs the Lua script to fetch the next timestamp for a message
func GetNextAllowedTime(ctx context.Context, rdb *redis.Client, spacingSec int64) (time.Time, error) {
	nowSec := time.Now().Unix()
	val, err := rateLimitLuaScript.Run(ctx, rdb, []string{"notif:whatsapp:next_allowed_at"}, nowSec, spacingSec).Result()
	if err != nil {
		return time.Time{}, err
	}

	var scheduledSec int64
	switch v := val.(type) {
	case int64:
		scheduledSec = v
	case int:
		scheduledSec = int64(v)
	case string:
		var parsed int64
		if _, pErr := fmt.Sscanf(v, "%d", &parsed); pErr == nil {
			scheduledSec = parsed
		} else {
			return time.Time{}, fmt.Errorf("unexpected string from Lua: %s", v)
		}
	default:
		return time.Time{}, fmt.Errorf("unexpected return type from Lua: %T (%v)", val, val)
	}

	return time.Unix(scheduledSec, 0), nil
}

// HandleWhatsAppDispatchTask returns the handler function for Asynq task execution
func HandleWhatsAppDispatchTask(p *Pipeline) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var pld WhatsAppDispatchPayload
		if err := json.Unmarshal(t.Payload(), &pld); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		// 1. Staleness Cutoff (Requirement 2.3)
		cutoff := 15 * time.Minute
		var taskTime time.Time
		if !pld.CreatedAt.IsZero() {
			taskTime = pld.CreatedAt
		} else {
			// Fallback: look at the first incident's start time
			if len(pld.IncidentIDs) > 0 && p.incidentRepo != nil {
				if inc, err := p.incidentRepo.GetByID(pld.IncidentIDs[0]); err == nil && inc != nil {
					taskTime = inc.StartedAt
					cutoff = 30 * time.Minute // wider threshold for fallback
				}
			}
		}

		if !taskTime.IsZero() && time.Since(taskTime) > cutoff {
			log.Printf("[Notifier] Discarding stale queued notification task (created/started at %v, age %v exceeds limit %v). Message snippet: %q",
				taskTime.Format("15:04:05 WIB"), time.Since(taskTime).Round(time.Second), cutoff, truncateSnippet(pld.Message, 60))
			return nil // Mark done so it's removed from queue
		}

		// 2. Relevance Re-validation (Requirement 2.2)
		isDownAlert := true
		if pld.TaskType == "RECOVERED" {
			isDownAlert = false
		} else if pld.TaskType == "DOWN" {
			isDownAlert = true
		} else {
			// Fallback: parse message text
			msgLower := strings.ToLower(pld.Message)
			if strings.Contains(msgLower, "recovered") || strings.Contains(msgLower, "up") || strings.Contains(msgLower, "online") ||
				strings.Contains(pld.Message, "🟢") || strings.Contains(pld.Message, "✅") || strings.Contains(pld.Message, "✨") {
				isDownAlert = false
			}
		}

		var validIncidentIDs []string
		for _, incID := range pld.IncidentIDs {
			if p.incidentRepo == nil {
				validIncidentIDs = append(validIncidentIDs, incID)
				continue
			}
			inc, err := p.incidentRepo.GetByID(incID)
			if err != nil || inc == nil {
				log.Printf("[Notifier] Incident %s not found in DB, skipping for relevance check", incID)
				continue
			}

			// Load current device status from DB to ensure it matches the alert relevance
			var currentStatus domain.DeviceStatus = domain.StatusUP
			if p.deviceRepo != nil {
				if dev, err := p.deviceRepo.GetByID(inc.DeviceID); err == nil && dev != nil {
					currentStatus = dev.Status
				}
			}

			if isDownAlert {
				// Task was for a DOWN alert.
				// Relevance condition: Incident must still be ACTIVE, and current device status must be DOWN.
				if inc.Status != "ACTIVE" || currentStatus == domain.StatusUP {
					log.Printf("[Notifier] Discarding stale DOWN alert for incident %s: device %s is currently UP/RESOLVED", incID, inc.DeviceName)
					continue
				}
			} else {
				// Task was for an UP/RECOVERY alert.
				// Relevance condition: Incident must be RESOLVED, and the device must NOT have flapped back to DOWN.
				if inc.Status == "ACTIVE" || currentStatus == domain.StatusDOWN {
					log.Printf("[Notifier] Discarding stale RECOVERY alert for incident %s: device %s is currently DOWN/ACTIVE again", incID, inc.DeviceName)
					continue
				}
			}
			validIncidentIDs = append(validIncidentIDs, incID)
		}

		if len(validIncidentIDs) == 0 {
			log.Printf("[Notifier] Discarding task (all %d incidents are no longer relevant)", len(pld.IncidentIDs))
			return nil
		}

		// 3. Aggregation Routing (Requirement 2.4)
		if !pld.Direct && p.notifyQueue != nil {
			log.Printf("[Notifier] Routing %d valid incidents back through Aggregator Queue for consolidation...", len(validIncidentIDs))
			for _, incID := range validIncidentIDs {
				inc, err := p.incidentRepo.GetByID(incID)
				if err != nil || inc == nil {
					continue
				}

				evtType := EventDown
				if !isDownAlert {
					evtType = EventRecovered
				}

				devType := "Access Point"
				ip := inc.DeviceIP
				location := "Main Area"
				if p.deviceRepo != nil {
					if dev, err := p.deviceRepo.GetByID(inc.DeviceID); err == nil && dev != nil {
						devType = string(dev.Type)
						ip = dev.IP
						location = dev.Location
					}
				}
				if location == "" {
					location = "Main Area"
				}

				p.notifyQueue.Enqueue(AggregatorEvent{
					IncidentID: inc.ID,
					DeviceID:   inc.DeviceID,
					DeviceName: inc.DeviceName,
					DeviceType: devType,
					IP:         ip,
					Location:   location,
					EventType:  evtType,
					Timestamp:  time.Now(),
				})
			}
			return nil // Re-queued successfully, complete the current unaggregated task!
		}

		log.Printf("[Asynq Worker] Executing task %s, payload len=%d", t.Type(), len(pld.Message))

		err := p.dispatchActualWithErr(ctx, pld.Message, validIncidentIDs)
		if err != nil {
			log.Printf("[Asynq Worker] Task execution failed: %v", err)
			return err // Return non-nil error so Asynq schedules automatic retry
		}

		log.Printf("[Asynq Worker] Task completed successfully")
		return nil
	}
}

func truncateSnippet(s string, limit int) string {
	runes := []rune(s)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return s
}
