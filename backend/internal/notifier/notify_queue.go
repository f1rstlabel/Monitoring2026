package notifier

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"sanoc/backend/internal/domain"
)

type EventType string

const (
	EventDown      EventType = "DOWN"
	EventRecovered EventType = "RECOVERED"
)

type AggregatorEvent struct {
	IncidentID string
	DeviceID   string
	DeviceName string
	DeviceType string
	IP         string
	Location   string
	Reason     string
	EventType  EventType
	Timestamp  time.Time
}

type NotifyQueue struct {
	pipeline *Pipeline
	repo     domainNotificationLogRepo

	mu        sync.Mutex
	pending   []AggregatorEvent
	aggTimer  *time.Timer
	aggWindow time.Duration
}

type domainNotificationLogRepo interface {
	Append(row *domain.NotificationLogRow) error
}

func NewNotifyQueue(pipeline *Pipeline, repo domainNotificationLogRepo) *NotifyQueue {
	return &NotifyQueue{
		pipeline:  pipeline,
		repo:      repo,
		aggWindow: 5 * time.Second,
	}
}

func (nq *NotifyQueue) Enqueue(evt AggregatorEvent) {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	nq.pending = append(nq.pending, evt)
	if nq.aggTimer == nil {
		nq.aggTimer = time.AfterFunc(nq.aggWindow, nq.flushPending)
	}
}

func (nq *NotifyQueue) flushPending() {
	nq.mu.Lock()
	events := nq.pending
	nq.pending = nil
	nq.aggTimer = nil
	nq.mu.Unlock()

	if len(events) == 0 {
		return
	}

	// Group events by EventType (DOWN / RECOVERED) and Location
	downEvents := make(map[string][]AggregatorEvent)
	upEvents := make(map[string][]AggregatorEvent)

	for _, e := range events {
		loc := e.Location
		if loc == "" {
			loc = "Main Area"
		}
		if e.EventType == EventDown {
			downEvents[loc] = append(downEvents[loc], e)
		} else {
			upEvents[loc] = append(upEvents[loc], e)
		}
	}

	// Formulate messages
	for loc, rawEvts := range downEvents {
		evts := dedupEvents(rawEvts)
		var msg string
		tStr := evts[0].Timestamp.Format("15:04:05 WIB")
		if len(evts) == 1 {
			e := evts[0]
			reason := e.Reason
			if reason == "" {
				reason = "ICMP timeout — no reply after 3 consecutive checks"
			}
			data := SingleEventData{
				DeviceName: e.DeviceName,
				DeviceType: e.DeviceType,
				IP:         e.IP,
				Location:   loc,
				Reason:     reason,
				Timestamp:  tStr,
			}
			indices.mu.Lock()
			idx := indices.singleDown % len(singleDownTemplates)
			indices.singleDown++
			indices.mu.Unlock()
			msg = renderTemplate(singleDownTemplates[idx], data)
		} else {
			limit := 15
			var listDevs []AggregatorEvent
			if len(evts) > limit {
				listDevs = evts[:limit]
			} else {
				listDevs = evts
			}

			var sb strings.Builder
			for _, e := range listDevs {
				reason := e.Reason
				if reason == "" {
					reason = "ICMP timeout — no reply after 3 consecutive checks"
				}
				sb.WriteString(fmt.Sprintf("• <b>%s</b> (%s) — <code>%s</code> (%s)\n", e.DeviceName, e.DeviceType, e.IP, reason))
			}
			if len(evts) > limit {
				sb.WriteString(fmt.Sprintf("\n• <i>...and %d more devices (check the NOC dashboard)</i>\n", len(evts)-limit))
			}

			data := AggregatedEventData{
				Count:      len(evts),
				Location:   loc,
				Timestamp:  tStr,
				DeviceList: sb.String(),
			}
			indices.mu.Lock()
			idx := indices.aggDown % len(aggDownTemplates)
			indices.aggDown++
			indices.mu.Unlock()
			msg = renderTemplate(aggDownTemplates[idx], data)
		}

		var incidentIDs []string
		for _, e := range evts {
			if e.IncidentID != "" {
				incidentIDs = append(incidentIDs, e.IncidentID)
				
				// Write aggregation event
				var detail string
				if len(evts) == 1 {
					detail = "Sent individually (only device DOWN in location)"
				} else {
					detail = fmt.Sprintf("Grouped with %d other devices in %s — aggregation window closed", len(evts)-1, loc)
				}
				nq.pipeline.writeEvent(e.IncidentID, "notification_aggregated", "", detail)
			}
		}
		nq.pipeline.SendExt(context.Background(), msg, incidentIDs, true)
	}

	for loc, rawEvts := range upEvents {
		evts := dedupEvents(rawEvts)
		var msg string
		tStr := evts[0].Timestamp.Format("15:04:05 WIB")
		if len(evts) == 1 {
			e := evts[0]
			reason := e.Reason
			if reason == "" {
				reason = "Connectivity restored — ICMP reply received"
			}
			data := SingleEventData{
				DeviceName: e.DeviceName,
				DeviceType: e.DeviceType,
				IP:         e.IP,
				Location:   loc,
				Reason:     reason,
				Timestamp:  tStr,
			}
			indices.mu.Lock()
			idx := indices.singleUp % len(singleRecoveryTemplates)
			indices.singleUp++
			indices.mu.Unlock()
			msg = renderTemplate(singleRecoveryTemplates[idx], data)
		} else {
			limit := 15
			var listDevs []AggregatorEvent
			if len(evts) > limit {
				listDevs = evts[:limit]
			} else {
				listDevs = evts
			}

			var sb strings.Builder
			for _, e := range listDevs {
				reason := e.Reason
				if reason == "" {
					reason = "Connectivity restored — ICMP reply received"
				}
				sb.WriteString(fmt.Sprintf("• <b>%s</b> (%s) — <code>%s</code> (%s)\n", e.DeviceName, e.DeviceType, e.IP, reason))
			}
			if len(evts) > limit {
				sb.WriteString(fmt.Sprintf("\n• <i>...and %d more devices (check the NOC dashboard)</i>\n", len(evts)-limit))
			}

			data := AggregatedEventData{
				Count:      len(evts),
				Location:   loc,
				Timestamp:  tStr,
				DeviceList: sb.String(),
			}
			indices.mu.Lock()
			idx := indices.aggUp % len(aggRecoveryTemplates)
			indices.aggUp++
			indices.mu.Unlock()
			msg = renderTemplate(aggRecoveryTemplates[idx], data)
		}

		var incidentIDs []string
		for _, e := range evts {
			if e.IncidentID != "" {
				incidentIDs = append(incidentIDs, e.IncidentID)
				
				// Write aggregation event
				var detail string
				if len(evts) == 1 {
					detail = "Sent recovery individually (only device UP in location)"
				} else {
					detail = fmt.Sprintf("Grouped recovery with %d other devices in %s — aggregation window closed", len(evts)-1, loc)
				}
				nq.pipeline.writeEvent(e.IncidentID, "notification_aggregated", "", detail)
			}
		}
		nq.pipeline.SendExt(context.Background(), msg, incidentIDs, true)
	}
}

