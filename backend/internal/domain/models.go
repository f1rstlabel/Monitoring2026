package domain

import "time"

// ─── Core Enums ──────────────────────────────────────────────────────────────

type DeviceType string

const (
	AccessPoint DeviceType = "Access Point"
	Switch      DeviceType = "Switch"
	Router      DeviceType = "Router"
	SmartPower  DeviceType = "SmartPower"
	CCTV        DeviceType = "CCTV"
	NVR         DeviceType = "NVR"
)

type DeviceStatus string

const (
	StatusUP             DeviceStatus = "UP"
	StatusDOWN           DeviceStatus = "DOWN"
	StatusUnresolvedDHCP DeviceStatus = "UNRESOLVED_DHCP"
)

// Role matches JWT claim values exactly.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleSuperAdmin Role = "admin" // alias for compatibility
	RolePimpinan Role = "pimpinan"
	RoleAnggota  Role = "anggota"
)

type AddressingMode string

const (
	AddressingStatic AddressingMode = "Static"
	AddressingDHCP   AddressingMode = "DHCP"
)

// ─── Core Domain Objects ─────────────────────────────────────────────────────

type Location struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RolePermission struct {
	Role       Role      `json:"role"`
	FeatureKey string    `json:"featureKey"`
	Enabled    bool      `json:"enabled"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Device struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Type             DeviceType     `json:"type"`
	IP               string         `json:"ip"`
	MAC              string         `json:"mac"`
	Status           DeviceStatus   `json:"status"`
	AddressingMode   AddressingMode `json:"addressingMode"`
	LocationID       string         `json:"locationId,omitempty"`
	Location         string         `json:"location"`
	Rack             string         `json:"rack,omitempty"`
	Model            string         `json:"model,omitempty"`
	FirmwareStatus   string         `json:"firmwareStatus,omitempty"`
	CheckedSecondsAgo int           `json:"checkedSecondsAgo"`
	LastChecked      string         `json:"lastChecked"`
	Uptime30d        float64        `json:"uptime30d"`
	FailureThreshold int            `json:"failureThreshold"`
	DownCount7d  int `json:"downCount7d,omitempty"`
	DownCount30d int `json:"downCount30d,omitempty"`
	SNMPEnabled  bool `json:"snmpEnabled"`
	SNMPCommunity string `json:"snmpCommunity,omitempty"`
	SNMPPort     int `json:"snmpPort,omitempty"`
	SNMPIfIndex  int `json:"snmpIfIndex,omitempty"`
	SNMPSysName  string `json:"snmpSysName,omitempty"`
	SNMPSysDescr string `json:"snmpSysDescr,omitempty"`
	SNMPSysUpTime string `json:"snmpSysUpTime,omitempty"`
	SNMPSysContact string `json:"snmpSysContact,omitempty"`
	SNMPSysLocation string `json:"snmpSysLocation,omitempty"`
	UseCustomThreshold     bool `json:"useCustomThreshold"`
	CustomFailureThreshold *int `json:"customFailureThreshold,omitempty"`
	CreatedByUserID        string `json:"createdByUserId,omitempty"`
	CreatedByUserName      string `json:"createdByUserName,omitempty"`
}

type DeviceDowntimeSummary struct {
	DeviceID             string `json:"deviceId"`
	DeviceName           string `json:"deviceName"`
	DeviceType           string `json:"deviceType"`
	Location             string `json:"location"`
	DownCount            int    `json:"downCount"`
	TotalDowntimeMinutes int    `json:"totalDowntimeMinutes"`
	LastDown             string `json:"lastDown"`
}

type DashboardSummary struct {
	TotalDevices    int     `json:"totalDevices"`
	DevicesUp       int     `json:"devicesUp"`
	DevicesDown     int     `json:"devicesDown"`
	ActiveIncidents int     `json:"activeIncidents"`
	UpPercentage    float64 `json:"upPercentage"`
	DownPercentage  float64 `json:"downPercentage"`
}

type EventTimelineItem struct {
	ID          string `json:"id"`
	Timestamp   string `json:"timestamp"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // critical, warning, info
	Channel     string `json:"channel,omitempty"`
}

