package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

type PostgresPublicMonitorRepository struct {
	db *sql.DB
}

func NewPostgresPublicMonitorRepository(db *sql.DB) *PostgresPublicMonitorRepository {
	return &PostgresPublicMonitorRepository{db: db}
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanPublicMonitor(row rowScanner) (*domain.PublicMonitor, error) {
	var m domain.PublicMonitor
	var lastChecked, createdAt, updatedAt, deletedAt, purgedAt sql.NullTime
	err := row.Scan(
		&m.ID, &m.Name, &m.MonitorType, &m.TargetURL, &m.GroupName, &m.GroupID,
		&m.TargetHost, &m.TargetPort, &m.DNSRecordType,
		&m.Keyword, &m.JSONPath, &m.ExpectedValue,
		&m.IntervalSeconds, &m.TimeoutSeconds, &m.RetryCount, &m.Enabled,
		&m.NotifyOnFailure, &m.Status, &m.Uptime24h, &m.Uptime30d,
		&m.LastStatusCode, &m.LastLatencyMs, &m.LastError, &m.ConsecutiveFailures,
		&lastChecked, &createdAt, &updatedAt, &m.CreatedByUserID, &deletedAt, &m.DeletedByUserID, &m.DeletionReason,
		&purgedAt, &m.PurgedByUserID, &m.PurgeReason,
	)
	if err != nil {
		return nil, err
	}
	m.LastChecked = nullableTimeString(lastChecked)
	m.CreatedAt = nullableTimeString(createdAt)
	m.UpdatedAt = nullableTimeString(updatedAt)
	m.DeletedAt = nullableTimeString(deletedAt)
	m.PurgedAt = nullableTimeString(purgedAt)
	return &m, nil
}

func nullableTimeString(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

const publicMonitorColumns = `id, name, monitor_type, target_url, group_name, COALESCE(group_id, ''),
    COALESCE(target_host, ''), COALESCE(target_port, 0), COALESCE(dns_record_type, ''),
    COALESCE(keyword, ''), COALESCE(json_path, ''), COALESCE(expected_value, ''),
    interval_seconds, timeout_seconds, retry_count, enabled, notify_on_failure,
    status, uptime_24h, uptime_30d, last_status_code, last_latency_ms,
    COALESCE(last_error, ''), consecutive_failures, last_checked, created_at,
    updated_at, COALESCE(created_by_user_id, ''), deleted_at,
    COALESCE(deleted_by_user_id, ''), COALESCE(deletion_reason, ''), purged_at,
    COALESCE(purged_by_user_id, ''), COALESCE(purge_reason, '')`

func (r *PostgresPublicMonitorRepository) GetAll(search, status, group string, page, limit int) ([]domain.PublicMonitor, int, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitor{}, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	where := " WHERE deleted_at IS NULL AND purged_at IS NULL"
	args := make([]interface{}, 0, 6)
	idx := 1
	if search != "" {
		where += fmt.Sprintf(" AND (LOWER(name) LIKE $%d OR LOWER(target_url) LIKE $%d OR LOWER(group_name) LIKE $%d)", idx, idx+1, idx+2)
		term := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		args = append(args, term, term, term)
		idx += 3
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, strings.ToUpper(strings.TrimSpace(status)))
		idx++
	}
	if group != "" {
		where += fmt.Sprintf(" AND group_name = $%d", idx)
		args = append(args, strings.TrimSpace(group))
		idx++
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM public_monitors"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := "SELECT " + publicMonitorColumns + " FROM public_monitors" + where +
		fmt.Sprintf(" ORDER BY CASE status WHEN 'DOWN' THEN 0 WHEN 'PAUSED' THEN 1 ELSE 2 END, name ASC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.PublicMonitor, 0)
	for rows.Next() {
		m, err := scanPublicMonitor(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *m)
	}
	return items, total, rows.Err()
}

func (r *PostgresPublicMonitorRepository) GetEnabled() ([]domain.PublicMonitor, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitor{}, nil
	}
	rows, err := r.db.Query("SELECT " + publicMonitorColumns + " FROM public_monitors WHERE enabled=true AND deleted_at IS NULL AND purged_at IS NULL ORDER BY last_checked NULLS FIRST, name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitor, 0)
	for rows.Next() {
		item, err := scanPublicMonitor(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *PostgresPublicMonitorRepository) GetByID(id string) (*domain.PublicMonitor, error) {
	if r == nil || r.db == nil {
		return nil, sql.ErrNoRows
	}
	return scanPublicMonitor(r.db.QueryRow("SELECT "+publicMonitorColumns+" FROM public_monitors WHERE id=$1 AND deleted_at IS NULL AND purged_at IS NULL", id))
}

func (r *PostgresPublicMonitorRepository) GetByIDIncludingArchived(id string) (*domain.PublicMonitor, error) {
	if r == nil || r.db == nil {
		return nil, sql.ErrNoRows
	}
	return scanPublicMonitor(r.db.QueryRow("SELECT "+publicMonitorColumns+" FROM public_monitors WHERE id=$1 AND purged_at IS NULL", id))
}

func (r *PostgresPublicMonitorRepository) GetArchived(search string, page, limit int) ([]domain.PublicMonitor, int, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitor{}, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	where := " WHERE deleted_at IS NOT NULL AND purged_at IS NULL"
	args := make([]interface{}, 0, 1)
	idx := 1
	if strings.TrimSpace(search) != "" {
		where += fmt.Sprintf(" AND (LOWER(name) LIKE $%d OR LOWER(target_url) LIKE $%d OR LOWER(deletion_reason) LIKE $%d)", idx, idx+1, idx+2)
		term := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		args = append(args, term, term, term)
		idx += 3
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM public_monitors"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	query := "SELECT " + publicMonitorColumns + " FROM public_monitors" + where + fmt.Sprintf(" ORDER BY deleted_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitor, 0)
	for rows.Next() {
		item, err := scanPublicMonitor(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *PostgresPublicMonitorRepository) Create(m *domain.PublicMonitor) (*domain.PublicMonitor, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("public monitor database is not initialized")
	}
	if m.ID == "" {
		m.ID = publicMonitorID("mon")
	}
	if m.MonitorType == "" {
		m.MonitorType = domain.PublicMonitorHTTP
	}
	if m.GroupName == "" {
		m.GroupName = "Ungrouped"
	}
	if m.Status == "" {
		m.Status = domain.PublicMonitorPaused
	}
	return scanPublicMonitor(r.db.QueryRow(
		"INSERT INTO public_monitors (id, name, monitor_type, target_url, group_name, group_id, target_host, target_port, dns_record_type, keyword, json_path, expected_value, interval_seconds, timeout_seconds, retry_count, enabled, notify_on_failure, status, created_by_user_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19) RETURNING "+publicMonitorColumns,
		m.ID, m.Name, m.MonitorType, m.TargetURL, m.GroupName, m.GroupID, m.TargetHost, nullableInt(m.TargetPort), nullableString(m.DNSRecordType), nullableString(m.Keyword), nullableString(m.JSONPath), nullableString(m.ExpectedValue), m.IntervalSeconds, m.TimeoutSeconds, m.RetryCount, m.Enabled, m.NotifyOnFailure, m.Status, m.CreatedByUserID,
	))
}

func (r *PostgresPublicMonitorRepository) Update(m *domain.PublicMonitor) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("public monitor database is not initialized")
	}
	_, err := r.db.Exec(`UPDATE public_monitors SET name=$1, monitor_type=$2, target_url=$3, group_name=$4,
		group_id=$5, target_host=$6, target_port=$7, dns_record_type=$8,
		keyword=$9, json_path=$10, expected_value=$11,
		interval_seconds=$12, timeout_seconds=$13, retry_count=$14, enabled=$15,
		notify_on_failure=$16, status=$17, updated_at=CURRENT_TIMESTAMP WHERE id=$18`,
		m.Name, m.MonitorType, m.TargetURL, m.GroupName, m.GroupID, m.TargetHost, nullableInt(m.TargetPort), nullableString(m.DNSRecordType), nullableString(m.Keyword), nullableString(m.JSONPath), nullableString(m.ExpectedValue),
		m.IntervalSeconds, m.TimeoutSeconds, m.RetryCount, m.Enabled, m.NotifyOnFailure, m.Status, m.ID)
	return err
}

func nullableInt(value int) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func nullableString(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (r *PostgresPublicMonitorRepository) Delete(id string) error {
	return r.Archive(id, "", "Deleted by user")
}

func (r *PostgresPublicMonitorRepository) Archive(id, deletedByUserID, reason string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("public monitor database is not initialized")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE public_monitors SET enabled=false, status='PAUSED',
        deleted_at=COALESCE(deleted_at, CURRENT_TIMESTAMP), deleted_by_user_id=COALESCE(NULLIF($2, ''), deleted_by_user_id),
        deletion_reason=CASE WHEN NULLIF($3, '') IS NULL THEN COALESCE(deletion_reason, 'Deleted by user') ELSE $3 END,
        updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND deleted_at IS NULL`, id, deletedByUserID, strings.TrimSpace(reason))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	archiveReason := strings.TrimSpace(reason)
	if archiveReason == "" {
		archiveReason = "Deleted by user"
	}
	if _, err := tx.Exec(`INSERT INTO public_monitor_events (id, monitor_id, event_type, status, message, occurred_at)
        VALUES ($1,$2,'archived','PAUSED',$3,CURRENT_TIMESTAMP)`, publicMonitorID("evt"), id, archiveReason); err != nil {
		return err
	}
	return tx.Commit()
}

// Purge removes a monitor from the active and archived catalogs while retaining
// the monitor row and all related checks, events, incidents, and reports for audit.
// This is deliberately a tombstone instead of SQL DELETE because the related
// tables use cascading foreign keys.
func (r *PostgresPublicMonitorRepository) Purge(id, purgedByUserID, reason string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("public monitor database is not initialized")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "Monitor configuration removed"
	}
	result, err := tx.Exec(`UPDATE public_monitors SET enabled=false, status='PAUSED',
        purged_at=COALESCE(purged_at, CURRENT_TIMESTAMP), purged_by_user_id=COALESCE(NULLIF($2, ''), purged_by_user_id),
        purge_reason=CASE WHEN NULLIF($3, '') IS NULL THEN COALESCE(purge_reason, 'Monitor configuration removed') ELSE $3 END,
        updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND deleted_at IS NOT NULL AND purged_at IS NULL`, id, purgedByUserID, reason)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(`INSERT INTO public_monitor_events (id, monitor_id, event_type, status, message, occurred_at)
        VALUES ($1,$2,'purged','PAUSED',$3,CURRENT_TIMESTAMP)`, publicMonitorID("evt"), id, reason); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresPublicMonitorRepository) Restore(id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("public monitor database is not initialized")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE public_monitors SET enabled=false, status='PAUSED', deleted_at=NULL,
        deleted_by_user_id=NULL, deletion_reason='', last_error='', consecutive_failures=0,
        updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(`INSERT INTO public_monitor_events (id, monitor_id, event_type, status, message, occurred_at)
        VALUES ($1,$2,'restored','PAUSED','Monitor restored from archive; awaiting first check',CURRENT_TIMESTAMP)`, publicMonitorID("evt"), id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresPublicMonitorRepository) RecordCheck(monitorID string, status domain.PublicMonitorStatus, statusCode, latencyMs int, errorMessage string, checkedAt time.Time) (string, bool, error) {
	if r == nil || r.db == nil {
		return "", false, fmt.Errorf("public monitor database is not initialized")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()

	var previous string
	var consecutiveFailures int
	var enabled bool
	var deletedAt sql.NullTime
	err = tx.QueryRow("SELECT status, consecutive_failures, enabled, deleted_at FROM public_monitors WHERE id=$1 FOR UPDATE", monitorID).Scan(&previous, &consecutiveFailures, &enabled, &deletedAt)
	if err != nil {
		return "", false, err
	}
	if !enabled || deletedAt.Valid {
		return "", false, fmt.Errorf("public monitor is archived or disabled")
	}

	if status == domain.PublicMonitorUp {
		consecutiveFailures = 0
	} else {
		consecutiveFailures++
	}
	checkID := publicMonitorID("chk")
	if _, err = tx.Exec(`INSERT INTO public_monitor_checks (id, monitor_id, checked_at, status, status_code, latency_ms, error_message)
        VALUES ($1,$2,$3,$4,$5,$6,$7)`, checkID, monitorID, checkedAt, status, statusCode, latencyMs, errorMessage); err != nil {
		return "", false, err
	}

	var uptime24h, uptime30d sql.NullFloat64
	_ = tx.QueryRow(`SELECT AVG(CASE WHEN status='UP' THEN 100.0 ELSE 0.0 END)
        FROM public_monitor_checks WHERE monitor_id=$1 AND checked_at >= $2`, monitorID, checkedAt.Add(-24*time.Hour)).Scan(&uptime24h)
	_ = tx.QueryRow(`SELECT AVG(CASE WHEN status='UP' THEN 100.0 ELSE 0.0 END)
        FROM public_monitor_checks WHERE monitor_id=$1 AND checked_at >= $2`, monitorID, checkedAt.Add(-30*24*time.Hour)).Scan(&uptime30d)
	if !uptime24h.Valid {
		uptime24h.Float64 = 100
	}
	if !uptime30d.Valid {
		uptime30d.Float64 = 100
	}
	if _, err = tx.Exec(`UPDATE public_monitors SET status=$1, uptime_24h=$2, uptime_30d=$3,
        last_status_code=$4, last_latency_ms=$5, last_error=$6, consecutive_failures=$7,
        last_checked=$8, updated_at=CURRENT_TIMESTAMP WHERE id=$9`,
		status, uptime24h.Float64, uptime30d.Float64, statusCode, latencyMs, errorMessage, consecutiveFailures, checkedAt, monitorID); err != nil {
		return "", false, err
	}

	changed := previous != "" && previous != string(status) && previous != string(domain.PublicMonitorPaused)
	if changed {
		eventType := "recovered"
		if status == domain.PublicMonitorDown {
			eventType = "down"
		}
		message := errorMessage
		if message == "" && status == domain.PublicMonitorUp {
			message = "Endpoint responded successfully"
		}
		_, err = tx.Exec(`INSERT INTO public_monitor_events (id, monitor_id, event_type, status, message, occurred_at)
            VALUES ($1,$2,$3,$4,$5,$6)`, publicMonitorID("evt"), monitorID, eventType, status, message, checkedAt)
		if err != nil {
			return "", false, err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", false, err
	}
	return previous, changed, nil
}

func (r *PostgresPublicMonitorRepository) GetChecks(monitorID string, since time.Time, limit int) ([]domain.PublicMonitorCheck, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitorCheck{}, nil
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.Query(`SELECT id, monitor_id, checked_at, status, status_code, latency_ms, COALESCE(error_message, '')
        FROM public_monitor_checks WHERE monitor_id=$1 AND checked_at >= $2 ORDER BY checked_at DESC LIMIT $3`, monitorID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitorCheck, 0)
	for rows.Next() {
		var item domain.PublicMonitorCheck
		var checkedAt time.Time
		if err := rows.Scan(&item.ID, &item.MonitorID, &checkedAt, &item.Status, &item.StatusCode, &item.LatencyMs, &item.ErrorMessage); err != nil {
			return nil, err
		}
		item.CheckedAt = checkedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresPublicMonitorRepository) GetChecksPage(monitorID string, since time.Time, page, limit int) ([]domain.PublicMonitorCheck, int, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitorCheck{}, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if since.IsZero() {
		since = time.Now().UTC().Add(-90 * 24 * time.Hour)
	}

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM public_monitor_checks WHERE monitor_id=$1 AND checked_at >= $2`, monitorID, since).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	rows, err := r.db.Query(`SELECT id, monitor_id, checked_at, status, status_code, latency_ms, COALESCE(error_message, '')
        FROM public_monitor_checks WHERE monitor_id=$1 AND checked_at >= $2
        ORDER BY checked_at DESC, id DESC LIMIT $3 OFFSET $4`, monitorID, since, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitorCheck, 0, limit)
	for rows.Next() {
		var item domain.PublicMonitorCheck
		var checkedAt time.Time
		if err := rows.Scan(&item.ID, &item.MonitorID, &checkedAt, &item.Status, &item.StatusCode, &item.LatencyMs, &item.ErrorMessage); err != nil {
			return nil, 0, err
		}
		item.CheckedAt = checkedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *PostgresPublicMonitorRepository) GetEvents(monitorID string, limit int) ([]domain.PublicMonitorEvent, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitorEvent{}, nil
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, monitor_id, event_type, status, message, occurred_at
        FROM public_monitor_events WHERE monitor_id=$1 ORDER BY occurred_at DESC LIMIT $2`, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.PublicMonitorEvent, 0)
	for rows.Next() {
		var item domain.PublicMonitorEvent
		var occurredAt time.Time
		if err := rows.Scan(&item.ID, &item.MonitorID, &item.EventType, &item.Status, &item.Message, &occurredAt); err != nil {
			return nil, err
		}
		item.OccurredAt = occurredAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func publicMonitorID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buf)
}
