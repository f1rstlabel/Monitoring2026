// Package poller implements the rate-limited, batch-scheduled ICMP polling engine.
// Worker pool caps concurrent requests; batch scheduler distributes 414 devices
// evenly across the polling interval; debounce tracker prevents flapping alerts.
package poller

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/notifier"
	"govmonitor-it/backend/internal/repository"
	"govmonitor-it/backend/internal/ws"

	"github.com/gosnmp/gosnmp"
)

var wibLocation = time.FixedZone("WIB", 7*3600)

func nowWIB() time.Time {
	return time.Now().In(wibLocation)
}

// DebounceState tracks consecutive failure count per device.
type DebounceState struct {
	mu          sync.Mutex
	consecutive map[string]int // deviceID → consecutive fail count
}

func newDebounceState() *DebounceState {
	return &DebounceState{consecutive: make(map[string]int)}
}

// RecordFail increments the fail count and returns the new value.
func (d *DebounceState) RecordFail(deviceID string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.consecutive[deviceID]++
	return d.consecutive[deviceID]
}

// RecordSuccess resets the fail count.
func (d *DebounceState) RecordSuccess(deviceID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.consecutive[deviceID] = 0
}

// ─── Engine ──────────────────────────────────────────────────────────────────

// Engine is the polling coordinator. It batches devices, distributes work
// across a goroutine worker pool, and drives status change callbacks.
type Engine struct {
	hub        *ws.Hub
	pipeline   *notifier.Pipeline
	deviceRepo   repository.DeviceRepository
	statusRepo   repository.StatusLogRepository
	settingsRepo *repository.SettingsRepository // optional: read live settings each cycle
	metricRepo   repository.DeviceMetricRepository
	incidentRepo repository.IncidentRepository
	notifyQueue  *notifier.NotifyQueue

	cfgMu       sync.RWMutex
	intervalSec int
	concurrency int // max simultaneous ICMP/SNMP workers
	debounceSec int
	thresholds  map[domain.DeviceType]int

	debounce *DebounceState

	stateMu       sync.Mutex
	notifiedState map[string]domain.DeviceStatus

	// reloadChan signals the main loop to rebuild its ticker with a new interval
	reloadChan chan struct{}
	stopChan   chan struct{}
}

// EngineConfig carries tunable parameters sourced from Settings UI.
type EngineConfig struct {
	IntervalSeconds      int
	ConcurrencyBatchSize int
	DebounceSeconds      int
	Thresholds           map[domain.DeviceType]int
}

func DefaultConfig() EngineConfig {
	return EngineConfig{
		IntervalSeconds:      15,
		ConcurrencyBatchSize: 50,
		DebounceSeconds:      60,
		Thresholds: map[domain.DeviceType]int{
			domain.AccessPoint: 3,
			domain.Switch:      2,
			domain.Router:      1,
			domain.SmartPower:  3,
			domain.CCTV:        3,
			domain.NVR:         3,
		},
	}
}

func NewEngine(hub *ws.Hub, deviceRepo repository.DeviceRepository, statusRepo repository.StatusLogRepository, cfg EngineConfig) *Engine {
	return &Engine{
		hub:           hub,
		deviceRepo:    deviceRepo,
		statusRepo:    statusRepo,
		intervalSec:   cfg.IntervalSeconds,
		concurrency:   cfg.ConcurrencyBatchSize,
		debounceSec:   cfg.DebounceSeconds,
		debounce:      newDebounceState(),
		notifiedState: make(map[string]domain.DeviceStatus),
		thresholds:    cfg.Thresholds,
		reloadChan:    make(chan struct{}, 1),
		stopChan:      make(chan struct{}),
	}
}

func (e *Engine) SetPipeline(p *notifier.Pipeline) {
	e.pipeline = p
}

func (e *Engine) SetNotifyQueue(nq *notifier.NotifyQueue) {
	e.notifyQueue = nq
}

func (e *Engine) SetSettingsRepo(repo *repository.SettingsRepository) {
	e.settingsRepo = repo
}

func (e *Engine) SetMetricRepo(repo repository.DeviceMetricRepository) {
	e.metricRepo = repo
}

func (e *Engine) SetIncidentRepo(repo repository.IncidentRepository) {
	e.incidentRepo = repo
}