type NotificationLogRow struct {
	ID          string `json:"id"`
	IncidentID  string `json:"incidentId,omitempty"`
	Channel     string `json:"channel"`
	ChannelIcon string `json:"channelIcon"`
	Recipient   string `json:"recipient"`
	Status      string `json:"status"` // Delivered, Failed, Sent
	Timestamp   string `json:"timestamp"`
}

type Incident struct {
	ID                   string               `json:"id"`
	DeviceID             string               `json:"deviceId"`
	DeviceName           string               `json:"deviceName"`
	DeviceType           DeviceType           `json:"deviceType"`
	DeviceIP             string               `json:"deviceIp"`
	Location             string               `json:"location,omitempty"`
	Status               string               `json:"status"` // ACTIVE, RESOLVED
	StartTime            string               `json:"startTime"`
	Duration             string               `json:"duration"`
	AffectedDevicesCount int                  `json:"affectedDevicesCount"`
	PacketLoss           int                  `json:"packetLoss"`
	LatencyMs            int                  `json:"latencyMs"`
	DependenciesCount    int                  `json:"dependenciesCount"`
	Timeline             []EventTimelineItem  `json:"timeline"`
	NotificationLog      []NotificationLogRow `json:"notificationLog"`
	ResolvedAt           string               `json:"resolvedAt,omitempty"`
	StartedAt            time.Time            `json:"startedAt"`
	ResolvedAtRaw        *time.Time           `json:"-"`
}

type UserLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName,omitempty"`
	Action    string    `json:"action"` // 'login' | 'logout' | 'action'
	Detail    string    `json:"detail"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	SessionID string    `json:"sessionId"`
	OccurredAt time.Time `json:"occurredAt"`
}

type StatusHistoryPoint struct {
	Date      string `json:"date"`
	UpCount   int    `json:"upCount"`
	DownCount int    `json:"downCount"`
}

// ─── Audit / Log Objects ─────────────────────────────────────────────────────

// DeviceStatusLog records each UP/DOWN transition.
type DeviceStatusLog struct {
	ID        string       `json:"id"`
	DeviceID  string       `json:"deviceId"`
	Status    DeviceStatus `json:"status"`
	Timestamp time.Time    `json:"timestamp"`
}

// IPChangeEvent is written when a DHCP device gets a new IP.
type IPChangeEvent struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"deviceId"`
	OldIP     string    `json:"oldIp"`
	NewIP     string    `json:"newIp"`
	Timestamp time.Time `json:"timestamp"`
}

// ─── Notification Objects ─────────────────────────────────────────────────────

type WhatsAppTarget struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	PhoneNumber string    `json:"phoneNumber"` // E.g., +6281234567890
	JID         string    `json:"jid"`         // E.g., 6281234567890@s.whatsapp.net
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NotificationJob is the unit queued to the notification pipeline.
type NotificationJob struct {
	JobID      string    `json:"jobId"`
	IncidentID string    `json:"incidentId"`
	Message    string    `json:"message"`
	Channel    string    `json:"channel"` // "whatsapp" | "telegram"
	Recipient  string    `json:"recipient"`
	CreatedAt  time.Time `json:"createdAt"`
	Retries    int       `json:"retries"`
}

// AggregatedIncident is the batched payload produced by the aggregation buffer.
type AggregatedIncident struct {
	BatchID      string         `json:"batchId"`
	RootDeviceID string         `json:"rootDeviceId"`
	RootDevice   *Device        `json:"rootDevice,omitempty"`
	Devices      []Device       `json:"devices"`
	DetectedAt   time.Time      `json:"detectedAt"`
	DebouncedAt  time.Time      `json:"debouncedAt"`
}

