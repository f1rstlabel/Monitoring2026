package notifier

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"govmonitor-it/backend/internal/domain"
)

type EventType string

const (
	EventDown      EventType = "DOWN"
	EventRecovered EventType = "RECOVERED"
)

type IncidentEvent struct {
	DeviceID   string
	DeviceName string
	DeviceType string
	IP         string
	Location   string
	EventType  EventType
	Timestamp  time.Time
}

type NotifyQueue struct {
	pipeline *Pipeline
	repo     domainNotificationLogRepo

	mu           sync.Mutex
	pending      []IncidentEvent
	aggTimer     *time.Timer
	aggWindow    time.Duration
	waMinInterval time.Duration

	lastWASend time.Time
	queueChan  chan string
	stopChan   chan struct{}
}

type domainNotificationLogRepo interface {
	Append(row *domain.NotificationLogRow) error
}

func NewNotifyQueue(pipeline *Pipeline, repo domainNotificationLogRepo) *NotifyQueue {
	nq := &NotifyQueue{
		pipeline:      pipeline,
		repo:          repo,
		aggWindow:     5 * time.Second,
		waMinInterval: 5 * time.Second,
		queueChan:     make(chan string, 100),
		stopChan:      make(chan struct{}),
	}
	go nq.workerLoop()
	return nq
}

func (nq *NotifyQueue) Enqueue(evt IncidentEvent) {
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
	downEvents := make(map[string][]IncidentEvent)
	upEvents := make(map[string][]IncidentEvent)

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
			msg = fmt.Sprintf("🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>%s</b> (%s)\nIP: <code>%s</code>\nLocation: %s\nTime: %s", e.DeviceName, e.DeviceType, e.IP, loc, tStr)
		} else {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("🔴 <b>INCIDENT ALERT — %d Devices DOWN</b>\nLocation: <b>%s</b>\nTime: %s\n\nAffected Devices:\n", len(evts), loc, tStr))
			for _, e := range evts {
				sb.WriteString(fmt.Sprintf("• <b>%s</b> (%s) — <code>%s</code>\n", e.DeviceName, e.DeviceType, e.IP))
			}
			msg = sb.String()
		}
		nq.queueChan <- msg
	}

	for loc, rawEvts := range upEvents {
		evts := dedupEvents(rawEvts)
		var msg string
		tStr := evts[0].Timestamp.Format("15:04:05 WIB")
		if len(evts) == 1 {
			e := evts[0]
			msg = fmt.Sprintf("🟢 <b>RECOVERED — Device UP</b>\nDevice: <b>%s</b> (%s)\nIP: <code>%s</code>\nLocation: %s\nTime: %s", e.DeviceName, e.DeviceType, e.IP, loc, tStr)
		} else {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("🟢 <b>RECOVERY ALERT — %d Devices UP</b>\nLocation: <b>%s</b>\nTime: %s\n\nRecovered Devices:\n", len(evts), loc, tStr))
			for _, e := range evts {
				sb.WriteString(fmt.Sprintf("• <b>%s</b> (%s) — <code>%s</code>\n", e.DeviceName, e.DeviceType, e.IP))
			}
			msg = sb.String()
		}
		nq.queueChan <- msg
	}
}

func dedupEvents(evts []IncidentEvent) []IncidentEvent {
	seen := make(map[string]bool)
	var res []IncidentEvent
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

func (nq *NotifyQueue) workerLoop() {
	for {
		select {
		case msg := <-nq.queueChan:
			go nq.dispatchWithRateLimit(msg)
		case <-nq.stopChan:
			return
		}
	}
}

func (nq *NotifyQueue) dispatchWithRateLimit(msg string) {
	nq.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(nq.lastWASend)

	if nq.lastWASend.IsZero() || elapsed >= nq.waMinInterval {
		nq.lastWASend = now
		nq.mu.Unlock()
		log.Printf("[NotifyQueue] Message ready, rate limit OK, sending now")
		if nq.pipeline != nil {
			nq.pipeline.Send(context.Background(), msg)
		}
		return
	}

	waitTime := nq.waMinInterval - elapsed
	nq.lastWASend = now.Add(waitTime)
	nq.mu.Unlock()

	log.Printf("[NotifyQueue] Message ready, rate limited, sending in %s", waitTime.Round(time.Second))
	time.Sleep(waitTime)

	log.Printf("[NotifyQueue] Rate limit delay elapsed, sending now")
	if nq.pipeline != nil {
		nq.pipeline.Send(context.Background(), msg)
	}
}

func (nq *NotifyQueue) Stop() {
	close(nq.stopChan)
}