// UpdateConfig atomically replaces the engine's runtime config.
// Called by settings handlers immediately after saving to DB so changes
// take effect within the current or next poll cycle — no restart required.
func (e *Engine) UpdateConfig(cfg EngineConfig) {
	e.cfgMu.Lock()
	defer e.cfgMu.Unlock()

	intervalChanged := cfg.IntervalSeconds > 0 && cfg.IntervalSeconds != e.intervalSec

	if cfg.IntervalSeconds > 0 {
		e.intervalSec = cfg.IntervalSeconds
	}
	if cfg.ConcurrencyBatchSize > 0 {
		e.concurrency = cfg.ConcurrencyBatchSize
	}
	if cfg.Thresholds != nil {
		e.thresholds = cfg.Thresholds
	}

	log.Printf("[PollerEngine] Config updated live: interval=%ds concurrency=%d", e.intervalSec, e.concurrency)

	// Signal main loop to rebuild ticker if interval changed
	if intervalChanged {
		select {
		case e.reloadChan <- struct{}{}:
		default:
		}
	}
}

func (e *Engine) getIntervalSec() int {
	e.cfgMu.RLock()
	defer e.cfgMu.RUnlock()
	return e.intervalSec
}

func (e *Engine) getConcurrency() int {
	e.cfgMu.RLock()
	defer e.cfgMu.RUnlock()
	return e.concurrency
}

func (e *Engine) getThresholds() map[domain.DeviceType]int {
	e.cfgMu.RLock()
	defer e.cfgMu.RUnlock()
	return e.thresholds
}

// Start launches the background polling ticker. Non-blocking.
func (e *Engine) Start() {
	go func() {
		currentInterval := e.getIntervalSec()
		ticker := time.NewTicker(time.Duration(currentInterval) * time.Second)
		defer ticker.Stop()

		// Initial device count for "poller starting" log
		initialDevices, err := e.deviceRepo.GetAll("", "", "")
		numInitialDevices := 0
		if err != nil {
			log.Printf("[PollerEngine] Failed to fetch initial devices for log: %v", err)
		} else {
			numInitialDevices = len(initialDevices)
		}
		log.Printf("[PollerEngine] Poller starting: %d devices loaded, interval=%ds", numInitialDevices, currentInterval)

		// Run an initial poll cycle immediately on boot
		e.runPollCycle()

		for {
			select {
			case <-ticker.C:
				e.runPollCycle()

			case <-e.reloadChan:
				// Interval changed — rebuild ticker
				newInterval := e.getIntervalSec()
				if newInterval != currentInterval && newInterval > 0 {
					ticker.Stop()
					currentInterval = newInterval
					ticker = time.NewTicker(time.Duration(currentInterval) * time.Second)
					log.Printf("[PollerEngine] Polling interval changed live to %ds", currentInterval)
				}

			case <-e.stopChan:
				log.Println("[PollerEngine] Stopped")
				return
			}
		}
	}()
}

// Stop signals the polling loop to exit.
func (e *Engine) Stop() {
	close(e.stopChan)
}

// TriggerPollNow forces an immediate out-of-schedule ICMP polling cycle.
func (e *Engine) TriggerPollNow() (int, time.Time) {
	log.Println("[PollerEngine] Manual refresh triggered via API")
	go e.runPollCycle()
	devices, _ := e.deviceRepo.GetAll("", "", "")
	return len(devices), nowWIB()
}

