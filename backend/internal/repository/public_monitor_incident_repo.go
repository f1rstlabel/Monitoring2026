package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

type PostgresPublicMonitorIncidentRepository struct {
	db *sql.DB
}

func NewPostgresPublicMonitorIncidentRepository(db *sql.DB) *PostgresPublicMonitorIncidentRepository {
	return &PostgresPublicMonitorIncidentRepository{db: db}
}

func (r *PostgresPublicMonitorIncidentRepository) SyncStatus(monitor domain.PublicMonitor, status domain.PublicMonitorStatus, statusCode, latencyMs int, errorMessage string, checkedAt time.Time) (*domain.PublicMonitorIncident, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, fmt.Errorf("public monitor incident database is not initialized")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	active, err := scanPublicMonitorIncident(tx.QueryRow(`
		SELECT i.id, i.monitor_id, COALESCE(m.name, ''), COALESCE(m.monitor_type, 'http'), COALESCE(m.target_url, ''), i.status,
		       i.started_at, i.resolved_at, i.duration_seconds, i.initial_status_code,
		       i.final_status_code, COALESCE(i.first_error, ''), COALESCE(i.last_error, ''), COALESCE(i.resolution_reason, '')
		FROM public_monitor_incidents i
		JOIN public_monitors m ON m.id = i.monitor_id
		WHERE i.monitor_id=$1 AND i.status='ACTIVE'
		FOR UPDATE`, monitor.ID))
	if err != nil && err != sql.ErrNoRows {
		return nil, false, err
	}

	if status == domain.PublicMonitorDown {
		if active != nil {
			_, err = tx.Exec(`UPDATE public_monitor_incidents
				SET final_status_code=$1, last_error=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3`, statusCode, errorMessage, active.ID)
			if err != nil {
				return nil, false, err
			}
			active.FinalStatusCode = statusCode
			active.LastError = errorMessage
			return commitPublicIncident(tx, active, false)
		}

		incident := &domain.PublicMonitorIncident{
			ID:                publicMonitorIncidentID("pinc"),
			MonitorID:         monitor.ID,
			MonitorName:       monitor.Name,
			MonitorType:       monitor.MonitorType,
			TargetURL:         monitor.TargetURL,
			Status:            domain.PublicMonitorIncidentActive,
			StartedAt:         checkedAt.UTC().Format(time.RFC3339),
			InitialStatusCode: statusCode,
			FinalStatusCode:   statusCode,
			FirstError:        errorMessage,
			LastError:         errorMessage,
		}
		if _, err = tx.Exec(`INSERT INTO public_monitor_incidents
			(id, monitor_id, status, started_at, initial_status_code, final_status_code, first_error, last_error)
			VALUES ($1,$2,'ACTIVE',$3,$4,$5,$6,$7)`, incident.ID, incident.MonitorID, checkedAt, statusCode, statusCode, errorMessage, errorMessage); err != nil {
			return nil, false, err
		}
		if err = insertPublicIncidentEvent(tx, incident.ID, "incident_opened", fmt.Sprintf("Endpoint is DOWN: %s", errorMessage), checkedAt); err != nil {
			return nil, false, err
		}
		return commitPublicIncident(tx, incident, true)
	}

	if status == domain.PublicMonitorUp && active != nil {
		resolvedAt := checkedAt.UTC()
		startedAt, parseErr := time.Parse(time.RFC3339, active.StartedAt)
		if parseErr != nil {
			startedAt = resolvedAt
		}
		duration := int64(resolvedAt.Sub(startedAt).Seconds())
		if duration < 0 {
			duration = 0
		}
		_, err = tx.Exec(`UPDATE public_monitor_incidents
			SET status='RESOLVED', resolved_at=$1, duration_seconds=$2, final_status_code=$3,
			    last_error='', resolution_reason='MONITOR_RECOVERED', updated_at=CURRENT_TIMESTAMP WHERE id=$4`, resolvedAt, duration, statusCode, active.ID)
		if err != nil {
			return nil, false, err
		}
		active.Status = domain.PublicMonitorIncidentResolved
		active.ResolvedAt = resolvedAt.Format(time.RFC3339)
		active.DurationSeconds = duration
		active.FinalStatusCode = statusCode
		active.LastError = ""
		active.ResolutionReason = "MONITOR_RECOVERED"
		if err = insertPublicIncidentEvent(tx, active.ID, "incident_resolved", "Endpoint recovered and is UP again", checkedAt); err != nil {
			return nil, false, err
		}
		return commitPublicIncident(tx, active, true)
	}

	return commitPublicIncident(tx, nil, false)
}

