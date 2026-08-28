package poller

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"log"
	"net"
	"strings"
	"time"

	"sanoc/backend/internal/config"
	"sanoc/backend/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

type DHCPSyncWorker struct {
	cfg        *config.Config
	deviceRepo repository.DeviceRepository
	stopChan   chan struct{}
}

func NewDHCPSyncWorker(cfg *config.Config, deviceRepo repository.DeviceRepository) *DHCPSyncWorker {
	return &DHCPSyncWorker{
		cfg:        cfg,
		deviceRepo: deviceRepo,
		stopChan:   make(chan struct{}),
	}
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
		w.runSync()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.runSync()
			case <-w.stopChan:
				log.Println("[SANOC-DHCP-SYNC] Worker stopped.")
				return
			}
		}
	}()
}

func (w *DHCPSyncWorker) Stop() {
	close(w.stopChan)
}

func (w *DHCPSyncWorker) runSync() {
	// 1. Get SANOC Devices that use DHCP
	devices, err := w.deviceRepo.GetDHCPDevices()
	if err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Failed to fetch DHCP devices from local DB: %v", err)
		return
	}
	if len(devices) == 0 {
		return // No DHCP devices to sync
	}

	// Create a map for O(1) lookup
	deviceMap := make(map[string]*repositoryDevicePointer)
	for i := range devices {
		// Normalize MAC address to lowercase without colons for easy matching
		normalizedMAC := strings.ReplaceAll(strings.ToLower(devices[i].MAC), ":", "")
		deviceMap[normalizedMAC] = &repositoryDevicePointer{
			ID:    devices[i].ID,
			IP:    devices[i].IP,
			Name:  devices[i].Name,
		}
	}

	// 2. Connect to Kea Database
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

	// 3. Query Active Leases
	// `hwaddr` is bytea, `address` is bigint in Kea PostgreSQL schema
	query := `SELECT hwaddr, address FROM lease4 WHERE state = 0` // 0 = default (active) state in Kea
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Failed to query lease4 table: %v", err)
		return
	}
	defer rows.Close()

	updateCount := 0

	for rows.Next() {
		var hwaddr []byte
		var addressInt int64
		if err := rows.Scan(&hwaddr, &addressInt); err != nil {
			continue
		}

		// Convert hardware address byte array to hex string without colons
		macStr := hex.EncodeToString(hwaddr)

		// Check if we have this MAC in our SANOC devices
		if dev, exists := deviceMap[macStr]; exists {
			// Convert unsigned 32-bit int to IP string
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, uint32(addressInt))
			newIP := ip.String()

			// Update if IP has changed
			if dev.IP != newIP {
				log.Printf("[SANOC-DHCP-SYNC] IP Change detected for %s (MAC: %s) -> Old: %s, New: %s", dev.Name, macStr, dev.IP, newIP)
				err := w.deviceRepo.UpdateLastKnownIP(dev.ID, newIP)
				if err != nil {
					log.Printf("[SANOC-DHCP-SYNC] Failed to update device %s IP in local DB: %v", dev.ID, err)
				} else {
					updateCount++
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("[SANOC-DHCP-SYNC] Error during lease4 row iteration: %v", err)
	}

	if updateCount > 0 {
		log.Printf("[SANOC-DHCP-SYNC] Sync complete. Updated %d device IPs.", updateCount)
	}
}

// Struct to hold minimum needed device info internally
type repositoryDevicePointer struct {
	ID   string
	IP   string
	Name string
}