// runPollCycle fetches all devices, resolves DHCP IPs, then polls in batches.
// At the start of each cycle it optionally reloads settings from DB to pick up
// any changes made in the Settings UI since the last cycle.
func (e *Engine) runPollCycle() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PollerEngine] CRITICAL: Recovered from panic in runPollCycle: %v", r)
		}
	}()

	startTime := time.Now()
	upCount := 0
	downCount := 0
	checkedCount := 0

	// Reload settings from DB if we have a settings repo (no-op if nil)
	if e.settingsRepo != nil {
		var s struct {
			Polling struct {
				IntervalSeconds      int `json:"intervalSeconds"`
				ConcurrencyBatchSize int `json:"concurrencyBatchSize"`
			} `json:"polling"`
			Thresholds []struct {
				Type                string `json:"type"`
				ConsecutiveFailures int    `json:"consecutiveFailures"`
			} `json:"thresholds"`
		}
		if err := e.settingsRepo.GetJSON("system_settings", &s); err == nil {
			thresholds := make(map[domain.DeviceType]int)
			for _, t := range s.Thresholds {
				thresholds[domain.DeviceType(t.Type)] = t.ConsecutiveFailures
			}
			e.UpdateConfig(EngineConfig{
				IntervalSeconds:      s.Polling.IntervalSeconds,
				ConcurrencyBatchSize: s.Polling.ConcurrencyBatchSize,
				Thresholds:           thresholds,
			})
		}
	}

	devices, err := e.deviceRepo.GetAll("", "", "")
	if err != nil {
		log.Printf("[PollerEngine] Failed to fetch devices: %v", err)
		return
	}
	if len(devices) == 0 {
		log.Printf("[PollerEngine] Poller cycle complete: 0 devices checked")
		return
	}

	// Worker pool channel pattern
	workChan := make(chan domain.Device, len(devices))
	resultChan := make(chan bool, len(devices)) // Channel to collect results
	var wg sync.WaitGroup

	numWorkers := min(e.getConcurrency(), len(devices))
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PollerEngine] Recovered from panic in worker: %v", r)
				}
			}()
			for dev := range workChan {
				e.pollDevice(dev)
				updatedDev, err := e.deviceRepo.GetByID(dev.ID) // Fetch updated status
				if err != nil || updatedDev == nil {
					log.Printf("[PollerEngine] Could not retrieve updated status for device %s", dev.ID)
					resultChan <- false // Assume failed if status not retrieved
				} else {
					resultChan <- (updatedDev.Status == domain.StatusUP)
				}
			}
		}()
	}

	for _, dev := range devices {
		workChan <- dev
	}
	close(workChan)
	wg.Wait()
	close(resultChan) // Close result channel after all workers are done

	// Collect results
	for isUp := range resultChan {
		checkedCount++
		if isUp {
			upCount++
		} else {
			downCount++
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("[PollerEngine] Poller cycle complete: %d checked, %d up, %d down, took %s",
		checkedCount, upCount, downCount, elapsed.Round(time.Millisecond))

	if e.hub != nil {
		e.hub.BroadcastMessage(domain.WSMessage{
			Type:        "DASHBOARD_UPDATE",
			Title:       "Poller Cycle Complete",
			Description: fmt.Sprintf("Checked %d devices (%d UP, %d DOWN) in %s", checkedCount, upCount, downCount, elapsed.Round(time.Millisecond)),
			Severity:    "info",
			Timestamp:   nowWIB(),
		})
		e.hub.BroadcastMessage(domain.WSMessage{
			Type:        "LIVE_FEED",
			Title:       "Poller Cycle Summary",
			Description: fmt.Sprintf("Completed poll cycle: %d checked, %d UP, %d DOWN in %s", checkedCount, upCount, downCount, elapsed.Round(time.Millisecond)),
			Severity:    "info",
			Timestamp:   nowWIB(),
		})
	}
}

func (e *Engine) pollDevice(dev domain.Device) {
	targetIP := dev.IP

	// If DHCP with MAC address, attempt resolution
	if dev.AddressingMode == domain.AddressingDHCP && dev.MAC != "" {
		resolved := ResolveIPFromARPWarm(dev.MAC, dev.IP)
		if resolved != "" && resolved != dev.IP {
			log.Printf("[PollerEngine] RESOLVED DHCP IP for %s (%s) -> %s", dev.Name, dev.MAC, resolved)
			_ = e.deviceRepo.UpdateLastKnownIP(dev.ID, resolved)
			e.hub.BroadcastMessage(domain.WSMessage{
				Type:        "IP_CHANGED",
				DeviceID:    dev.ID,
				DeviceName:  dev.Name,
				IP:          resolved,
				Description: fmt.Sprintf("%s DHCP lease renewed — IP changed from %s to %s", dev.Name, dev.IP, resolved),
				Severity:    "info",
				Timestamp:   nowWIB(),
			})
			targetIP = resolved
		} else if resolved != "" {
			targetIP = resolved
		}
	}

	// Skip poll if no IP to check (DHCP device with no known IP yet)
	if targetIP == "" {
		_ = e.deviceRepo.UpdateStatus(dev.ID, domain.StatusUnresolvedDHCP)
		log.Printf("[PollerEngine] Skipping %q — no IP resolved (DHCP awaiting lease)", dev.Name)
		log.Printf("[PollerEngine] Skipping poll for %s — no IP resolved (DHCP, awaiting first lease)", dev.Name)
		return
	}

	log.Printf("[PollerEngine] Polling device %q (%s) via ICMP/SNMP...", dev.Name, targetIP)

	var (
		snmpUp     bool
		snmpErrMsg string
		icmpUp     bool
		latencyMs  int64
		wg         sync.WaitGroup
	)

	// ─── Concurrent Non-Blocking SNMP & ICMP Checks ─────────────────────────────
	if dev.SNMPEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("[PollerEngine] Attempting SNMP check for %s (%s)", dev.Name, targetIP)
			snmpResult := e.snmpCheck(dev, targetIP)
			if snmpResult.Up {
				log.Printf("[PollerEngine] SNMP UP for %s (%s)", dev.Name, targetIP)
				snmpUp = true
			} else {
				snmpErrMsg = snmpResult.Error
				log.Printf("[PollerEngine] SNMP DOWN for %s (%s): %s", dev.Name, targetIP, snmpResult.Error)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		icmpUp, latencyMs = e.icmpCheckWithLatency(targetIP)
	}()

	wg.Wait()

	up := icmpUp

	// Always record latency metric for per-poll time-series (0 ms for DOWN/timeout)
	if e.metricRepo != nil {
		latVal := float64(latencyMs)
		if !up {
			latVal = 0
		}
		_ = e.metricRepo.SaveMetric(&domain.DeviceMetric{
			DeviceID:   dev.ID,
			MetricType: "latency",
			Value:      latVal,
			RecordedAt: time.Now(),
		})
	}

	snmpStatus := "DISABLED"
	if dev.SNMPEnabled {
		if snmpUp {
			snmpStatus = "OK"
		} else {
			snmpStatus = fmt.Sprintf("FAIL (%s)", snmpErrMsg)
		}
	}

	// High-frequency Live Feed broadcast per device per poll cycle
	if e.hub != nil {
		feedStatusStr := "UP"
		feedSeverity := "info"
		icmpStr := fmt.Sprintf("ok (%dms)", latencyMs)
		if !up {
			feedStatusStr = "DOWN"
			feedSeverity = "critical"
			icmpStr = "fail (timeout)"
		}
		e.hub.BroadcastMessage(domain.WSMessage{
			Type:        "LIVE_FEED",
			DeviceID:    dev.ID,
			DeviceName:  dev.Name,
			Status:      domain.DeviceStatus(feedStatusStr),
			IP:          targetIP,
			Title:       fmt.Sprintf("%s — %s", dev.Name, feedStatusStr),
			Description: fmt.Sprintf("IP: %s | ICMP: %s | SNMP: %s", targetIP, icmpStr, snmpStatus),
			Severity:    feedSeverity,
			Timestamp:   nowWIB(),
			LatencyMs:   latencyMs,
		})
	}

	thresholds := e.getThresholds()

	if up {
		log.Printf("[PollerEngine] Device %q (%s) -> Status: UP | ICMP: ok (%dms) | SNMP: %s", dev.Name, targetIP, latencyMs, snmpStatus)
		prevStatus := dev.Status
		notified := e.getNotifiedState(dev.ID)
		e.debounce.RecordSuccess(dev.ID)
		if (prevStatus == domain.StatusDOWN || notified == domain.StatusDOWN) && notified != domain.StatusUP {
			e.setNotifiedState(dev.ID, domain.StatusUP)
			tNow := nowWIB()
			_ = e.deviceRepo.UpdateStatus(dev.ID, domain.StatusUP)
			_ = e.statusRepo.Append(&domain.DeviceStatusLog{
				ID:        fmt.Sprintf("log-%d", time.Now().UnixNano()),
				DeviceID:  dev.ID,
				Status:    domain.StatusUP,
				Timestamp: tNow,
			})
			if e.incidentRepo != nil {
				_ = e.incidentRepo.ResolveActiveByDeviceID(dev.ID, tNow)
			}
			e.hub.BroadcastMessage(domain.WSMessage{
				Type:        "STATUS_CHANGE",
				DeviceID:    dev.ID,
				DeviceName:  dev.Name,
				Status:      domain.StatusUP,
				IP:          targetIP,
				Title:       fmt.Sprintf("🟢 %s recovered", dev.Name),
				Description: fmt.Sprintf("%s (%s) is back online", dev.Name, targetIP),
				Severity:    "info",
				Timestamp:   tNow,
				LatencyMs:   latencyMs,
			})

			log.Printf("[Notifier] Device %s (%s) recovered to UP, queueing notification", dev.Name, targetIP)
			if e.notifyQueue != nil {
				e.notifyQueue.Enqueue(notifier.IncidentEvent{
					DeviceID:   dev.ID,
					DeviceName: dev.Name,
					DeviceType: string(dev.Type),
					IP:         targetIP,
					Location:   dev.Location,
					EventType:  notifier.EventRecovered,
					Timestamp:  tNow,
				})
			} else if e.pipeline != nil {
				msg := fmt.Sprintf("🟢 <b>RECOVERED — Device UP</b>\nDevice: <b>%s</b> (%s)\nIP: <code>%s</code>\nTime: %s", dev.Name, dev.Type, targetIP, tNow.Format("15:04:05 WIB"))
				go e.pipeline.Send(context.Background(), msg)
			}
		}
	} else {
		log.Printf("[PollerEngine] Device %q (%s) -> Status: DOWN | ICMP: fail (timeout) | SNMP: %s", dev.Name, targetIP, snmpStatus)

		// Failure — apply custom or default threshold and debounce
		threshold := thresholds[dev.Type]
		if dev.UseCustomThreshold && dev.CustomFailureThreshold != nil && *dev.CustomFailureThreshold > 0 {
			threshold = *dev.CustomFailureThreshold
		} else if dev.FailureThreshold > 0 {
			threshold = dev.FailureThreshold
		}
		if threshold <= 0 {
			threshold = 3
		}

		failCount := e.debounce.RecordFail(dev.ID)
		log.Printf("[PollerEngine] %s (%s) fail #%d/%d", dev.Name, targetIP, failCount, threshold)

		// Declare DOWN if threshold reached and not yet notified DOWN in memory
		if failCount >= threshold && (dev.Status == domain.StatusUP || e.getNotifiedState(dev.ID) != domain.StatusDOWN) {
			e.setNotifiedState(dev.ID, domain.StatusDOWN)

			// Threshold reached — declare DOWN
			tNow := nowWIB()
			_ = e.deviceRepo.UpdateStatus(dev.ID, domain.StatusDOWN)
			_ = e.statusRepo.Append(&domain.DeviceStatusLog{
				ID:        fmt.Sprintf("log-%d", time.Now().UnixNano()),
				DeviceID:  dev.ID,
				Status:    domain.StatusDOWN,
				Timestamp: tNow,
			})
			if e.incidentRepo != nil {
				_, _ = e.incidentRepo.Create(&domain.Incident{
					ID:         fmt.Sprintf("INC-%d", time.Now().UnixNano()/1e6),
					DeviceID:   dev.ID,
					PacketLoss: 100,
					LatencyMs:  0,
				})
			}
			e.hub.BroadcastMessage(domain.WSMessage{
				Type:        "STATUS_CHANGE",
				DeviceID:    dev.ID,
				DeviceName:  dev.Name,
				Status:      domain.StatusDOWN,
				IP:          targetIP,
				Title:       fmt.Sprintf("🔴 %s is DOWN", dev.Name),
				Description: fmt.Sprintf("%s (%s) failed %d consecutive polls — marking DOWN", dev.Name, targetIP, failCount),
				Severity:    "critical",
				Timestamp:   tNow,
				LatencyMs:   0,
			})

			log.Printf("[Incident] Device %s (%s) transitioned to DOWN, queueing notification", dev.Name, targetIP)
			if e.notifyQueue != nil {
				e.notifyQueue.Enqueue(notifier.IncidentEvent{
					DeviceID:   dev.ID,
					DeviceName: dev.Name,
					DeviceType: string(dev.Type),
					IP:         targetIP,
					Location:   dev.Location,
					EventType:  notifier.EventDown,
					Timestamp:  tNow,
				})
			} else if e.pipeline != nil {
				msg := fmt.Sprintf("🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>%s</b> (%s)\nIP: <code>%s</code>\nFailed Checks: %d\nTime: %s", dev.Name, dev.Type, targetIP, failCount, tNow.Format("15:04:05 WIB"))
				go e.pipeline.Send(context.Background(), msg)
			}
		}
	}
}

// SNMPCheckResult holds the result of an SNMP check.
type SNMPCheckResult struct {
	Up        bool
	SysUpTime string
	SysName   string
	Error     string
}

func (e *Engine) snmpCheck(dev domain.Device, ip string) SNMPCheckResult {
	if ip == "" {
		return SNMPCheckResult{Up: false, Error: "IP address is empty for SNMP check"}
	}

	community := dev.SNMPCommunity
	if community == "" {
		community = "public"
	}
	port := uint16(dev.SNMPPort)
	if port == 0 {
		port = 161
	}

	gosp := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      port,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   1500 * time.Millisecond,
		Retries:   0,
	}

	err := gosp.Connect()
	if err != nil {
		return SNMPCheckResult{Up: false, Error: fmt.Sprintf("SNMP connect failed: %v", err)}
	}
	defer gosp.Conn.Close()

	oids := []string{
		".1.3.6.1.2.1.1.3.0", // sysUpTime.0
		".1.3.6.1.2.1.1.5.0", // sysName.0
	}
	if dev.SNMPIfIndex > 0 {
		// ifOperStatus.dev.SNMPIfIndex
		oids = append(oids, fmt.Sprintf(".1.3.6.1.2.1.2.2.1.8.%d", dev.SNMPIfIndex))
	}

	resp, err := gosp.Get(oids)
	if err != nil {
		return SNMPCheckResult{Up: false, Error: fmt.Sprintf("SNMP GET failed: %v", err)}
	}

	result := SNMPCheckResult{Up: true}
	for _, pdu := range resp.Variables {
		switch pdu.Name {
		case ".1.3.6.1.2.1.1.3.0": // sysUpTime.0
			if pdu.Type == gosnmp.OctetString {
				result.SysUpTime = string(pdu.Value.([]byte))
			} else {
				result.SysUpTime = gosnmp.ToBigInt(pdu.Value).String()
			}
		case ".1.3.6.1.2.1.1.5.0": // sysName.0
			if pdu.Type == gosnmp.OctetString {
				result.SysName = string(pdu.Value.([]byte))
			}
		case fmt.Sprintf(".1.3.6.1.2.1.2.2.1.8.%d", dev.SNMPIfIndex): // ifOperStatus
			if pdu.Type == gosnmp.Integer {
				// ifOperStatus: 1=up, 2=down, 3=testing, 4=unknown, 5=dormant, 6=notPresent, 7=lowerLayerDown
				status := gosnmp.ToBigInt(pdu.Value).Int64()
				if status != 1 {
					result.Up = false
					result.Error = fmt.Sprintf("SNMP ifOperStatus is %d (not up)", status)
				}
			}
		}
	}

	// Fetch CPU & Memory metrics (HOST-RESOURCES-MIB) if metricRepo is available
	if e.metricRepo != nil {
		metricOids := []string{
			".1.3.6.1.2.1.25.3.3.1.2.1", // hrProcessorLoad.1
			".1.3.6.1.2.1.25.2.3.1.5.1", // hrStorageSize.1 (RAM)
			".1.3.6.1.2.1.25.2.3.1.6.1", // hrStorageUsed.1 (RAM)
		}
		if metricResp, mErr := gosp.Get(metricOids); mErr == nil && metricResp != nil {
			var cpuVal float64 = -1
			var memUsed float64 = -1
			var memTotal float64 = -1

			for _, pdu := range metricResp.Variables {
				switch pdu.Name {
				case ".1.3.6.1.2.1.25.3.3.1.2.1":
					cpuVal = float64(gosnmp.ToBigInt(pdu.Value).Int64())
				case ".1.3.6.1.2.1.25.2.3.1.5.1":
					memTotal = float64(gosnmp.ToBigInt(pdu.Value).Int64())
				case ".1.3.6.1.2.1.25.2.3.1.6.1":
					memUsed = float64(gosnmp.ToBigInt(pdu.Value).Int64())
				}
			}

			if cpuVal >= 0 {
				_ = e.metricRepo.SaveMetric(&domain.DeviceMetric{
					DeviceID:   dev.ID,
					MetricType: "cpu",
					Value:      cpuVal,
					RecordedAt: time.Now(),
				})
			}
			if memUsed >= 0 && memTotal > 0 {
				memPct := (memUsed / memTotal) * 100.0
				_ = e.metricRepo.SaveMetric(&domain.DeviceMetric{
					DeviceID:   dev.ID,
					MetricType: "memory",
					Value:      memPct,
					RecordedAt: time.Now(),
				})
			}
		}
	}

	return result
}