func commitPublicIncident(tx *sql.Tx, incident *domain.PublicMonitorIncident, changed bool) (*domain.PublicMonitorIncident, bool, error) {
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	return incident, changed, nil
}

func insertPublicIncidentEvent(tx *sql.Tx, incidentID, eventType, detail string, occurredAt time.Time) error {
	_, err := tx.Exec(`INSERT INTO public_monitor_incident_events (id, incident_id, event_type, detail, occurred_at)
		VALUES ($1,$2,$3,$4,$5)`, publicMonitorIncidentID("pevt"), incidentID, eventType, detail, occurredAt)
	return err
}

func (r *PostgresPublicMonitorIncidentRepository) ResolveForPause(monitorID string, resolvedAt time.Time) error {
	return r.resolveActiveIncidents(monitorID, resolvedAt, "MONITOR_PAUSED", "incident_paused", "Monitor paused; downtime clock stopped")
}

func (r *PostgresPublicMonitorIncidentRepository) ResolveForDeletion(monitorID string, resolvedAt time.Time) error {
	return r.resolveActiveIncidents(monitorID, resolvedAt, "MONITOR_DELETED", "incident_deleted", "Monitor deleted; incident closed for audit")
}

func (r *PostgresPublicMonitorIncidentRepository) resolveActiveIncidents(monitorID string, resolvedAt time.Time, reason, eventType, detail string) error {
	if r == nil || r.db == nil || monitorID == "" {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT id, started_at FROM public_monitor_incidents WHERE monitor_id=$1 AND status='ACTIVE' FOR UPDATE`, monitorID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var startedAt time.Time
		if err := rows.Scan(&id, &startedAt); err != nil {
			return err
		}
		duration := int64(resolvedAt.Sub(startedAt).Seconds())
		if duration < 0 {
			duration = 0
		}
		if _, err := tx.Exec(`UPDATE public_monitor_incidents SET status='RESOLVED', resolved_at=$1,
			duration_seconds=$2, resolution_reason=$3, last_error=$4, updated_at=CURRENT_TIMESTAMP WHERE id=$5`, resolvedAt, duration, reason, detail, id); err != nil {
			return err
		}
		if err := insertPublicIncidentEvent(tx, id, eventType, detail, resolvedAt); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()
}

const publicMonitorIncidentColumns = `i.id, i.monitor_id, COALESCE(m.name, ''), COALESCE(m.monitor_type, 'http'), COALESCE(m.target_url, ''), i.status,
	i.started_at, i.resolved_at, CASE WHEN i.status='ACTIVE' THEN GREATEST(0, EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - i.started_at))::BIGINT) ELSE i.duration_seconds END,
	i.initial_status_code, i.final_status_code, COALESCE(i.first_error, ''), COALESCE(i.last_error, ''), COALESCE(i.resolution_reason, '')`

func (r *PostgresPublicMonitorIncidentRepository) GetAll(monitorID, status, search string, from, to time.Time, page, limit int) ([]domain.PublicMonitorIncident, int, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitorIncident{}, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	where := " WHERE 1=1"
	args := make([]interface{}, 0, 8)
	idx := 1
	if monitorID != "" {
		where += fmt.Sprintf(" AND i.monitor_id=$%d", idx)
		args = append(args, monitorID)
		idx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND i.status=$%d", idx)
		args = append(args, strings.ToUpper(strings.TrimSpace(status)))
		idx++
	}
	if search != "" {
		where += fmt.Sprintf(" AND (LOWER(m.name) LIKE $%d OR LOWER(m.target_url) LIKE $%d OR LOWER(i.id) LIKE $%d)", idx, idx+1, idx+2)
		term := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		args = append(args, term, term, term)
		idx += 3
	}
	if !from.IsZero() {
		where += fmt.Sprintf(" AND i.started_at <= $%d", idx)
		args = append(args, to)
		idx++
		where += fmt.Sprintf(" AND COALESCE(i.resolved_at, CURRENT_TIMESTAMP) >= $%d", idx)
		args = append(args, from)
		idx++
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM public_monitor_incidents i LEFT JOIN public_monitors m ON m.id=i.monitor_id"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := "SELECT " + publicMonitorIncidentColumns + " FROM public_monitor_incidents i LEFT JOIN public_monitors m ON m.id=i.monitor_id" + where + fmt.Sprintf(" ORDER BY CASE i.status WHEN 'ACTIVE' THEN 0 ELSE 1 END, i.started_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, (page-1)*limit)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitorIncident, 0)
	for rows.Next() {
		item, err := scanPublicMonitorIncident(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *PostgresPublicMonitorIncidentRepository) GetByID(id string) (*domain.PublicMonitorIncident, error) {
	if r == nil || r.db == nil {
		return nil, sql.ErrNoRows
	}
	return scanPublicMonitorIncident(r.db.QueryRow("SELECT "+publicMonitorIncidentColumns+" FROM public_monitor_incidents i LEFT JOIN public_monitors m ON m.id=i.monitor_id WHERE i.id=$1", id))
}

func (r *PostgresPublicMonitorIncidentRepository) GetEvents(incidentID string, limit int) ([]domain.PublicMonitorIncidentEvent, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, incident_id, event_type, channel, detail, occurred_at
		FROM public_monitor_incident_events WHERE incident_id=$1 ORDER BY occurred_at DESC LIMIT $2`, incidentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitorIncidentEvent, 0)
	for rows.Next() {
		var item domain.PublicMonitorIncidentEvent
		var occurredAt time.Time
		if err := rows.Scan(&item.ID, &item.IncidentID, &item.EventType, &item.Channel, &item.Detail, &occurredAt); err != nil {
			return nil, err
		}
		item.OccurredAt = occurredAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresPublicMonitorIncidentRepository) AppendEvent(event *domain.PublicMonitorIncidentEvent) error {
	if r == nil || r.db == nil || event == nil || event.IncidentID == "" {
		return nil
	}
	if event.ID == "" {
		event.ID = publicMonitorIncidentID("pevt")
	}
	if strings.TrimSpace(event.OccurredAt) == "" {
		event.OccurredAt = time.Now().UTC().Format(time.RFC3339)
	}
	occurredAt, err := time.Parse(time.RFC3339, event.OccurredAt)
	if err != nil {
		occurredAt = time.Now().UTC()
	}
	_, err = r.db.Exec(`INSERT INTO public_monitor_incident_events
		(id, incident_id, event_type, channel, detail, occurred_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		event.ID, event.IncidentID, event.EventType, event.Channel, event.Detail, occurredAt)
	return err
}

func (r *PostgresPublicMonitorIncidentRepository) GetNotificationLogs(incidentID string, limit int) ([]domain.PublicMonitorNotificationLog, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, incident_id, channel, recipient, status, COALESCE(error_msg, ''), sent_at
		FROM public_monitor_notification_logs WHERE incident_id=$1 ORDER BY sent_at DESC LIMIT $2`, incidentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitorNotificationLog, 0)
	for rows.Next() {
		var item domain.PublicMonitorNotificationLog
		var sentAt time.Time
		if err := rows.Scan(&item.ID, &item.IncidentID, &item.Channel, &item.Recipient, &item.Status, &item.Error, &sentAt); err != nil {
			return nil, err
		}
		item.SentAt = sentAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresPublicMonitorIncidentRepository) AppendNotificationLog(row *domain.PublicMonitorNotificationLog) error {
	if r == nil || r.db == nil || row == nil || row.IncidentID == "" {
		return nil
	}
	if row.ID == "" {
		row.ID = publicMonitorIncidentID("pnot")
	}
	_, err := r.db.Exec(`INSERT INTO public_monitor_notification_logs
		(id, incident_id, channel, recipient, status, error_msg, sent_at) VALUES ($1,$2,$3,$4,$5,$6,CURRENT_TIMESTAMP)`,
		row.ID, row.IncidentID, row.Channel, row.Recipient, row.Status, row.Error)
	return err
}

func (r *PostgresPublicMonitorIncidentRepository) GetReport(monitorID string, from, to time.Time) (*domain.PublicMonitorReport, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("public monitor report database is not initialized")
	}
	if from.IsZero() {
		from = time.Now().UTC().Add(-30 * 24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if !to.After(from) {
		return nil, fmt.Errorf("report end must be after report start")
	}
	report := &domain.PublicMonitorReport{MonitorID: monitorID, From: from.UTC().Format(time.RFC3339), To: to.UTC().Format(time.RFC3339), Points: make([]domain.PublicMonitorReportPoint, 0), MonitorSummaries: make([]domain.PublicMonitorSummary, 0)}
	summaryByMonitor := make(map[string]int)
	type summaryLatencyAggregate struct {
		sum   float64
		count int
	}
	summaryLatency := make(map[string]summaryLatencyAggregate)
	if monitorID == "" {
		monitorRows, err := r.db.Query(`SELECT id, name, monitor_type, COALESCE(target_url, ''), COALESCE(group_name, 'Ungrouped'), status, enabled,
			COALESCE(last_checked, TIMESTAMPTZ 'epoch'), COALESCE(last_latency_ms, 0), COALESCE(last_error, '')
			FROM public_monitors WHERE deleted_at IS NULL AND purged_at IS NULL ORDER BY name ASC`)
		if err != nil {
			return nil, err
		}
		defer monitorRows.Close()
		for monitorRows.Next() {
			var summary domain.PublicMonitorSummary
			var monitorType, status string
			var lastChecked sql.NullTime
			if err := monitorRows.Scan(&summary.MonitorID, &summary.MonitorName, &monitorType, &summary.TargetURL, &summary.GroupName, &status, &summary.Enabled, &lastChecked, &summary.LastLatencyMs, &summary.LastError); err != nil {
				return nil, err
			}
			summary.MonitorType = domain.PublicMonitorType(monitorType)
			summary.Status = domain.PublicMonitorStatus(status)
			if lastChecked.Valid && !lastChecked.Time.Equal(time.Unix(0, 0)) {
				summary.LastChecked = lastChecked.Time.UTC().Format(time.RFC3339)
			}
			report.MonitorSummaries = append(report.MonitorSummaries, summary)
			summaryByMonitor[summary.MonitorID] = len(report.MonitorSummaries) - 1
		}
		if err := monitorRows.Err(); err != nil {
			return nil, err
		}
	}
	upLatencyCount := 0
	if monitorID != "" {
		_ = r.db.QueryRow("SELECT COALESCE(name, '') FROM public_monitors WHERE id=$1", monitorID).Scan(&report.MonitorName)
	}
	where := " WHERE m.deleted_at IS NULL AND m.purged_at IS NULL AND c.checked_at >= $1 AND c.checked_at <= $2"
	args := []interface{}{from, to}
	if monitorID != "" {
		where += " AND c.monitor_id=$3"
		args = append(args, monitorID)
	}
	rows, err := r.db.Query("SELECT c.monitor_id, c.checked_at, c.status, c.latency_ms FROM public_monitor_checks c JOIN public_monitors m ON m.id = c.monitor_id"+where+" ORDER BY c.checked_at ASC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type aggregate struct {
		up, down, latencyCount int
		latencySum             float64
	}
	bucketSize := time.Hour
	if to.Sub(from) > 48*time.Hour {
		bucketSize = 24 * time.Hour
	}
	points := make(map[time.Time]*aggregate)
	for rows.Next() {
		var checkedMonitorID string
		var checkedAt time.Time
		var status string
		var latency int
		if err := rows.Scan(&checkedMonitorID, &checkedAt, &status, &latency); err != nil {
			return nil, err
		}
		if summaryIndex, ok := summaryByMonitor[checkedMonitorID]; ok {
			summary := &report.MonitorSummaries[summaryIndex]
			summary.TotalChecks++
			if status == string(domain.PublicMonitorUp) {
				summary.UpChecks++
				if latency > 0 {
					latencyAggregate := summaryLatency[checkedMonitorID]
					latencyAggregate.sum += float64(latency)
					latencyAggregate.count++
					summaryLatency[checkedMonitorID] = latencyAggregate
				}
			} else {
				summary.DownChecks++
			}
		}
		report.TotalChecks++
		bucket := checkedAt.UTC().Truncate(bucketSize)
		if points[bucket] == nil {
			points[bucket] = &aggregate{}
		}
		if status == string(domain.PublicMonitorUp) {
			report.UpChecks++
			points[bucket].up++
			if latency > 0 {
				points[bucket].latencyCount++
				points[bucket].latencySum += float64(latency)
				report.AvgLatencyMs += float64(latency)
				upLatencyCount++
			}
		} else {
			report.DownChecks++
			points[bucket].down++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if report.TotalChecks > 0 {
		report.UptimePercent = round2(float64(report.UpChecks) / float64(report.TotalChecks) * 100)
	} else {
		report.UptimePercent = 100
	}
	if report.AvgLatencyMs > 0 {
		if upLatencyCount > 0 {
			report.AvgLatencyMs = round2(report.AvgLatencyMs / float64(upLatencyCount))
		}
	}
	incidentWhere := " WHERE started_at <= $1 AND COALESCE(resolved_at, CURRENT_TIMESTAMP) >= $2"
	incidentArgs := []interface{}{to, from}
	if monitorID != "" {
		incidentWhere += " AND monitor_id=$3"
		incidentArgs = append(incidentArgs, monitorID)
	}
	incidentRows, err := r.db.Query("SELECT monitor_id, started_at, resolved_at FROM public_monitor_incidents"+incidentWhere, incidentArgs...)
	if err != nil {
		return nil, err
	}
	for incidentRows.Next() {
		var incidentMonitorID string
		var started time.Time
		var resolved sql.NullTime
		if err := incidentRows.Scan(&incidentMonitorID, &started, &resolved); err != nil {
			incidentRows.Close()
			return nil, err
		}
		report.IncidentCount++
		if summaryIndex, ok := summaryByMonitor[incidentMonitorID]; ok {
			report.MonitorSummaries[summaryIndex].IncidentCount++
		}
		end := to
		if resolved.Valid && resolved.Time.Before(end) {
			end = resolved.Time
		}
		start := started
		if start.Before(from) {
			start = from
		}
		if end.After(start) {
			report.DowntimeMinutes += int(end.Sub(start).Minutes())
		}
	}
	if err := incidentRows.Err(); err != nil {
		incidentRows.Close()
		return nil, err
	}
	incidentRows.Close()
	keys := make([]time.Time, 0, len(points))
	for bucket := range points {
		keys = append(keys, bucket)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	for _, bucket := range keys {
		agg := points[bucket]
		total := agg.up + agg.down
		point := domain.PublicMonitorReportPoint{Bucket: bucket.UTC().Format(time.RFC3339), UpCount: agg.up, DownCount: agg.down}
		if total > 0 {
			point.UptimePercent = round2(float64(agg.up) / float64(total) * 100)
		} else {
			point.UptimePercent = 100
		}
		if agg.latencyCount > 0 {
			point.AvgLatencyMs = round2(agg.latencySum / float64(agg.latencyCount))
		}
		report.Points = append(report.Points, point)
	}
	for i := range report.MonitorSummaries {
		summary := &report.MonitorSummaries[i]
		if summary.TotalChecks > 0 {
			summary.UptimePercent = round2(float64(summary.UpChecks) / float64(summary.TotalChecks) * 100)
		} else {
			summary.UptimePercent = 100
		}
		if latencyAggregate := summaryLatency[summary.MonitorID]; latencyAggregate.count > 0 {
			summary.AvgLatencyMs = round2(latencyAggregate.sum / float64(latencyAggregate.count))
		}
	}
	return report, nil
}

func scanPublicMonitorIncident(row rowScanner) (*domain.PublicMonitorIncident, error) {
	var incident domain.PublicMonitorIncident
	var startedAt time.Time
	var resolvedAt sql.NullTime
	if err := row.Scan(&incident.ID, &incident.MonitorID, &incident.MonitorName, &incident.MonitorType, &incident.TargetURL, &incident.Status,
		&startedAt, &resolvedAt, &incident.DurationSeconds, &incident.InitialStatusCode, &incident.FinalStatusCode,
		&incident.FirstError, &incident.LastError, &incident.ResolutionReason); err != nil {
		return nil, err
	}
	incident.StartedAt = startedAt.UTC().Format(time.RFC3339)
	if resolvedAt.Valid {
		incident.ResolvedAt = resolvedAt.Time.UTC().Format(time.RFC3339)
	}
	return &incident, nil
}

func publicMonitorIncidentID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buf)
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
