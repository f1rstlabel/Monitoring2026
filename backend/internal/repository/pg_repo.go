package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

type PostgresDeviceRepository struct {
	db    *sql.DB
	redis *RedisClient
}

func NewPostgresDeviceRepository(db *sql.DB, redisClient *RedisClient) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{
		db:    db,
		redis: redisClient,
	}
}

func (r *PostgresDeviceRepository) GetAll(typeFilter, statusFilter, search string) ([]domain.Device, error) {
	// Try Redis cache if filters are empty
	if typeFilter == "" && statusFilter == "" && search == "" && r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		val, err := r.redis.Client.Get(ctx, "devices:all").Result()
		cancel()
		if err == nil && val != "" {
			var cached []domain.Device
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil
			}
		}
	}

	query := `SELECT d.id, d.name, d.type, d.mac_address, COALESCE(d.last_known_ip, ''), d.addressing_mode, COALESCE(d.model, ''), COALESCE(d.firmware_status, ''), d.status, COALESCE(d.location_id, ''), COALESCE(l.name, d.location, ''), COALESCE(d.rack, ''), d.failure_threshold, d.uptime_30d, d.down_count_7d, d.down_count_30d, d.checked_seconds_ago, COALESCE(d.last_checked, ''), COALESCE(d.snmp_enabled, false), COALESCE(d.snmp_community, 'public'), COALESCE(d.snmp_port, 161), COALESCE(d.snmp_if_index, 0), COALESCE(d.snmp_sys_name, ''), COALESCE(d.snmp_sys_descr, ''), COALESCE(d.snmp_sys_uptime, ''), COALESCE(d.snmp_sys_contact, ''), COALESCE(d.snmp_sys_location, ''), COALESCE(d.use_custom_threshold, false), d.custom_failure_threshold, COALESCE(d.created_by_user_id, ''), COALESCE(u.name, '')
		FROM devices d
		LEFT JOIN locations l ON d.location_id = l.id
		LEFT JOIN users u ON d.created_by_user_id = u.id
		WHERE 1=1`
	var args []interface{}
	idx := 1

	if typeFilter != "" {
		query += fmt.Sprintf(" AND d.type = $%d", idx)
		args = append(args, typeFilter)
		idx++
	}
	if statusFilter != "" {
		query += fmt.Sprintf(" AND d.status = $%d", idx)
		args = append(args, statusFilter)
		idx++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (LOWER(d.name) LIKE $%d OR LOWER(d.last_known_ip) LIKE $%d OR LOWER(d.mac_address) LIKE $%d OR LOWER(COALESCE(l.name, d.location, '')) LIKE $%d)", idx, idx+1, idx+2, idx+3)
		searchTerm := "%" + strings.ToLower(search) + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		idx += 4
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		var d domain.Device
		err := rows.Scan(
			&d.ID, &d.Name, &d.Type, &d.MAC, &d.IP, &d.AddressingMode, &d.Model, &d.FirmwareStatus,
			&d.Status, &d.LocationID, &d.Location, &d.Rack, &d.FailureThreshold, &d.Uptime30d,
			&d.DownCount7d, &d.DownCount30d, &d.CheckedSecondsAgo, &d.LastChecked,
			&d.SNMPEnabled, &d.SNMPCommunity, &d.SNMPPort, &d.SNMPIfIndex, &d.SNMPSysName,
			&d.SNMPSysDescr, &d.SNMPSysUpTime, &d.SNMPSysContact, &d.SNMPSysLocation,
			&d.UseCustomThreshold, &d.CustomFailureThreshold, &d.CreatedByUserID, &d.CreatedByUserName,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Store in Redis if unfiltered
	if typeFilter == "" && statusFilter == "" && search == "" && r.redis != nil {
		data, _ := json.Marshal(devices)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Set(ctx, "devices:all", string(data), 10*time.Second).Err()
		cancel()
	}

	return devices, nil
}

func (r *PostgresDeviceRepository) GetByID(id string) (*domain.Device, error) {
	query := `SELECT d.id, d.name, d.type, d.mac_address, COALESCE(d.last_known_ip, ''), d.addressing_mode, COALESCE(d.model, ''), COALESCE(d.firmware_status, ''), d.status, COALESCE(d.location_id, ''), COALESCE(l.name, d.location, ''), COALESCE(d.rack, ''), d.failure_threshold, d.uptime_30d, d.down_count_7d, d.down_count_30d, d.checked_seconds_ago, COALESCE(d.last_checked, ''), COALESCE(d.snmp_enabled, false), COALESCE(d.snmp_community, 'public'), COALESCE(d.snmp_port, 161), COALESCE(d.snmp_if_index, 0), COALESCE(d.snmp_sys_name, ''), COALESCE(d.snmp_sys_descr, ''), COALESCE(d.snmp_sys_uptime, ''), COALESCE(d.snmp_sys_contact, ''), COALESCE(d.snmp_sys_location, ''), COALESCE(d.use_custom_threshold, false), d.custom_failure_threshold, COALESCE(d.created_by_user_id, ''), COALESCE(u.name, '')
		FROM devices d
		LEFT JOIN locations l ON d.location_id = l.id
		LEFT JOIN users u ON d.created_by_user_id = u.id
		WHERE d.id = $1`
	var d domain.Device
	err := r.db.QueryRow(query, id).Scan(
		&d.ID, &d.Name, &d.Type, &d.MAC, &d.IP, &d.AddressingMode, &d.Model, &d.FirmwareStatus,
		&d.Status, &d.LocationID, &d.Location, &d.Rack, &d.FailureThreshold, &d.Uptime30d,
		&d.DownCount7d, &d.DownCount30d, &d.CheckedSecondsAgo, &d.LastChecked,
		&d.SNMPEnabled, &d.SNMPCommunity, &d.SNMPPort, &d.SNMPIfIndex, &d.SNMPSysName,
		&d.SNMPSysDescr, &d.SNMPSysUpTime, &d.SNMPSysContact, &d.SNMPSysLocation,
		&d.UseCustomThreshold, &d.CustomFailureThreshold, &d.CreatedByUserID, &d.CreatedByUserName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDeviceRepository) GetByMAC(mac string) (*domain.Device, error) {
	query := `SELECT d.id, d.name, d.type, d.mac_address, COALESCE(d.last_known_ip, ''), d.addressing_mode, COALESCE(d.model, ''), COALESCE(d.firmware_status, ''), d.status, COALESCE(d.location_id, ''), COALESCE(l.name, d.location, ''), COALESCE(d.rack, ''), d.failure_threshold, d.uptime_30d, d.down_count_7d, d.down_count_30d, d.checked_seconds_ago, COALESCE(d.last_checked, ''), COALESCE(d.snmp_enabled, false), COALESCE(d.snmp_community, 'public'), COALESCE(d.snmp_port, 161), COALESCE(d.snmp_if_index, 0), COALESCE(d.snmp_sys_name, ''), COALESCE(d.snmp_sys_descr, ''), COALESCE(d.snmp_sys_uptime, ''), COALESCE(d.snmp_sys_contact, ''), COALESCE(d.snmp_sys_location, ''), COALESCE(d.use_custom_threshold, false), d.custom_failure_threshold, COALESCE(d.created_by_user_id, ''), COALESCE(u.name, '')
		FROM devices d
		LEFT JOIN locations l ON d.location_id = l.id
		LEFT JOIN users u ON d.created_by_user_id = u.id
		WHERE LOWER(d.mac_address) = LOWER($1)`
	var d domain.Device
	err := r.db.QueryRow(query, mac).Scan(
		&d.ID, &d.Name, &d.Type, &d.MAC, &d.IP, &d.AddressingMode, &d.Model, &d.FirmwareStatus,
		&d.Status, &d.LocationID, &d.Location, &d.Rack, &d.FailureThreshold, &d.Uptime30d,
		&d.DownCount7d, &d.DownCount30d, &d.CheckedSecondsAgo, &d.LastChecked,
		&d.SNMPEnabled, &d.SNMPCommunity, &d.SNMPPort, &d.SNMPIfIndex, &d.SNMPSysName,
		&d.SNMPSysDescr, &d.SNMPSysUpTime, &d.SNMPSysContact, &d.SNMPSysLocation,
		&d.UseCustomThreshold, &d.CustomFailureThreshold, &d.CreatedByUserID, &d.CreatedByUserName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDeviceRepository) GetDHCPDevices() ([]domain.Device, error) {
	return r.GetAll("", "", "")
}

func (r *PostgresDeviceRepository) resolveLocationID(locationName string) string {
	if strings.TrimSpace(locationName) == "" {
		return ""
	}
	name := strings.TrimSpace(locationName)
	var id string
	err := r.db.QueryRow(`SELECT id FROM locations WHERE LOWER(name) = LOWER($1)`, name).Scan(&id)
	if err == nil && id != "" {
		return id
	}
	newID := fmt.Sprintf("loc-%d", time.Now().UnixNano()/1e6)
	_, err = r.db.Exec(`INSERT INTO locations (id, name, description) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`,
		newID, name, "Location created from device entry")
	if err == nil {
		return newID
	}
	_ = r.db.QueryRow(`SELECT id FROM locations WHERE LOWER(name) = LOWER($1)`, name).Scan(&id)
	return id
}

func (r *PostgresDeviceRepository) Create(d *domain.Device) (*domain.Device, error) {
	if d.ID == "" {
		d.ID = fmt.Sprintf("dev-%d", time.Now().UnixNano()/1e6)
	}
	if d.LocationID == "" && d.Location != "" {
		d.LocationID = r.resolveLocationID(d.Location)
	}
	var locID interface{}
	if d.LocationID != "" {
		locID = d.LocationID
	}
	var createdBy interface{}
	if d.CreatedByUserID != "" {
		createdBy = d.CreatedByUserID
	}
	query := `INSERT INTO devices (id, name, type, mac_address, last_known_ip, addressing_mode, model, status, location_id, location, rack, failure_threshold, uptime_30d, snmp_enabled, snmp_community, snmp_port, snmp_if_index, snmp_sys_name, use_custom_threshold, custom_failure_threshold, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.Exec(query, d.ID, d.Name, d.Type, d.MAC, d.IP, d.AddressingMode, d.Model, d.Status, locID, d.Location, d.Rack, d.FailureThreshold, d.Uptime30d, d.SNMPEnabled, d.SNMPCommunity, d.SNMPPort, d.SNMPIfIndex, d.SNMPSysName, d.UseCustomThreshold, d.CustomFailureThreshold, createdBy)
	if err != nil {
		return nil, err
	}

	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Del(ctx, "devices:all").Err()
		cancel()
	}
	return d, nil
}

func (r *PostgresDeviceRepository) Update(d *domain.Device) error {
	if d.LocationID == "" && d.Location != "" {
		d.LocationID = r.resolveLocationID(d.Location)
	}
	var locID interface{}
	if d.LocationID != "" {
		locID = d.LocationID
	}
	query := `UPDATE devices SET name=$1, type=$2, last_known_ip=$3, addressing_mode=$4, location_id=$5, location=$6, rack=$7, failure_threshold=$8, snmp_enabled=$9, snmp_community=$10, snmp_port=$11, snmp_if_index=$12, snmp_sys_name=$13, use_custom_threshold=$14, custom_failure_threshold=$15, updated_at=CURRENT_TIMESTAMP WHERE id=$16`
	_, err := r.db.Exec(query, d.Name, d.Type, d.IP, d.AddressingMode, locID, d.Location, d.Rack, d.FailureThreshold, d.SNMPEnabled, d.SNMPCommunity, d.SNMPPort, d.SNMPIfIndex, d.SNMPSysName, d.UseCustomThreshold, d.CustomFailureThreshold, d.ID)
	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Del(ctx, "devices:all").Err()
		cancel()
	}
	return err
}

func (r *PostgresDeviceRepository) UpdateSNMPSysName(deviceID, sysName string) error {
	query := `UPDATE devices SET snmp_sys_name=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`
	_, err := r.db.Exec(query, sysName, deviceID)
	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Del(ctx, "devices:all").Err()
		cancel()
	}
	return err
}

func (r *PostgresDeviceRepository) UpdateSNMPMetadata(deviceID, sysName, sysDescr, sysUpTime, sysContact, sysLocation string) error {
	query := `UPDATE devices SET snmp_sys_name=$1, snmp_sys_descr=$2, snmp_sys_uptime=$3, snmp_sys_contact=$4, snmp_sys_location=$5, updated_at=CURRENT_TIMESTAMP WHERE id=$6`
	_, err := r.db.Exec(query, sysName, sysDescr, sysUpTime, sysContact, sysLocation, deviceID)
	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Del(ctx, "devices:all").Err()
		cancel()
	}
	return err
}

func (r *PostgresDeviceRepository) UpdateLastKnownIP(deviceID, newIP string) error {
	var oldIP string
	_ = r.db.QueryRow(`SELECT COALESCE(last_known_ip, '') FROM devices WHERE id=$1`, deviceID).Scan(&oldIP)
	query := `UPDATE devices SET last_known_ip=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`
	_, err := r.db.Exec(query, newIP, deviceID)
	if err == nil && oldIP != "" && oldIP != newIP {
		log.Printf("[IPChangeLog] Recording IP change for device %s: %s -> %s", deviceID, oldIP, newIP)
		_, logErr := r.db.Exec(`INSERT INTO ip_change_log (id, device_id, old_ip, new_ip, created_at) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)`,
			fmt.Sprintf("ipchange-%d", time.Now().UnixNano()), deviceID, oldIP, newIP)
		if logErr != nil {
			log.Printf("[IPChangeLog] Error writing to ip_change_log: %v", logErr)
		}
	}
	return err
}

func (r *PostgresDeviceRepository) UpdateStatus(deviceID string, status domain.DeviceStatus) error {
	query := `UPDATE devices SET status=$1, last_checked=$2, checked_seconds_ago=0, updated_at=CURRENT_TIMESTAMP WHERE id=$3`
	nowStr := time.Now().Format("15:04:05 WIB")
	_, err := r.db.Exec(query, status, nowStr, deviceID)

	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Set(ctx, fmt.Sprintf("device:status:%s", deviceID), string(status), 60*time.Second).Err()
		_ = r.redis.Client.Del(ctx, "devices:all").Err()
		cancel()
	}
	return err
}

func (r *PostgresDeviceRepository) Delete(id string) error {
	query := `DELETE FROM devices WHERE id=$1`
	_, err := r.db.Exec(query, id)
	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = r.redis.Client.Del(ctx, "devices:all").Err()
		cancel()
	}
	return err
}

func (r *PostgresDeviceRepository) ExistsByMAC(mac string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM devices WHERE LOWER(mac_address) = LOWER($1))`
	err := r.db.QueryRow(query, mac).Scan(&exists)
	return exists, err
}

// ─── User Repository PostgreSQL ───────────────────────────────────────────────

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetAll() ([]domain.User, error) {
	rows, err := r.db.Query(`SELECT id, COALESCE(username, ''), name, email, role, status, COALESCE(avatar_url, ''), COALESCE(last_active, ''), COALESCE(permissions, '[]'::jsonb), COALESCE(is_active, true), COALESCE(mfa_enabled, false) FROM users WHERE COALESCE(is_active, true) = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var permJSON []byte
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.Role, &u.Status, &u.AvatarURL, &u.LastActive, &permJSON, &u.IsActive, &u.MFAEnabled); err != nil {
			return nil, err
		}
		if u.Username == "" && u.Email != "" {
			parts := strings.Split(u.Email, "@")
			u.Username = parts[0]
		}
		_ = json.Unmarshal(permJSON, &u.Permissions)
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresUserRepository) GetByID(id string) (*domain.User, error) {
	var u domain.User
	var permJSON []byte
	err := r.db.QueryRow(`SELECT id, COALESCE(username, ''), name, email, role, status, COALESCE(avatar_url, ''), COALESCE(last_active, ''), COALESCE(permissions, '[]'::jsonb), COALESCE(is_active, true), COALESCE(mfa_enabled, false), COALESCE(mfa_secret, '') FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.Role, &u.Status, &u.AvatarURL, &u.LastActive, &permJSON, &u.IsActive, &u.MFAEnabled, &u.MFASecret)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if u.Username == "" && u.Email != "" {
		parts := strings.Split(u.Email, "@")
		u.Username = parts[0]
	}
	_ = json.Unmarshal(permJSON, &u.Permissions)
	return &u, err
}

func (r *PostgresUserRepository) GetByEmail(email string) (*domain.User, error) {
	u, _, err := r.GetWithPasswordByEmail(email)
	return u, err
}

func (r *PostgresUserRepository) GetWithPasswordByEmail(email string) (*domain.User, string, error) {
	return r.GetWithPasswordByUsernameOrEmail(email)
}

func (r *PostgresUserRepository) GetWithPasswordByUsernameOrEmail(identifier string) (*domain.User, string, error) {
	var u domain.User
	var passwordHash string
	var permJSON []byte
	clean := strings.TrimSpace(strings.ToLower(identifier))
	err := r.db.QueryRow(`SELECT id, COALESCE(username, ''), name, email, COALESCE(password, ''), role, status, COALESCE(avatar_url, ''), COALESCE(last_active, ''), COALESCE(permissions, '[]'::jsonb), COALESCE(is_active, true), COALESCE(mfa_enabled, false), COALESCE(mfa_secret, '') FROM users WHERE (LOWER(email)=$1 OR LOWER(username)=$1) AND COALESCE(is_active, true) = true`, clean).
		Scan(&u.ID, &u.Username, &u.Name, &u.Email, &passwordHash, &u.Role, &u.Status, &u.AvatarURL, &u.LastActive, &permJSON, &u.IsActive, &u.MFAEnabled, &u.MFASecret)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if u.Username == "" && u.Email != "" {
		parts := strings.Split(u.Email, "@")
		u.Username = parts[0]
	}
	_ = json.Unmarshal(permJSON, &u.Permissions)
	return &u, passwordHash, err
}

func (r *PostgresUserRepository) Create(u *domain.User, passwordHash string) error {
	if u.ID == "" {
		u.ID = fmt.Sprintf("u-%d", time.Now().UnixNano()/1e6)
	}
	if u.Username == "" {
		parts := strings.Split(u.Email, "@")
		u.Username = parts[0]
	}
	permBytes, _ := json.Marshal(u.Permissions)
	if len(permBytes) == 0 {
		permBytes = []byte("[]")
	}
	_, err := r.db.Exec(`INSERT INTO users (id, username, name, email, password, role, status, avatar_url, last_active, permissions, is_active, mfa_enabled, mfa_secret) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		u.ID, u.Username, u.Name, u.Email, passwordHash, u.Role, u.Status, u.AvatarURL, u.LastActive, permBytes, true, u.MFAEnabled, u.MFASecret)
	return err
}

func (r *PostgresUserRepository) UpdateRole(id string, role domain.Role) error {
	_, err := r.db.Exec(`UPDATE users SET role=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, role, id)
	return err
}

func (r *PostgresUserRepository) UpdateUser(u *domain.User) error {
	permBytes, _ := json.Marshal(u.Permissions)
	if len(permBytes) == 0 {
		permBytes = []byte("[]")
	}
	if u.Username == "" && u.Email != "" {
		parts := strings.Split(u.Email, "@")
		u.Username = parts[0]
	}
	_, err := r.db.Exec(`UPDATE users SET username=$1, name=$2, email=$3, role=$4, permissions=$5, updated_at=CURRENT_TIMESTAMP WHERE id=$6`,
		u.Username, u.Name, u.Email, u.Role, permBytes, u.ID)
	return err
}

func (r *PostgresUserRepository) UpdatePassword(id string, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, passwordHash, id)
	return err
}

func (r *PostgresUserRepository) UpdateUserProfile(id string, username string, name string, email string, avatarUrl string, newPasswordHash string) error {
	if username == "" && email != "" {
		parts := strings.Split(email, "@")
		username = parts[0]
	}
	if newPasswordHash != "" {
		_, err := r.db.Exec(`UPDATE users SET username=$1, name=$2, email=$3, avatar_url=$4, password=$5, updated_at=CURRENT_TIMESTAMP WHERE id=$6`,
			username, name, email, avatarUrl, newPasswordHash, id)
		return err
	}
	_, err := r.db.Exec(`UPDATE users SET username=$1, name=$2, email=$3, avatar_url=$4, updated_at=CURRENT_TIMESTAMP WHERE id=$5`,
		username, name, email, avatarUrl, id)
	return err
}

func (r *PostgresUserRepository) UpdateMFA(id string, enabled bool, secret string) error {
	_, err := r.db.Exec(`UPDATE users SET mfa_enabled=$1, mfa_secret=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3`, enabled, secret, id)
	return err
}

func (r *PostgresUserRepository) DeactivateUser(id string) error {
	_, err := r.db.Exec(`UPDATE users SET status='Inactive', is_active=false, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id)
	return err
}

// ─── Device Metric Repository ──────────────────────────────────────────────────

type PostgresDeviceMetricRepository struct {
	db *sql.DB
}

func NewPostgresDeviceMetricRepository(db *sql.DB) *PostgresDeviceMetricRepository {
	return &PostgresDeviceMetricRepository{db: db}
}

func (r *PostgresDeviceMetricRepository) SaveMetric(m *domain.DeviceMetric) error {
	if r == nil || r.db == nil {
		return nil
	}
	if m.ID == "" {
		m.ID = fmt.Sprintf("m-%d", time.Now().UnixNano())
	}
	if m.RecordedAt.IsZero() {
		m.RecordedAt = time.Now()
	}
	_, err := r.db.Exec(`INSERT INTO device_metrics (id, device_id, metric_type, value, recorded_at) VALUES ($1, $2, $3, $4, $5)`,
		m.ID, m.DeviceID, m.MetricType, m.Value, m.RecordedAt)
	return err
}

func (r *PostgresDeviceMetricRepository) GetMetricsByDeviceID(deviceID, metricType string, from, to time.Time) ([]domain.DeviceMetric, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	query := `SELECT id, device_id, metric_type, value, recorded_at FROM device_metrics WHERE device_id = $1 AND metric_type = $2 AND recorded_at >= $3 AND recorded_at <= $4 ORDER BY recorded_at ASC`
	rows, err := r.db.Query(query, deviceID, metricType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.DeviceMetric
	for rows.Next() {
		var m domain.DeviceMetric
		if err := rows.Scan(&m.ID, &m.DeviceID, &m.MetricType, &m.Value, &m.RecordedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func SeedPostgresData(db *sql.DB) {
	log.Println("[PostgreSQL] Checking database table counts and seeding initial inventory...")
}

// ─── WhatsApp Target Repository ──────────────────────────────────────────────

type PostgresWhatsAppTargetRepository struct {
	db *sql.DB
}

func NewPostgresWhatsAppTargetRepository(db *sql.DB) *PostgresWhatsAppTargetRepository {
	return &PostgresWhatsAppTargetRepository{db: db}
}

func (r *PostgresWhatsAppTargetRepository) GetAll() ([]domain.WhatsAppTarget, error) {
	rows, err := r.db.Query(`SELECT id, label, phone_number, jid, created_at, updated_at FROM whatsapp_targets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []domain.WhatsAppTarget
	for rows.Next() {
		var t domain.WhatsAppTarget
		if err := rows.Scan(&t.ID, &t.Label, &t.PhoneNumber, &t.JID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

func (r *PostgresWhatsAppTargetRepository) GetByID(id string) (*domain.WhatsAppTarget, error) {
	var t domain.WhatsAppTarget
	err := r.db.QueryRow(`SELECT id, label, phone_number, jid, created_at, updated_at FROM whatsapp_targets WHERE id=$1`, id).
		Scan(&t.ID, &t.Label, &t.PhoneNumber, &t.JID, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *PostgresWhatsAppTargetRepository) Create(target *domain.WhatsAppTarget) (*domain.WhatsAppTarget, error) {
	if target.ID == "" {
		target.ID = fmt.Sprintf("wat-%d", time.Now().UnixNano()/1e6)
	}
	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()
	_, err := r.db.Exec(`INSERT INTO whatsapp_targets (id, label, phone_number, jid, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		target.ID, target.Label, target.PhoneNumber, target.JID, target.CreatedAt, target.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (r *PostgresWhatsAppTargetRepository) Update(target *domain.WhatsAppTarget) error {
	target.UpdatedAt = time.Now()
	_, err := r.db.Exec(`UPDATE whatsapp_targets SET label=$1, phone_number=$2, jid=$3, updated_at=$4 WHERE id=$5`,
		target.Label, target.PhoneNumber, target.JID, target.UpdatedAt, target.ID)
	return err
}

func (r *PostgresWhatsAppTargetRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM whatsapp_targets WHERE id=$1`, id)
	return err
}

// ─── Status Log Repository PostgreSQL ───────────────────────────────────────

type PostgresStatusLogRepository struct {
	db *sql.DB
}

func NewPostgresStatusLogRepository(db *sql.DB) *PostgresStatusLogRepository {
	return &PostgresStatusLogRepository{db: db}
}

func (r *PostgresStatusLogRepository) Append(log *domain.DeviceStatusLog) error {
	if log.ID == "" {
		log.ID = fmt.Sprintf("log-%d", time.Now().UnixNano())
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	_, err := r.db.Exec(`INSERT INTO device_status_log (id, device_id, status, created_at) VALUES ($1, $2, $3, $4)`,
		log.ID, log.DeviceID, log.Status, log.Timestamp)
	return err
}

func (r *PostgresStatusLogRepository) CountInWindow(deviceID string, status domain.DeviceStatus, from, to time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM device_status_log WHERE device_id=$1 AND status=$2 AND created_at >= $3 AND created_at <= $4`,
		deviceID, status, from, to).Scan(&count)
	return count, err
}

func (r *PostgresStatusLogRepository) SumDowntimeMinutes(deviceID string, from, to time.Time) (int, error) {
	count, err := r.CountInWindow(deviceID, domain.StatusDOWN, from, to)
	return count * 8, err
}

func (r *PostgresStatusLogRepository) GetFlapDevices(threshold int, from, to time.Time) ([]domain.FlapReport, error) {
	query := `
		SELECT l.device_id, COUNT(*) as down_count,
		       COALESCE(d.name, 'Unknown Device') as device_name,
		       COALESCE(d.type, 'Access Point') as device_type,
		       COALESCE(loc.name, d.location, 'Gedung Sate / NOC Server Room') as location,
		       COALESCE(d.last_known_ip, '0.0.0.0') as ip
		FROM device_status_log l
		LEFT JOIN devices d ON l.device_id = d.id
		LEFT JOIN locations loc ON d.location_id = loc.id
		WHERE l.status='DOWN' AND l.created_at >= $1 AND l.created_at <= $2
		GROUP BY l.device_id, d.name, d.type, loc.name, d.location, d.last_known_ip
		HAVING COUNT(*) >= $3
	`
	rows, err := r.db.Query(query, from, to, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.FlapReport
	for rows.Next() {
		var fr domain.FlapReport
		if err := rows.Scan(&fr.DeviceID, &fr.DownCount7d, &fr.DeviceName, &fr.DeviceType, &fr.Location, &fr.IP); err != nil {
			return nil, err
		}
		fr.TotalDowntimeMinutes = fr.DownCount7d * 8
		reports = append(reports, fr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *PostgresStatusLogRepository) GetHistoryGrouped(deviceID string, from, to time.Time, truncateUnit string) ([]domain.StatusHistoryPoint, error) {
	var dateFormat string
	if truncateUnit == "hour" {
		dateFormat = "HH24:00"
	} else {
		dateFormat = "Mon DD"
	}

	query := fmt.Sprintf(`SELECT TO_CHAR(date_trunc('%s', created_at), '%s') as period,
		COUNT(*) FILTER (WHERE status = 'UP') as up_count,
		COUNT(*) FILTER (WHERE status = 'DOWN') as down_count
		FROM device_status_log
		WHERE ($1 = '' OR device_id = $1) AND created_at >= $2 AND created_at <= $3
		GROUP BY date_trunc('%s', created_at)
		ORDER BY date_trunc('%s', created_at) ASC`, truncateUnit, dateFormat, truncateUnit, truncateUnit)

	rows, err := r.db.Query(query, deviceID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []domain.StatusHistoryPoint
	for rows.Next() {
		var p domain.StatusHistoryPoint
		if err := rows.Scan(&p.Date, &p.UpCount, &p.DownCount); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

func (r *PostgresStatusLogRepository) GetDowntimeReport(from, to time.Time) ([]domain.DeviceDowntimeSummary, error) {
	query := `
		SELECT d.id as device_id,
		       COALESCE(d.name, 'Unknown Device') as device_name,
		       COALESCE(d.type, 'Access Point') as device_type,
		       COALESCE(loc.name, d.location, 'Gedung Sate / NOC Server Room') as location,
		       COUNT(l.id) as down_count,
		       COALESCE(SUM(CASE WHEN l.status='DOWN' THEN 8 ELSE 0 END), 0) as total_downtime_minutes,
		       COALESCE(MAX(l.created_at), '1970-01-01 00:00:00+00')::timestamp with time zone as last_down
		FROM devices d
		LEFT JOIN locations loc ON d.location_id = loc.id
		LEFT JOIN device_status_log l ON d.id = l.device_id AND l.status='DOWN' AND l.created_at >= $1 AND l.created_at <= $2
		GROUP BY d.id, d.name, d.type, loc.name, d.location
		ORDER BY total_downtime_minutes DESC, down_count DESC
	`
	rows, err := r.db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DeviceDowntimeSummary
	for rows.Next() {
		var s domain.DeviceDowntimeSummary
		var lastDown time.Time
		if err := rows.Scan(&s.DeviceID, &s.DeviceName, &s.DeviceType, &s.Location, &s.DownCount, &s.TotalDowntimeMinutes, &lastDown); err != nil {
			return nil, err
		}
		if lastDown.Year() <= 1970 {
			s.LastDown = "N/A"
		} else {
			s.LastDown = lastDown.Format("2006-01-02 15:04")
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// ─── Postgres Incident Repository ─────────────────────────────────────────────

type PostgresIncidentRepository struct {
	db *sql.DB
}

func NewPostgresIncidentRepository(db *sql.DB) *PostgresIncidentRepository {
	return &PostgresIncidentRepository{db: db}
}

func (r *PostgresIncidentRepository) Create(inc *domain.Incident) (*domain.Incident, error) {
	if r == nil || r.db == nil {
		return inc, nil
	}
	if inc.ID == "" {
		inc.ID = fmt.Sprintf("INC-%d", time.Now().UnixNano()/1e6)
	}
	startedAt := time.Now()
	_, err := r.db.Exec(`INSERT INTO incidents (id, device_id, status, packet_loss, latency_ms, affected_devices_count, dependencies_count, started_at)
		VALUES ($1, $2, 'ACTIVE', $3, $4, $5, $6, $7)`,
		inc.ID, inc.DeviceID, inc.PacketLoss, inc.LatencyMs, max(inc.AffectedDevicesCount, 1), inc.DependenciesCount, startedAt)
	if err != nil {
		return nil, err
	}
	inc.Status = "ACTIVE"
	inc.StartTime = startedAt.Format("15:04:05 WIB")
	return inc, nil
}

func (r *PostgresIncidentRepository) ResolveActiveByDeviceID(deviceID string, resolvedAt time.Time) error {
	if r == nil || r.db == nil {
		return nil
	}
	_, err := r.db.Exec(`UPDATE incidents SET status='RESOLVED', resolved_at=$1 WHERE device_id=$2 AND status='ACTIVE'`,
		resolvedAt, deviceID)
	return err
}

func (r *PostgresIncidentRepository) Resolve(id, resolvedAt string) error {
	if r == nil || r.db == nil {
		return nil
	}
	tRes := time.Now()
	if t, err := time.Parse(time.RFC3339, resolvedAt); err == nil {
		tRes = t
	}
	_, err := r.db.Exec(`UPDATE incidents SET status='RESOLVED', resolved_at=$1 WHERE id=$2`, tRes, id)
	return err
}

func (r *PostgresIncidentRepository) Reopen(id string) error {
	if r == nil || r.db == nil {
		return nil
	}
	_, err := r.db.Exec(`UPDATE incidents SET status='ACTIVE', resolved_at=NULL WHERE id=$1`, id)
	return err
}


func (r *PostgresIncidentRepository) GetAll() ([]domain.Incident, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	query := `SELECT i.id, i.device_id, COALESCE(d.name, 'Unknown'), COALESCE(d.type, 'Access Point'), COALESCE(d.last_known_ip, ''),
		i.status, i.packet_loss, i.latency_ms, i.affected_devices_count, i.dependencies_count, i.started_at, i.resolved_at
		FROM incidents i
		LEFT JOIN devices d ON i.device_id = d.id
		ORDER BY i.started_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		var startedAt time.Time
		var resolvedAt sql.NullTime
		var devType string

		if err := rows.Scan(&inc.ID, &inc.DeviceID, &inc.DeviceName, &devType, &inc.DeviceIP,
			&inc.Status, &inc.PacketLoss, &inc.LatencyMs, &inc.AffectedDevicesCount, &inc.DependenciesCount,
			&startedAt, &resolvedAt); err != nil {
			return nil, err
		}
		inc.DeviceType = domain.DeviceType(devType)
		inc.StartTime = startedAt.Format("15:04:05 WIB")
		inc.StartedAt = startedAt

		if resolvedAt.Valid {
			inc.ResolvedAt = resolvedAt.Time.Format("15:04:05 WIB")
			tVal := resolvedAt.Time
			inc.ResolvedAtRaw = &tVal
			dur := resolvedAt.Time.Sub(startedAt)
			inc.Duration = fmt.Sprintf("%dm %ds", int(dur.Minutes()), int(dur.Seconds())%60)
		} else {
			dur := time.Since(startedAt)
			inc.Duration = fmt.Sprintf("%dm %ds (ongoing)", int(dur.Minutes()), int(dur.Seconds())%60)
		}
		if inc.Timeline == nil {
			inc.Timeline = []domain.EventTimelineItem{}
		}
		if inc.NotificationLog == nil {
			inc.NotificationLog = []domain.NotificationLogRow{}
		}
		list = append(list, inc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PostgresIncidentRepository) GetActive() ([]domain.Incident, error) {
	all, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	var active []domain.Incident
	for _, inc := range all {
		if inc.Status == "ACTIVE" {
			active = append(active, inc)
		}
	}
	return active, nil
}

func (r *PostgresIncidentRepository) GetByID(id string) (*domain.Incident, error) {
	all, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, nil
}

func (r *PostgresIncidentRepository) GetByDeviceID(deviceID string) ([]domain.Incident, error) {
	all, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	var res []domain.Incident
	for _, inc := range all {
		if inc.DeviceID == deviceID {
			res = append(res, inc)
		}
	}
	return res, nil
}

// GetOpenByDeviceID returns the single ACTIVE incident for a device, or nil if none exists.
func (r *PostgresIncidentRepository) GetOpenByDeviceID(deviceID string) (*domain.Incident, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	var id string
	err := r.db.QueryRow(
		`SELECT id FROM incidents WHERE device_id = $1 AND status = 'ACTIVE' LIMIT 1`,
		deviceID,
	).Scan(&id)
	if err != nil {
		// sql.ErrNoRows means no open incident — not an error for our purposes
		return nil, nil
	}
	return r.GetByID(id)
}

func (r *PostgresIncidentRepository) CreateEvent(evt *domain.IncidentEvent) error {
	if r == nil || r.db == nil {
		return nil
	}
	if evt.ID == "" {
		evt.ID = fmt.Sprintf("ev-%d", time.Now().UnixNano())
	}
	query := `
		INSERT INTO incident_events (id, incident_id, event_type, channel, detail, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, evt.ID, evt.IncidentID, evt.EventType, evt.Channel, evt.Detail, evt.OccurredAt)
	return err
}

func (r *PostgresIncidentRepository) GetEventsByIncidentID(incidentID string) ([]domain.IncidentEvent, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	query := `
		SELECT id, incident_id, event_type, COALESCE(channel, ''), detail, occurred_at
		FROM incident_events
		WHERE incident_id = $1
		ORDER BY occurred_at ASC
	`
	rows, err := r.db.Query(query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.IncidentEvent
	for rows.Next() {
		var evt domain.IncidentEvent
		var channel string
		err := rows.Scan(&evt.ID, &evt.IncidentID, &evt.EventType, &channel, &evt.Detail, &evt.OccurredAt)
		if err != nil {
			return nil, err
		}
		evt.Channel = channel
		events = append(events, evt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}



// ─── Postgres Notification Log Repository ───────────────────────────────────

type PostgresNotificationLogRepository struct {
	db *sql.DB
}

func NewPostgresNotificationLogRepository(db *sql.DB) *PostgresNotificationLogRepository {
	return &PostgresNotificationLogRepository{db: db}
}

func (r *PostgresNotificationLogRepository) Append(row *domain.NotificationLogRow) error {
	if r == nil || r.db == nil {
		return nil
	}
	if row.ID == "" {
		row.ID = fmt.Sprintf("notif-%d", time.Now().UnixNano())
	}
	sentAt := time.Now()
	status := row.Status
	switch strings.ToLower(status) {
	case "delivered":
		status = "Delivered"
	case "failed":
		status = "Failed"
	case "sent":
		status = "Sent"
	case "pending":
		status = "Pending"
	default:
		if status == "" {
			status = "Sent"
		}
	}

	log.Printf("[NotificationLog] Appending log: id=%s incident_id=%s channel=%s recipient=%s status=%s", row.ID, row.IncidentID, row.Channel, row.Recipient, status)
	var incID sql.NullString
	if row.IncidentID != "" {
		incID = sql.NullString{String: row.IncidentID, Valid: true}
	}
	_, err := r.db.Exec(`INSERT INTO notification_log (id, incident_id, channel, recipient, status, error_msg, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		row.ID, incID, row.Channel, row.Recipient, status, row.ChannelIcon, sentAt)
	if err != nil {
		log.Printf("[NotificationLog] DB write FAILED for id=%s: %v", row.ID, err)
	} else {
		log.Printf("[NotificationLog] DB write Succeeded for id=%s", row.ID)
	}
	return err
}

func (r *PostgresNotificationLogRepository) GetByIncidentID(incidentID string) ([]domain.NotificationLogRow, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	query := `SELECT id, channel, recipient, status, COALESCE(error_msg, ''), TO_CHAR(sent_at, 'HH24:MI:SS WIB')
		FROM notification_log WHERE ($1 = '' OR incident_id = $1) ORDER BY sent_at DESC LIMIT 50`
	rows, err := r.db.Query(query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.NotificationLogRow
	for rows.Next() {
		var row domain.NotificationLogRow
		if err := rows.Scan(&row.ID, &row.Channel, &row.Recipient, &row.Status, &row.ChannelIcon, &row.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// ─── Postgres User Log Repository ─────────────────────────────────────────────

type PostgresUserLogRepository struct {
	db *sql.DB
}

func NewPostgresUserLogRepository(db *sql.DB) *PostgresUserLogRepository {
	return &PostgresUserLogRepository{db: db}
}

func (r *PostgresUserLogRepository) Append(ul *domain.UserLog) error {
	if r == nil || r.db == nil {
		return nil
	}
	if ul.ID == "" {
		ul.ID = fmt.Sprintf("ulog-%d", time.Now().UnixNano())
	}
	if ul.OccurredAt.IsZero() {
		ul.OccurredAt = time.Now()
	}
	if ul.IPAddress == "::1" || ul.IPAddress == "localhost" || ul.IPAddress == "" {
		ul.IPAddress = "127.0.0.1"
	}
	query := `INSERT INTO user_logs (id, user_id, action, detail, ip_address, user_agent, session_id, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query, ul.ID, ul.UserID, ul.Action, ul.Detail, ul.IPAddress, ul.UserAgent, ul.SessionID, ul.OccurredAt)
	if err != nil {
		log.Printf("[UserLog] ERROR inserting user_log: %v", err)
	}
	return err
}

func (r *PostgresUserLogRepository) GetRecent(limit int) ([]domain.UserLog, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT l.id, COALESCE(l.user_id, ''), COALESCE(u.name, 'System/Guest'), l.action, COALESCE(l.detail, ''),
		COALESCE(l.ip_address, ''), COALESCE(l.user_agent, ''), COALESCE(l.session_id, ''), l.occurred_at
		FROM user_logs l
		LEFT JOIN users u ON l.user_id = u.id
		ORDER BY l.occurred_at DESC LIMIT $1`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.UserLog
	for rows.Next() {
		var ul domain.UserLog
		if err := rows.Scan(&ul.ID, &ul.UserID, &ul.UserName, &ul.Action, &ul.Detail, &ul.IPAddress, &ul.UserAgent, &ul.SessionID, &ul.OccurredAt); err != nil {
			return nil, err
		}
		if ul.IPAddress == "::1" || ul.IPAddress == "localhost" || ul.IPAddress == "" {
			ul.IPAddress = "127.0.0.1"
		}
		logs = append(logs, ul)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// ─── Location Repository PostgreSQL ─────────────────────────────────────────

type PostgresLocationRepository struct {
	db *sql.DB
}

func SyncLocationsFromDevices(db *sql.DB) error {
	if db == nil {
		return nil
	}
	// 1. Populate locations table dynamically from distinct location strings in devices table
	_, _ = db.Exec(`
		INSERT INTO locations (id, name, description)
		SELECT 'loc-' || MD5(TRIM(location)), TRIM(location), 'Location from registered devices'
		FROM devices
		WHERE location IS NOT NULL AND TRIM(location) != ''
		ON CONFLICT (name) DO NOTHING
	`)

	// 2. Update devices.location_id foreign key relation
	_, _ = db.Exec(`
		UPDATE devices d
		SET location_id = l.id
		FROM locations l
		WHERE LOWER(TRIM(d.location)) = LOWER(TRIM(l.name)) AND (d.location_id IS NULL OR d.location_id = '')
	`)

	// 3. Purge dummy locations not in devices table
	_, _ = db.Exec(`
		DELETE FROM locations
		WHERE id NOT IN (SELECT DISTINCT location_id FROM devices WHERE location_id IS NOT NULL AND location_id != '')
		  AND name NOT IN (SELECT DISTINCT location FROM devices WHERE location IS NOT NULL AND location != '')
	`)

	log.Println("[PostgreSQL] Synced real device locations & linked foreign key relations.")
	return nil
}

func NewPostgresLocationRepository(db *sql.DB) *PostgresLocationRepository {
	_ = SyncLocationsFromDevices(db)
	return &PostgresLocationRepository{db: db}
}

func (r *PostgresLocationRepository) GetAll(search string) ([]domain.Location, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	query := `SELECT id, name, COALESCE(description, ''), created_at FROM locations WHERE 1=1`
	var args []interface{}
	if search != "" {
		query += ` AND LOWER(name) LIKE $1`
		args = append(args, "%"+strings.ToLower(search)+"%")
	}
	query += ` ORDER BY name ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []domain.Location
	existingNames := make(map[string]bool)
	for rows.Next() {
		var l domain.Location
		if err := rows.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt); err != nil {
			return nil, err
		}
		locs = append(locs, l)
		existingNames[strings.ToLower(l.Name)] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Also fetch distinct locations registered in devices table
	devQuery := `SELECT DISTINCT location FROM devices WHERE location IS NOT NULL AND location != ''`
	devRows, devErr := r.db.Query(devQuery)
	if devErr == nil {
		defer devRows.Close()
		for devRows.Next() {
			var locName string
			if err := devRows.Scan(&locName); err == nil && locName != "" {
				lower := strings.ToLower(locName)
				if !existingNames[lower] {
					if search == "" || strings.Contains(lower, strings.ToLower(search)) {
						locs = append(locs, domain.Location{
							ID:          "loc-dev-" + locName,
							Name:        locName,
							Description: "Location from registered devices",
							CreatedAt:   time.Now(),
						})
						existingNames[lower] = true
					}
				}
			}
		}
		if err := devRows.Err(); err != nil {
			return nil, err
		}
	}

	if locs == nil {
		locs = []domain.Location{}
	}
	return locs, nil
}

func (r *PostgresLocationRepository) GetByID(id string) (*domain.Location, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	var l domain.Location
	err := r.db.QueryRow(`SELECT id, name, COALESCE(description, ''), created_at FROM locations WHERE id=$1`, id).
		Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &l, err
}

func (r *PostgresLocationRepository) GetByName(name string) (*domain.Location, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	var l domain.Location
	err := r.db.QueryRow(`SELECT id, name, COALESCE(description, ''), created_at FROM locations WHERE LOWER(name)=LOWER($1)`, name).
		Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &l, err
}

func (r *PostgresLocationRepository) Create(loc *domain.Location) (*domain.Location, error) {
	if r == nil || r.db == nil {
		return loc, nil
	}
	if loc.ID == "" {
		loc.ID = fmt.Sprintf("loc-%d", time.Now().UnixNano()/1e6)
	}
	loc.CreatedAt = time.Now()
	_, err := r.db.Exec(`INSERT INTO locations (id, name, description, created_at) VALUES ($1, $2, $3, $4)`,
		loc.ID, loc.Name, loc.Description, loc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return loc, nil
}

// ─── Permission Repository PostgreSQL ───────────────────────────────────────

type PostgresPermissionRepository struct {
	db *sql.DB
}

func NewPostgresPermissionRepository(db *sql.DB) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{db: db}
}

func (r *PostgresPermissionRepository) GetAll() ([]domain.RolePermission, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	rows, err := r.db.Query(`SELECT role, feature_key, enabled, updated_at FROM role_permissions ORDER BY role ASC, feature_key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.RolePermission
	for rows.Next() {
		var p domain.RolePermission
		var rStr string
		if err := rows.Scan(&rStr, &p.FeatureKey, &p.Enabled, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Role = domain.Role(rStr)
		list = append(list, p)
	}
	if list == nil {
		list = []domain.RolePermission{}
	}
	return list, rows.Err()
}

func (r *PostgresPermissionRepository) GetByRole(role domain.Role) ([]domain.RolePermission, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	rows, err := r.db.Query(`SELECT role, feature_key, enabled, updated_at FROM role_permissions WHERE role=$1 ORDER BY feature_key ASC`, string(role))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.RolePermission
	for rows.Next() {
		var p domain.RolePermission
		var rStr string
		if err := rows.Scan(&rStr, &p.FeatureKey, &p.Enabled, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Role = domain.Role(rStr)
		list = append(list, p)
	}
	if list == nil {
		list = []domain.RolePermission{}
	}
	return list, rows.Err()
}

func (r *PostgresPermissionRepository) HasPermission(role domain.Role, featureKey string) (bool, error) {
	if role == domain.RoleSuperAdmin {
		return true, nil
	}
	if r == nil || r.db == nil {
		return true, nil
	}
	var enabled bool
	err := r.db.QueryRow(`SELECT enabled FROM role_permissions WHERE role=$1 AND feature_key=$2`, string(role), featureKey).Scan(&enabled)
	if err == sql.ErrNoRows {
		if role == domain.RolePimpinan {
			return strings.HasSuffix(featureKey, ".view") || featureKey == "reports.export", nil
		}
		if role == domain.RoleAnggota {
			return !strings.HasPrefix(featureKey, "settings.") && featureKey != "devices.delete" && featureKey != "devices.import", nil
		}
		return true, nil
	}
	return enabled, err
}

func (r *PostgresPermissionRepository) UpdatePermissions(permissions []domain.RolePermission) error {
	if r == nil || r.db == nil {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO role_permissions (role, feature_key, enabled, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (role, feature_key) DO UPDATE SET enabled=EXCLUDED.enabled, updated_at=CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range permissions {
		if _, err := stmt.Exec(string(p.Role), p.FeatureKey, p.Enabled); err != nil {
			return err
		}
	}
	return tx.Commit()
}