func (e *Engine) icmpCheck(ip string) bool {
	up, _ := e.icmpCheckWithLatency(ip)
	return up
}

// IsPingSuccessForTest exposes isPingSuccess for package tests.
func IsPingSuccessForTest(outStr string, err error) bool {
	return isPingSuccess(outStr, err)
}

// isPingSuccess parses the command output string and exit status to accurately determine
// if an ICMP reply was received, accounting for Windows ping.exe exit code 0 on timeout/unreachable,
// localized OS strings (English, Indonesian, etc.), and numeric patterns (Received = 0, Lost = 100%).
func isPingSuccess(outStr string, err error) bool {
	if err != nil {
		return false
	}
	outLower := strings.ToLower(outStr)

	// Explicit failure string checks (English, Indonesian, and common localized ping strings)
	failureSubstrings := []string{
		"request timed out",
		"timed out",
		"waktu minta habis",
		"destination host unreachable",
		"host unreachable",
		"tujuan tidak terjangkau",
		"tuan rumah tidak terjangkau",
		"general failure",
		"transmit failed",
		"could not find host",
		"unknown host",
		"100% loss",
		"100% packet loss",
		"100% hilang",
	}

	for _, sub := range failureSubstrings {
		if strings.Contains(outLower, sub) {
			return false
		}
	}

	// Regex check for Received = 0 / Diterima = 0 or 100% loss numeric patterns
	if matched, _ := regexp.MatchString(`(?i)(received|diterima)\s*[:=]\s*0\b`, outStr); matched {
		return false
	}
	if matched, _ := regexp.MatchString(`(?i)\b0\s*(received|diterima)\b`, outStr); matched {
		return false
	}
	if matched, _ := regexp.MatchString(`(?i)\(100%\s*(loss|packet loss|hilang)\)`, outStr); matched {
		return false
	}

	// Positive validation: must have indication of successful reply
	hasBytes := strings.Contains(outLower, "bytes=") || strings.Contains(outLower, "bytes from") || strings.Contains(outLower, "bita=")
	hasTTL := strings.Contains(outLower, "ttl=")
	hasTime := strings.Contains(outLower, "time=") || strings.Contains(outLower, "time<") || strings.Contains(outLower, "waktu=")

	if hasBytes || hasTTL || hasTime {
		return true
	}

	// Fallback numeric check for Received >= 1
	if matched, _ := regexp.MatchString(`(?i)(received|diterima)\s*[:=]\s*[1-9]\d*`, outStr); matched {
		return true
	}
	if matched, _ := regexp.MatchString(`(?i)\b[1-9]\d*\s*(received|diterima)\b`, outStr); matched {
		return true
	}

	return false
}

