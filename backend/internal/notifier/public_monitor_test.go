package notifier

import (
	"context"
	"testing"
	"time"

	"sanoc/backend/internal/domain"
)

type publicNotificationRepo struct {
	events []domain.PublicMonitorIncidentEvent
	logs   []domain.PublicMonitorNotificationLog
}

func (r *publicNotificationRepo) SyncStatus(domain.PublicMonitor, domain.PublicMonitorStatus, int, int, string, time.Time) (*domain.PublicMonitorIncident, bool, error) {
	return nil, false, nil
}
func (r *publicNotificationRepo) ResolveForPause(string, time.Time) error    { return nil }
func (r *publicNotificationRepo) ResolveForDeletion(string, time.Time) error { return nil }
func (r *publicNotificationRepo) GetAll(string, string, string, time.Time, time.Time, int, int) ([]domain.PublicMonitorIncident, int, error) {
	return nil, 0, nil
}
func (r *publicNotificationRepo) GetByID(string) (*domain.PublicMonitorIncident, error) {
	return nil, nil
}
func (r *publicNotificationRepo) GetEvents(string, int) ([]domain.PublicMonitorIncidentEvent, error) {
	return r.events, nil
}
func (r *publicNotificationRepo) AppendEvent(event *domain.PublicMonitorIncidentEvent) error {
	r.events = append(r.events, *event)
	return nil
}
func (r *publicNotificationRepo) GetNotificationLogs(string, int) ([]domain.PublicMonitorNotificationLog, error) {
	return r.logs, nil
}
func (r *publicNotificationRepo) AppendNotificationLog(log *domain.PublicMonitorNotificationLog) error {
	r.logs = append(r.logs, *log)
	return nil
}
func (r *publicNotificationRepo) GetReport(string, time.Time, time.Time) (*domain.PublicMonitorReport, error) {
	return nil, nil
}

func TestPublicNotificationFallbackIsRecorded(t *testing.T) {
	repo := &publicNotificationRepo{}
	pipeline := NewPipeline("", "", "", "", "", 60, nil, nil, nil)
	pipeline.SetPublicMonitorIncidentRepo(repo)

	err := pipeline.dispatchActualWithErr(context.Background(), "Public API is DOWN", nil, []string{"pinc-test"})
	if err == nil {
		t.Fatal("expected both unconfigured WhatsApp and Telegram delivery to fail")
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for len(repo.logs) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	seen := map[string]bool{}
	for _, event := range repo.events {
		seen[event.EventType+":"+event.Channel] = true
	}
	for _, expected := range []string{"channel_attempt:whatsapp", "channel_failed:whatsapp", "channel_fallback:telegram", "channel_attempt:telegram", "channel_failed:telegram"} {
		if !seen[expected] {
			t.Errorf("expected public incident event %q, events=%+v", expected, repo.events)
		}
	}
	if len(repo.logs) != 1 || repo.logs[0].IncidentID != "pinc-test" || repo.logs[0].Channel != "Telegram" || repo.logs[0].Status != "failed" {
		t.Fatalf("expected failed Telegram fallback log for public incident, logs=%+v", repo.logs)
	}
}
