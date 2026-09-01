package poller

import (
	"context"
	"database/sql"
	"encoding/binary"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"sanoc/backend/internal/config"
	"sanoc/backend/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

type DHCPSyncWorker struct {
	cfg          *config.Config
	deviceRepo   repository.DeviceRepository
	stopChan     chan struct{}
	leaseCacheMu sync.RWMutex
	leaseCache   map[string]string // normalized hex MAC -> IP
}

func NewDHCPSyncWorker(cfg *config.Config, deviceRepo repository.DeviceRepository) *DHCPSyncWorker {
	return &DHCPSyncWorker{
		cfg:        cfg,
		deviceRepo: deviceRepo,
		stopChan:   make(chan struct{}),
		leaseCache: make(map[string]string),
	}
}

// GetCachedLease returns the active Kea lease IP for a normalized MAC, if known.
func (w *DHCPSyncWorker) GetCachedLease(normalizedMAC string) (string, bool) {
	w.leaseCacheMu.RLock()
	defer w.leaseCacheMu.RUnlock()
	if w.leaseCache == nil {
		return "", false
	}
	norm := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(normalizedMAC, ":", ""), "-", ""))
	ip, ok := w.leaseCache[norm]
	return ip, ok
}

// GetAllCachedLeases returns a snapshot copy of all active Kea leases.
func (w *DHCPSyncWorker) GetAllCachedLeases() map[string]string {
	w.leaseCacheMu.RLock()
	defer w.leaseCacheMu.RUnlock()
	if w.leaseCache == nil {
		return nil
	}
	cp := make(map[string]string, len(w.leaseCache))
	for k, v := range w.leaseCache {
		cp[k] = v
	}
	return cp
}

func (w *DHCPSyncWorker) Start() {
	if w.cfg.KeaDBHost == "" {
		log.Println("[SANOC-DHCP-SYNC] Kea DHCP Database Host is not configured. Auto-sync disabled.")
		return
	}

	interval, err := time.ParseDuration(w.cfg.KeaSyncInterval)
	if err != nil || interval <= 0 {
		interval = 5 * time.Minute
		log.Printf("[SANOC-DHCP-SYNC] Invalid sync interval %s, defaulting to 5m", w.cfg.KeaSyncInterval)
	}

	log.Printf("[SANOC-DHCP-SYNC] Starting worker. Syncing every %v with Kea DB at %s:%s", interval, w.cfg.KeaDBHost, w.cfg.KeaDBPort)

	go func() {
		// Run first sync immediately
		w.SyncNow()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.SyncNow()
			case <-w.stopChan:
				log.Println("[SANOC-DHCP-SYNC] Worker stopped.")
				return
			}
		}
	}()
}

func (w *DHCPSyncWorker) Stop() {
	select {
	case <-w.stopChan:
		// already closed
	default:
		close(w.stopChan)
	}
}

func (w *DHCPSyncWorker) SyncNow() {
	// 1. Connect to Kea Database
	db, err := sql.Open("mysql", w.cfg.KeaMySQLDSN())
	if err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Failed to open connection to Kea DB: %v", err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Cannot ping Kea DB: %v", err)
		return
	}

	// 2. Get SANOC Devices that use DHCP
	devices, err := w.deviceRepo.GetDHCPDevices()
	if err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Failed to fetch DHCP devices from local DB: %v", err)
	}

	// Create a map for O(1) lookup by normalized MAC
	deviceMap := make(map[string]*repositoryDevicePointer)
	for i := range devices {
		normalizedMAC := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(devices[i].MAC), ":", ""), "-", "")
		deviceMap[normalizedMAC] = &repositoryDevicePointer{
			ID:   devices[i].ID,
			IP:   devices[i].IP,
			Name: devices[i].Name,
		}
	}

	// 3. Query Active Leases — hex(hwaddr) converts VARBINARY to hex string
	query := `SELECT hex(hwaddr), address FROM lease4 WHERE state = 0`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Failed to query lease4 table: %v", err)
		return
	}
	defer rows.Close()

	leaseCount := 0
	updateCount := 0
	newLeaseCache := make(map[string]string)

	for rows.Next() {
		var hwaddrStr string
		var addressInt int64
		if err := rows.Scan(&hwaddrStr, &addressInt); err != nil {
			log.Printf("[SANOC-DHCP-SYNC] Scan error: %v", err)
			continue
		}
		leaseCount++

		macStr := strings.ToLower(hwaddrStr)
		// Convert integer address to dotted IP string
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, uint32(addressInt))
		newIP := ip.String()

		newLeaseCache[macStr] = newIP

		if dev, exists := deviceMap[macStr]; exists {
			if dev.IP != newIP {
				log.Printf("[SANOC-DHCP-SYNC] IP change (KEA_DHCP): %s (MAC: %s) %s -> %s", dev.Name, macStr, dev.IP, newIP)
				if err := w.deviceRepo.UpdateLastKnownIPWithSource(dev.ID, newIP, "KEA_DHCP"); err != nil {
					log.Printf("[SANOC-DHCP-SYNC] Failed to update %s: %v", dev.Name, err)
				} else {
					updateCount++
				}
			}
		}
	}

	w.leaseCacheMu.Lock()
	w.leaseCache = newLeaseCache
	w.leaseCacheMu.Unlock()

	if err := rows.Err(); err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Row iteration error: %v", err)
	}

	log.Printf("[SANOC-DHCP-SYNC] Sync done. Leases: %d, SANOC DHCP devices: %d, Updated: %d", leaseCount, len(devices), updateCount)
}

// repositoryDevicePointer holds minimal device info for DHCP sync lookups.
type repositoryDevicePointer struct {
	ID   string
	IP   string
	Name string
}