// FlapReport summarises recurring-down devices for the daily technician cron.
type FlapReport struct {
	DeviceID             string     `json:"deviceId"`
	DeviceName           string     `json:"deviceName"`
	DeviceType           DeviceType `json:"deviceType"`
	Location             string     `json:"location"`
	IP                   string     `json:"ip"`
	DownCount7d          int        `json:"downCount7d"`
	TotalDowntimeMinutes int        `json:"totalDowntimeMinutes"`
}

// ─── Import Objects ───────────────────────────────────────────────────────────

type ImportRowStatus string

const (
	ImportValid     ImportRowStatus = "valid"
	ImportDuplicate ImportRowStatus = "duplicate_mac"
	ImportMissing   ImportRowStatus = "missing_field"
	ImportBadMAC    ImportRowStatus = "invalid_mac"
	ImportFailed    ImportRowStatus = "failed"
)

type ImportResult struct {
	RowIndex int             `json:"rowIndex"`
	Status   string          `json:"status"` // "imported" | "skipped" | "failed"
	Reason   string          `json:"reason,omitempty"`
	DeviceID string          `json:"deviceId,omitempty"`
}

type ImportSummary struct {
	Imported int            `json:"imported"`
	Skipped  int            `json:"skipped"`
	Failed   int            `json:"failed"`
	Results  []ImportResult `json:"results"`
}

// ─── Settings Objects ─────────────────────────────────────────────────────────

type DeviceMetric struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"deviceId"`
	MetricType string    `json:"metricType"` // "cpu" | "memory"
	Value      float64   `json:"value"`
	RecordedAt time.Time `json:"recordedAt"`
}

type SystemSettings struct {
	Channels []struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Type           string `json:"type"`
		Connected      bool   `json:"connected"`
		HandleOrNumber string `json:"handleOrNumber"`
		LastSync       string `json:"lastSync"`
	} `json:"channels"`
	RateLimitMaxMsgPerMin int `json:"rateLimitMaxMsgPerMin"`
	Thresholds            []struct {
		Type                DeviceType `json:"type"`
		ConsecutiveFailures int        `json:"consecutiveFailures"`
	} `json:"thresholds"`
	Polling struct {
		IntervalSeconds      int `json:"intervalSeconds"`
		ConcurrencyBatchSize int `json:"concurrencyBatchSize"`
		DebounceSeconds      int `json:"debounceSeconds"`
	} `json:"polling"`
}

// ─── User ─────────────────────────────────────────────────────────────────────

type User struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Role        Role     `json:"role"`
	Status      string   `json:"status"`
	AvatarURL   string   `json:"avatarUrl"`
	LastActive  string   `json:"lastActive"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"isActive"`
	MFAEnabled  bool     `json:"mfaEnabled"`
	MFASecret   string   `json:"-"`
}

type UpdateProfileRequest struct {
	Username        string `json:"username,omitempty"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	AvatarURL       string `json:"avatarUrl"`
	CurrentPassword string `json:"currentPassword,omitempty"`
	NewPassword     string `json:"newPassword,omitempty"`
}

// ─── WebSocket ────────────────────────────────────────────────────────────────

type WSMessage struct {
	Type        string       `json:"type"`
	DeviceID    string       `json:"deviceId,omitempty"`
	DeviceName  string       `json:"deviceName,omitempty"`
	Status      DeviceStatus `json:"status,omitempty"`
	IP          string       `json:"ip,omitempty"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Severity    string       `json:"severity,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
	LatencyMs   int64        `json:"latencyMs,omitempty"`
}

type IncidentEvent struct {
	ID         string    `json:"id"`
	IncidentID string    `json:"incidentId"`
	EventType  string    `json:"eventType"` // check_failed, threshold_reached, channel_attempt, channel_failed, channel_fallback, channel_delivered, resolved
	Channel    string    `json:"channel,omitempty"`
	Detail     string    `json:"detail"`
	OccurredAt time.Time `json:"occurredAt"`
}
