package repository

import (
	"govmonitor-it/backend/internal/domain"
	"time"
)

// ─── Device Repository ────────────────────────────────────────────────────────

// DeviceRepository defines all persistence operations for devices.
// Swap implementations (SQLite → PostgreSQL) without touching handlers.
type DeviceRepository interface {
	// GetAll returns all devices, optionally filtered by type/status.
	GetAll(typeFilter, statusFilter, search string) ([]domain.Device, error)

	// GetByID returns a device by its string ID.
	GetByID(id string) (*domain.Device, error)

	// GetByMAC returns a device by its MAC address (the stable identity key for DHCP devices).
	GetByMAC(mac string) (*domain.Device, error)

	// GetDHCPDevices returns all devices with addressing_mode = DHCP.
	GetDHCPDevices() ([]domain.Device, error)

	// Create inserts a new device and returns the created row.
	Create(d *domain.Device) (*domain.Device, error)

	// Update updates all mutable fields of a device.
	Update(d *domain.Device) error

	// UpdateLastKnownIP updates the last_known_ip and logs an IPChangeEvent.
	UpdateLastKnownIP(deviceID, newIP string) error

	// UpdateStatus updates a device's status field.
	UpdateStatus(deviceID string, status domain.DeviceStatus) error

	// Delete removes a device by ID.
	Delete(id string) error

	// ExistsByMAC returns true if a device with this MAC already exists.
	ExistsByMAC(mac string) (bool, error)
}

// ─── Incident Repository ──────────────────────────────────────────────────────

type IncidentRepository interface {
	GetAll() ([]domain.Incident, error)
	GetByID(id string) (*domain.Incident, error)
	GetActive() ([]domain.Incident, error)
	GetByDeviceID(deviceID string) ([]domain.Incident, error)
	Create(inc *domain.Incident) (*domain.Incident, error)
	Resolve(id, resolvedAt string) error
	ResolveActiveByDeviceID(deviceID string, resolvedAt time.Time) error
}

// ─── User Log Repository ──────────────────────────────────────────────────────

type UserLogRepository interface {
	Append(log *domain.UserLog) error
	GetRecent(limit int) ([]domain.UserLog, error)
}

// ─── Status Log Repository ────────────────────────────────────────────────────

type StatusLogRepository interface {
	// Append writes a new UP/DOWN transition record.
	Append(log *domain.DeviceStatusLog) error

	// CountInWindow returns the number of DOWN events for a device within a window.
	CountInWindow(deviceID string, status domain.DeviceStatus, from, to time.Time) (int, error)

	// SumDowntimeMinutes returns total downtime in minutes for a device in a window.
	SumDowntimeMinutes(deviceID string, from, to time.Time) (int, error)

	// GetFlapDevices returns devices with more than threshold DOWN events in the window.
	GetFlapDevices(threshold int, from, to time.Time) ([]domain.FlapReport, error)

	// GetHistoryGrouped returns aggregated status history points truncated by hour or day.
	GetHistoryGrouped(deviceID string, from, to time.Time, truncateUnit string) ([]domain.StatusHistoryPoint, error)
	GetDowntimeReport(from, to time.Time) ([]domain.DeviceDowntimeSummary, error)
}

// ─── Notification Log Repository ─────────────────────────────────────────────

type NotificationLogRepository interface {
	Append(row *domain.NotificationLogRow) error
	GetByIncidentID(incidentID string) ([]domain.NotificationLogRow, error)
}

// ─── User Repository ──────────────────────────────────────────────────────────

type UserRepository interface {
	GetAll() ([]domain.User, error)
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Create(u *domain.User, passwordHash string) error
	UpdateRole(id string, role domain.Role) error
	UpdateUser(u *domain.User) error
	DeactivateUser(id string) error
}

// ─── Device Metric Repository ──────────────────────────────────────────────────

type DeviceMetricRepository interface {
	SaveMetric(m *domain.DeviceMetric) error
	GetMetricsByDeviceID(deviceID, metricType string, from, to time.Time) ([]domain.DeviceMetric, error)
}

// ─── Notification Targets ────────────────────────────────────────────────────

type WhatsAppTargetRepository interface {
	GetAll() ([]domain.WhatsAppTarget, error)
	GetByID(id string) (*domain.WhatsAppTarget, error)
	Create(target *domain.WhatsAppTarget) (*domain.WhatsAppTarget, error)
	Update(target *domain.WhatsAppTarget) error
	Delete(id string) error
}

// ─── Location Repository ──────────────────────────────────────────────────────

type LocationRepository interface {
	GetAll(search string) ([]domain.Location, error)
	GetByID(id string) (*domain.Location, error)
	GetByName(name string) (*domain.Location, error)
	Create(loc *domain.Location) (*domain.Location, error)
}

// ─── Permission Repository ───────────────────────────────────────────────────

type PermissionRepository interface {
	GetAll() ([]domain.RolePermission, error)
	GetByRole(role domain.Role) ([]domain.RolePermission, error)
	HasPermission(role domain.Role, featureKey string) (bool, error)
	UpdatePermissions(permissions []domain.RolePermission) error
}