func (e *Engine) icmpCheckWithLatency(ip string) (bool, int64) {
	if ip == "" {
		return false, 0
	}

	start := time.Now()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	}

	outBytes, err := cmd.CombinedOutput()
	lat := time.Since(start).Milliseconds()
	outStr := string(outBytes)

	icmpUp := isPingSuccess(outStr, err)

	// Explicit raw debug logging as required by §1.2
	log.Printf("[PollerEngine] ICMP raw check for %s: received=%v, latency=%dms, err=%v, output=%q",
		ip, icmpUp, lat, err, strings.TrimSpace(outStr))

	if icmpUp {
		if lat <= 0 {
			lat = 1
		}
		return true, lat
	}

	// Fallback to TCP port check
	for _, port := range []string{"80", "443", "22", "53", "8080"} {
		conn, tcpErr := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 600*time.Millisecond)
		if tcpErr == nil {
			conn.Close()
			tcpLat := time.Since(start).Milliseconds()
			log.Printf("[PollerEngine] TCP fallback UP for %s on port %s (latency: %dms)", ip, port, tcpLat)
			return true, tcpLat
		}
	}

	log.Printf("[PollerEngine] ICMP and TCP check both DOWN for %s", ip)
	return false, 0
}

// resolveDHCPIP attempts to find the current IP for a given MAC address
// by parsing the local ARP table. Returns empty string if not found.
func (e *Engine) resolveDHCPIP(mac string) string {
	return ResolveIPFromARP(mac)
}

func (e *Engine) getNotifiedState(id string) domain.DeviceStatus {
	e.stateMu.Lock()
	defer e.stateMu.Unlock()
	return e.notifiedState[id]
}

func (e *Engine) setNotifiedState(id string, status domain.DeviceStatus) {
	e.stateMu.Lock()
	defer e.stateMu.Unlock()
	e.notifiedState[id] = status
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