func dedupEvents(evts []AggregatorEvent) []AggregatorEvent {
	seen := make(map[string]bool)
	var res []AggregatorEvent
	for _, e := range evts {
		key := e.DeviceID
		if key == "" {
			key = e.DeviceName
		}
		if !seen[key] {
			seen[key] = true
			res = append(res, e)
		}
	}
	return res
}

func (nq *NotifyQueue) Stop() {
	// No-op for compatibility
}

// ─── Template Variation Rotation ────────────────────────────────────────────

type SingleEventData struct {
	DeviceName string
	DeviceType string
	IP         string
	Location   string
	Reason     string
	Timestamp  string
}

type AggregatedEventData struct {
	Count      int
	Location   string
	Timestamp  string
	DeviceList string
}

type templateIndexTracker struct {
	mu         sync.Mutex
	singleDown int
	aggDown    int
	singleUp   int
	aggUp      int
}

var indices templateIndexTracker

var singleDownTemplates = []string{
	"🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>{{.DeviceName}}</b> ({{.DeviceType}})\nIP: <code>{{.IP}}</code>\nLocation: {{.Location}}\nReason: {{.Reason}}\nTime: {{.Timestamp}}",
	"⚠️ <b>ALERT — Device Offline</b>\nName: <b>{{.DeviceName}}</b> ({{.DeviceType}})\nIP: <code>{{.IP}}</code>\nLocation: {{.Location}}\nReason: {{.Reason}}\nDetected: {{.Timestamp}}",
	"🚨 <b>NETWORK ALERT — Unreachable</b>\nDevice <b>{{.DeviceName}}</b> is offline.\nIP: <code>{{.IP}}</code>\nZone: {{.Location}}\nReason: {{.Reason}}\nTime: {{.Timestamp}}",
}

var aggDownTemplates = []string{
	"🔴 <b>INCIDENT ALERT — {{.Count}} Devices DOWN</b>\nLocation: <b>{{.Location}}</b>\nTime: {{.Timestamp}}\n\nAffected Devices:\n{{.DeviceList}}",
	"⚠️ <b>MASS OUTAGE — {{.Count}} Nodes Offline</b>\nLocation: <b>{{.Location}}</b>\nReported: {{.Timestamp}}\n\nDisconnected units:\n{{.DeviceList}}",
	"🚨 <b>CRITICAL ALARM — {{.Count}} Devices Unreachable</b>\nZone: <b>{{.Location}}</b>\nDetected: {{.Timestamp}}\n\nUnreachable infrastructure:\n{{.DeviceList}}",
}

var singleRecoveryTemplates = []string{
	"🟢 <b>RECOVERED — Device UP</b>\nDevice: <b>{{.DeviceName}}</b> ({{.DeviceType}})\nIP: <code>{{.IP}}</code>\nLocation: {{.Location}}\nReason: {{.Reason}}\nTime: {{.Timestamp}}",
	"✅ <b>ONLINE — Device Restored</b>\nName: <b>{{.DeviceName}}</b> ({{.DeviceType}})\nIP: <code>{{.IP}}</code>\nLocation: {{.Location}}\nReason: {{.Reason}}\nRestored: {{.Timestamp}}",
	"✨ <b>RESTORED — Connection Re-established</b>\nDevice <b>{{.DeviceName}}</b> is back online.\nIP: <code>{{.IP}}</code>\nZone: {{.Location}}\nReason: {{.Reason}}\nTime: {{.Timestamp}}",
}

var aggRecoveryTemplates = []string{
	"🟢 <b>RECOVERY ALERT — {{.Count}} Devices UP</b>\nLocation: <b>{{.Location}}</b>\nTime: {{.Timestamp}}\n\nRecovered Devices:\n{{.DeviceList}}",
	"✅ <b>RESOLVED — {{.Count}} Nodes Back Online</b>\nLocation: <b>{{.Location}}</b>\nReported: {{.Timestamp}}\n\nActive infrastructure:\n{{.DeviceList}}",
	"✨ <b>SITUATION RESOLVED — {{.Count}} Devices Recovered</b>\nZone: <b>{{.Location}}</b>\nTimestamp: {{.Timestamp}}\n\nRecovered units:\n{{.DeviceList}}",
}

func renderTemplate(tmplStr string, data interface{}) string {
	t, err := template.New("msg").Parse(tmplStr)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}
	return buf.String()
}
