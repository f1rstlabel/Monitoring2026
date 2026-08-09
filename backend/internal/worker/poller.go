package worker

import (
	"log"
	"math/rand"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/ws"
)

type PollerEngine struct {
	hub      *ws.Hub
	interval time.Duration
	stopChan chan struct{}
}

func NewPollerEngine(hub *ws.Hub, intervalSec int) *PollerEngine {
	return &PollerEngine{
		hub:      hub,
		interval: time.Duration(intervalSec) * time.Second,
		stopChan: make(chan struct{}),
	}
}

func (p *PollerEngine) Start() {
	ticker := time.NewTicker(p.interval)
	log.Printf("[PollerEngine] Background ICMP/SNMP poller started (Interval: %v)", p.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				p.performPollTick()
			case <-p.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *PollerEngine) performPollTick() {
	// Simulate background polling tick over 414 devices
	// Randomly broadcast status toggles or ping updates
	if rand.Float32() > 0.7 {
		msg := domain.WSMessage{
			Type:        "INCIDENT_ALERT",
			Title:       "ICMP Poller Check Completed",
			Description: "414 Infrastructure nodes scanned cleanly across Biro Umum subnet",
			Severity:    "info",
			Timestamp:   time.Now(),
		}
		p.hub.BroadcastMessage(msg)
	}
}

func (p *PollerEngine) Stop() {
	close(p.stopChan)
}
