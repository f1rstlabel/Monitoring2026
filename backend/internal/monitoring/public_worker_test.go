package monitoring

import (
	"testing"
	"time"

	"sanoc/backend/internal/domain"
)

type lifecycleIncidentRepo struct {
	statuses []domain.PublicMonitorStatus
	incident *domain.PublicMonitorIncident
}

func (r *lifecycleIncidentRepo) SyncStatus(_ domain.PublicMonitor, status domain.PublicMonitorStatus, statusCode, latencyMs int, errorMessage string, checkedAt time.Time) (*domain.PublicMonitorIncident, bool, error) {
	r.statuses = append(r.statuses, status)
	changed := false
	if status == domain.PublicMonitorDown && r.incident == nil {
		r.incident = &domain.PublicMonitorIncident{ID: "pinc-test", MonitorID: "mon-test", Status: domain.PublicMonitorIncidentActive, InitialStatusCode: statusCode, FirstError: errorMessage, StartedAt: checkedAt.UTC().Format(time.RFC3339)}
		changed = true
	} else if status == domain.PublicMonitorUp && r.incident != nil && r.incident.Status == domain.PublicMonitorIncidentActive {
		r.incident.Status = domain.PublicMonitorIncidentResolved
		r.incident.FinalStatusCode = statusCode
		r.incident.ResolvedAt = checkedAt.UTC().Format(time.RFC3339)
		r.incident.DurationSeconds = int64(checkedAt.Sub(parseTestTime(r.incident.StartedAt)).Seconds())
		r.incident.LastError = ""
		changed = true
	}
	if r.incident != nil {
		r.incident.LastError = errorMessage
		r.incident.MonitorName = "Public API"
		r.incident.TargetURL = "https://example.com/health"
		_ = latencyMs
	}
	return r.incident, changed, nil
}

func (r *lifecycleIncidentRepo) ResolveForPause(string, time.Time) error    { return nil }
func (r *lifecycleIncidentRepo) ResolveForDeletion(string, time.Time) error { return nil }
func (r *lifecycleIncidentRepo) GetAll(string, string, string, time.Time, time.Time, int, int) ([]domain.PublicMonitorIncident, int, error) {
	return nil, 0, nil
}
func (r *lifecycleIncidentRepo) GetByID(string) (*domain.PublicMonitorIncident, error) {
	return r.incident, nil
}
func (r *lifecycleIncidentRepo) GetEvents(string, int) ([]domain.PublicMonitorIncidentEvent, error) {
	return nil, nil
}
func (r *lifecycleIncidentRepo) AppendEvent(*domain.PublicMonitorIncidentEvent) error { return nil }
func (r *lifecycleIncidentRepo) GetNotificationLogs(string, int) ([]domain.PublicMonitorNotificationLog, error) {
	return nil, nil
}
func (r *lifecycleIncidentRepo) AppendNotificationLog(*domain.PublicMonitorNotificationLog) error {
	return nil
}
func (r *lifecycleIncidentRepo) GetReport(string, time.Time, time.Time) (*domain.PublicMonitorReport, error) {
	return nil, nil
}

func parseTestTime(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339, value)
	return parsed
}

func TestPublicIncidentLifecycle(t *testing.T) {
	repo := &lifecycleIncidentRepo{}
	worker := NewWorker(nil, nil)
	worker.SetIncidentRepo(repo)
	monitor := domain.PublicMonitor{ID: "mon-test", Name: "Public API", TargetURL: "https://example.com/health"}
	active, changed, err := worker.syncIncident(monitor, CheckResult{Status: domain.PublicMonitorDown, StatusCode: 503, LatencyMs: 100, Error: "HTTP status 503"}, time.Now().UTC())
	if err != nil || !changed || active == nil || active.Status != domain.PublicMonitorIncidentActive {
		t.Fatalf("expected DOWN to open active public incident, incident=%+v changed=%v err=%v", active, changed, err)
	}

	resolved, changed, err := worker.syncIncident(monitor, CheckResult{Status: domain.PublicMonitorUp, StatusCode: 200, LatencyMs: 80}, time.Now().UTC())
	if err != nil || !changed || resolved == nil || resolved.Status != domain.PublicMonitorIncidentResolved {
		t.Fatalf("expected UP to resolve public incident, incident=%+v changed=%v err=%v", resolved, changed, err)
	}

	if len(repo.statuses) != 2 || repo.statuses[0] != domain.PublicMonitorDown || repo.statuses[1] != domain.PublicMonitorUp {
		t.Fatalf("expected lifecycle statuses [DOWN UP], got %v", repo.statuses)
	}
}
