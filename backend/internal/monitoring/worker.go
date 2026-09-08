package monitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/notifier"
	"sanoc/backend/internal/repository"
	"sanoc/backend/internal/ws"
)

type Worker struct {
	repo         repository.PublicMonitorRepository
	incidentRepo repository.PublicMonitorIncidentRepository
	hub          *ws.Hub
	pipeline     *notifier.Pipeline
	stop         chan struct{}
	wg           sync.WaitGroup
	running      map[string]bool
	mu           sync.Mutex
}

func NewWorker(repo repository.PublicMonitorRepository, hub *ws.Hub) *Worker {
	return &Worker{
		repo:    repo,
		hub:     hub,
		stop:    make(chan struct{}),
		running: make(map[string]bool),
	}
}

func (w *Worker) SetPipeline(pipeline *notifier.Pipeline) {
	w.pipeline = pipeline
}

func (w *Worker) SetIncidentRepo(repo repository.PublicMonitorIncidentRepository) {
	if w != nil {
		w.incidentRepo = repo
	}
}

func (w *Worker) Start() {
	if w == nil || w.repo == nil {
		return
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.runOnce()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.runOnce()
			case <-w.stop:
				return
			}
		}
	}()
	log.Println("[Public Monitoring] Worker started")
}

func (w *Worker) Stop() {
	if w == nil {
		return
	}
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	w.wg.Wait()
}

func (w *Worker) runOnce() {
	monitors, err := w.repo.GetEnabled()
	if err != nil {
		log.Printf("[Public Monitoring] Failed to load monitors: %v", err)
		return
	}
	for _, monitor := range monitors {
		if !isDue(monitor) || !w.claim(monitor.ID) {
			continue
		}
		monitor := monitor
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			defer w.release(monitor.ID)
			w.check(context.Background(), monitor)
		}()
	}
}

func isDue(monitor domain.PublicMonitor) bool {
	if monitor.LastChecked == "" {
		return true
	}
	last, err := time.Parse(time.RFC3339, monitor.LastChecked)
	if err != nil {
		return true
	}
	interval := time.Duration(monitor.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return time.Since(last) >= interval
}

func (w *Worker) claim(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running[id] {
		return false
	}
	w.running[id] = true
	return true
}

func (w *Worker) release(id string) {
	w.mu.Lock()
	delete(w.running, id)
	w.mu.Unlock()
}

func (w *Worker) check(ctx context.Context, monitor domain.PublicMonitor) {
	result := RunPublicMonitorCheck(ctx, monitor)
	checkedAt := time.Now().UTC()
	previous, changed, err := w.repo.RecordCheck(monitor.ID, result.Status, result.StatusCode, result.LatencyMs, result.Error, checkedAt)
	if err != nil {
		log.Printf("[Public Monitoring] Failed to record %s: %v", monitor.Name, err)
		return
	}
	incident, incidentChanged, err := w.syncIncident(monitor, result, checkedAt)
	if err != nil {
		log.Printf("[Public Monitoring] Failed to sync incident for %s: %v", monitor.Name, err)
	}
	w.broadcast(monitor, result, checkedAt, changed, incidentChanged)
	if changed || incidentChanged {
		if monitor.NotifyOnFailure && w.pipeline != nil && incident != nil && incidentChanged {
			message := fmt.Sprintf("SANOC Public Monitoring: %s is DOWN. Target: %s. Reason: %s", monitor.Name, monitor.TargetURL, result.Error)
			if result.Status == domain.PublicMonitorUp {
				message = fmt.Sprintf("SANOC Public Monitoring: %s recovered and is UP again. Target: %s. Response: %dms", monitor.Name, monitor.TargetURL, result.LatencyMs)
			}
			w.pipeline.SendPublic(context.Background(), message, incident.ID)
		}
		log.Printf("[Public Monitoring] %s changed %s -> %s", monitor.Name, previous, result.Status)
	}
}

func (w *Worker) syncIncident(monitor domain.PublicMonitor, result CheckResult, checkedAt time.Time) (*domain.PublicMonitorIncident, bool, error) {
	if w.incidentRepo == nil {
		return nil, false, nil
	}
	return w.incidentRepo.SyncStatus(monitor, result.Status, result.StatusCode, result.LatencyMs, result.Error, checkedAt)
}

func (w *Worker) broadcast(monitor domain.PublicMonitor, result CheckResult, checkedAt time.Time, statusChanged, incidentChanged bool) {
	if w.hub == nil {
		return
	}
	severity := "info"
	if result.Status == domain.PublicMonitorDown {
		severity = "critical"
	}
	w.hub.Broadcast(map[string]interface{}{
		"type":            "PUBLIC_MONITOR_UPDATE",
		"monitorId":       monitor.ID,
		"status":          result.Status,
		"title":           monitor.Name,
		"latencyMs":       result.LatencyMs,
		"statusCode":      result.StatusCode,
		"error":           result.Error,
		"checkedAt":       checkedAt.Format(time.RFC3339),
		"statusChanged":   statusChanged,
		"incidentChanged": incidentChanged,
		"severity":        severity,
	})
}

// CheckNow is used by the API's manual-check action.
func (w *Worker) CheckNow(ctx context.Context, monitor domain.PublicMonitor) (CheckResult, error) {
	result := RunPublicMonitorCheck(ctx, monitor)
	checkedAt := time.Now().UTC()
	_, statusChanged, err := w.repo.RecordCheck(monitor.ID, result.Status, result.StatusCode, result.LatencyMs, result.Error, checkedAt)
	if err != nil {
		return result, err
	}
	incident, incidentChanged, err := w.syncIncident(monitor, result, checkedAt)
	if err != nil {
		return result, err
	}
	if monitor.NotifyOnFailure && w.pipeline != nil && incident != nil && incidentChanged {
		message := fmt.Sprintf("SANOC Public Monitoring: %s is DOWN. Target: %s. Reason: %s", monitor.Name, monitor.TargetURL, result.Error)
		if result.Status == domain.PublicMonitorUp {
			message = fmt.Sprintf("SANOC Public Monitoring: %s recovered and is UP again. Target: %s. Response: %dms", monitor.Name, monitor.TargetURL, result.LatencyMs)
		}
		w.pipeline.SendPublic(context.Background(), message, incident.ID)
	}
	w.broadcast(monitor, result, checkedAt, statusChanged, incidentChanged)
	return result, nil
}
