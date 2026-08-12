// Package repository provides an in-memory implementation of all repository
// interfaces. This is suitable for local development and the existing demo
// flow. Replace with the SQLite or PostgreSQL implementation in production
// by satisfying the same interfaces in a new file.
package repository

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"sanoc/backend/internal/domain"
)

// ─── In-Memory Device Repository ─────────────────────────────────────────────

type MemoryDeviceRepository struct {
	mu      sync.RWMutex
	devices []domain.Device
	nextID  int
}

func NewMemoryDeviceRepository(seed []domain.Device) *MemoryDeviceRepository {
	return &MemoryDeviceRepository{devices: seed, nextID: len(seed) + 1}
}

func (r *MemoryDeviceRepository) GetAll(typeFilter, statusFilter, search string) ([]domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []domain.Device
	for _, d := range r.devices {
		if typeFilter != "" && typeFilter != "All" && string(d.Type) != typeFilter {
			continue
		}
		if statusFilter != "" && statusFilter != "All" && string(d.Status) != statusFilter {
			continue
		}
		if search != "" {
			q := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(d.Name), q) &&
				!strings.Contains(strings.ToLower(d.IP), q) &&
				!strings.Contains(strings.ToLower(d.MAC), q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *MemoryDeviceRepository) GetByID(id string) (*domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.devices {
		if r.devices[i].ID == id {
			cp := r.devices[i]
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("device %s not found", id)
}

func (r *MemoryDeviceRepository) GetByMAC(mac string) (*domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mac = strings.ToLower(mac)
	for i := range r.devices {
		if strings.ToLower(r.devices[i].MAC) == mac {
			cp := r.devices[i]
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("device with MAC %s not found", mac)
}

func (r *MemoryDeviceRepository) GetDHCPDevices() ([]domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Device
	for _, d := range r.devices {
		if d.AddressingMode == domain.AddressingDHCP {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *MemoryDeviceRepository) Create(d *domain.Device) (*domain.Device, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d.ID = fmt.Sprintf("dev-%d", r.nextID)
	r.nextID++
	r.devices = append([]domain.Device{*d}, r.devices...)
	return d, nil
}

func (r *MemoryDeviceRepository) Update(d *domain.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.devices {
		if r.devices[i].ID == d.ID {
			r.devices[i] = *d
			return nil
		}
	}
	return fmt.Errorf("device %s not found", d.ID)
}

func (r *MemoryDeviceRepository) UpdateLastKnownIP(deviceID, newIP string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.devices {
		if r.devices[i].ID == deviceID {
			r.devices[i].IP = newIP
			return nil
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}

func (r *MemoryDeviceRepository) UpdateStatus(deviceID string, status domain.DeviceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.devices {
		if r.devices[i].ID == deviceID {
			r.devices[i].Status = status
			r.devices[i].CheckedSecondsAgo = 0
			r.devices[i].LastChecked = time.Now().Format("15:04:05")
			return nil
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}

func (r *MemoryDeviceRepository) UpdateSNMPSysName(deviceID, sysName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.devices {
		if r.devices[i].ID == deviceID {
			r.devices[i].SNMPSysName = sysName
			return nil
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}

func (r *MemoryDeviceRepository) UpdateSNMPMetadata(deviceID, sysName, sysDescr, sysUpTime, sysContact, sysLocation string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.devices {
		if r.devices[i].ID == deviceID {
			if sysName != "" {
				r.devices[i].SNMPSysName = sysName
			}
			if sysDescr != "" {
				r.devices[i].SNMPSysDescr = sysDescr
			}
			if sysUpTime != "" {
				r.devices[i].SNMPSysUpTime = sysUpTime
			}
			if sysContact != "" {
				r.devices[i].SNMPSysContact = sysContact
			}
			if sysLocation != "" {
				r.devices[i].SNMPSysLocation = sysLocation
			}
			return nil
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}

func (r *MemoryDeviceRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.devices {
		if r.devices[i].ID == id {
			r.devices = append(r.devices[:i], r.devices[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("device %s not found", id)
}

func (r *MemoryDeviceRepository) ExistsByMAC(mac string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mac = strings.ToLower(mac)
	for _, d := range r.devices {
		if strings.ToLower(d.MAC) == mac {
			return true, nil
		}
	}
	return false, nil
}

// ─── In-Memory Status Log Repository ─────────────────────────────────────────

type MemoryStatusLogRepository struct {
	mu   sync.RWMutex
	logs []domain.DeviceStatusLog
}

func NewMemoryStatusLogRepository() *MemoryStatusLogRepository {
	return &MemoryStatusLogRepository{}
}

func (r *MemoryStatusLogRepository) Append(log *domain.DeviceStatusLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, *log)
	return nil
}

func (r *MemoryStatusLogRepository) GetLogs() []domain.DeviceStatusLog {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.DeviceStatusLog, len(r.logs))
	copy(res, r.logs)
	return res
}

func (r *MemoryStatusLogRepository) CountInWindow(deviceID string, status domain.DeviceStatus, from, to time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, l := range r.logs {
		if l.DeviceID == deviceID && l.Status == status &&
			(l.Timestamp.Equal(from) || l.Timestamp.After(from)) &&
			(l.Timestamp.Before(to) || l.Timestamp.Equal(to)) {
			count++
		}
	}
	return count, nil
}

func (r *MemoryStatusLogRepository) SumDowntimeMinutes(deviceID string, from, to time.Time) (int, error) {
	// Simplified: count DOWN events * estimated avg downtime per event
	count, err := r.CountInWindow(deviceID, domain.StatusDOWN, from, to)
	return count * 8, err // stub: 8 min average per event
}

func (r *MemoryStatusLogRepository) GetFlapDevices(threshold int, from, to time.Time) ([]domain.FlapReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[string]int)
	for _, l := range r.logs {
		if l.Status == domain.StatusDOWN &&
			(l.Timestamp.Equal(from) || l.Timestamp.After(from)) &&
			(l.Timestamp.Before(to) || l.Timestamp.Equal(to)) {
			counts[l.DeviceID]++
		}
	}

	var reports []domain.FlapReport
	for deviceID, count := range counts {
		if count >= threshold {
			reports = append(reports, domain.FlapReport{
				DeviceID:             deviceID,
				DownCount7d:          count,
				TotalDowntimeMinutes: count * 8,
			})
		}
	}
	return reports, nil
}

func (r *MemoryStatusLogRepository) GetHistoryGrouped(deviceID string, from, to time.Time, truncateUnit string) ([]domain.StatusHistoryPoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	grouped := make(map[string]*domain.StatusHistoryPoint)
	var keys []string

	for _, l := range r.logs {
		if (deviceID == "" || l.DeviceID == deviceID) &&
			(l.Timestamp.Equal(from) || l.Timestamp.After(from)) &&
			(l.Timestamp.Before(to) || l.Timestamp.Equal(to)) {

			var key string
			if truncateUnit == "hour" {
				key = l.Timestamp.Format("15:00")
			} else {
				key = l.Timestamp.Format("Jan 02")
			}

			if _, exists := grouped[key]; !exists {
				grouped[key] = &domain.StatusHistoryPoint{Date: key}
				keys = append(keys, key)
			}
			if l.Status == domain.StatusUP {
				grouped[key].UpCount++
			} else {
				grouped[key].DownCount++
			}
		}
	}

	var res []domain.StatusHistoryPoint
	for _, k := range keys {
		res = append(res, *grouped[k])
	}
	return res, nil
}

func (r *MemoryStatusLogRepository) GetDowntimeReport(from, to time.Time) ([]domain.DeviceDowntimeSummary, error) {
	return []domain.DeviceDowntimeSummary{}, nil
}

// ─── In-Memory Notification Log Repository ───────────────────────────────────

type MemoryNotificationLogRepository struct {
	mu   sync.RWMutex
	logs []domain.NotificationLogRow
}

func NewMemoryNotificationLogRepository() *MemoryNotificationLogRepository {
	return &MemoryNotificationLogRepository{}
}

func (r *MemoryNotificationLogRepository) Append(row *domain.NotificationLogRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, *row)
	return nil
}

func (r *MemoryNotificationLogRepository) GetByIncidentID(incidentID string) ([]domain.NotificationLogRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// In a real DB this would join via incident_id column
	return r.logs, nil
}
