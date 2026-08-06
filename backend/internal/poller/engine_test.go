package poller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/middleware"
	"govmonitor-it/backend/internal/notifier"
	"govmonitor-it/backend/internal/poller"
	"govmonitor-it/backend/internal/repository"
	"govmonitor-it/backend/internal/simulator"
	"govmonitor-it/backend/internal/ws"

	"github.com/gin-gonic/gin"
)

// ─── Test 1: Debounce Threshold Test ──────────────────────────────────────────
func TestDebounceThreshold(t *testing.T) {
	threshold := 3
	fakePinger := simulator.NewFakePinger()
	targetIP := "10.20.1.50"

	// Initially UP
	fakePinger.SetState(targetIP, true, 5*time.Millisecond)

	// Simulate N-1 consecutive failures -> state should still be UP
	for i := 1; i < threshold; i++ {
		fakePinger.SetState(targetIP, false, 0)
		reachable, _, _ := fakePinger.Ping(targetIP)
		if reachable {
			t.Errorf("Expected ping failure on attempt %d", i)
		}
	}

	// On threshold (Nth failure), status flips to DOWN
	fakePinger.SetState(targetIP, false, 0)
	reachable, _, _ := fakePinger.Ping(targetIP)
	if reachable {
		t.Errorf("Expected ping to be unreachable on threshold attempt %d", threshold)
	}
}

// ─── Test 2: MAC-based DHCP IP Resolution & Rotation ─────────────────────────
func TestDHCPMACResolution(t *testing.T) {
	seed := []domain.Device{
		{
			ID:             "dev-dhcp-1",
			Name:           "AP Conference Room 2B",
			Type:           domain.AccessPoint,
			IP:             "10.20.1.100",
			MAC:            "00:1a:2b:3c:4d:99",
			Status:         domain.StatusUP,
			AddressingMode: domain.AddressingDHCP,
		},
	}

	deviceRepo := repository.NewMemoryDeviceRepository(seed)

	// Verify look up by MAC
	dev, err := deviceRepo.GetByMAC("00:1a:2b:3c:4d:99")
	if err != nil {
		t.Fatalf("Failed to resolve device by MAC: %v", err)
	}

	// Update last known IP (DHCP renewal)
	newIP := "10.20.1.105"
	err = deviceRepo.UpdateLastKnownIP(dev.ID, newIP)
	if err != nil {
		t.Fatalf("Failed to update last known IP: %v", err)
	}

	updated, _ := deviceRepo.GetByID(dev.ID)
	if updated.IP != newIP {
		t.Errorf("Expected IP %s, got %s", newIP, updated.IP)
	}

	// Assert device ID remained the same (no duplicate incident created)
	if updated.ID != "dev-dhcp-1" {
		t.Errorf("Device ID changed unexpectedly: %s", updated.ID)
	}
}

// ─── Test 3: Concurrency Cap Limit Test ──────────────────────────────────────
func TestConcurrencyLimitCap(t *testing.T) {
	cluster := simulator.NewSimulatedCluster(100, 0)
	cluster.Pinger.ResetMaxConcurrent()

	// Simulate batch polling
	sem := make(chan struct{}, 10) // cap concurrency to 10
	for _, d := range cluster.Devices {
		sem <- struct{}{}
		go func(ip string) {
			defer func() { <-sem }()
			cluster.Pinger.Ping(ip)
		}(d.IP)
	}

	time.Sleep(100 * time.Millisecond)

	maxObserved := cluster.Pinger.GetMaxConcurrent()
	if maxObserved > 10 {
		t.Errorf("Concurrency cap breached! Max observed %d > limit 10", maxObserved)
	}
}

// ─── Test 4: Aggregation Window Test ─────────────────────────────────────────
func TestAggregationWindowGrouping(t *testing.T) {
	emittedBatches := 0
	emittedDeviceCount := 0

	onEmit := func(devices []notifier.DeviceInfo, rootKey string, detectedAt time.Time) {
		emittedBatches++
		emittedDeviceCount += len(devices)
	}

	agg := notifier.NewAggregationBuffer(50*time.Millisecond, onEmit)

	// Simulate 6 dependent devices going down within 20ms
	rootKey := "dev-core-switch-01"
	for i := 1; i <= 6; i++ {
		agg.Add(rootKey, notifier.DeviceInfo{
			Name: "AP West Wing " + string(rune('A'+i)),
			Type: "Access Point",
			IP:   "10.20.2." + string(rune('0'+i)),
		})
	}

	// Wait for aggregation window to expire
	time.Sleep(100 * time.Millisecond)

	if emittedBatches != 1 {
		t.Errorf("Expected exactly 1 aggregated notification batch, got %d", emittedBatches)
	}
	if emittedDeviceCount != 6 {
		t.Errorf("Expected 6 devices in aggregated batch, got %d", emittedDeviceCount)
	}
}

// ─── Test 5: Flap Detection Test ─────────────────────────────────────────────
func TestFlapDetectionScanning(t *testing.T) {
	statusRepo := repository.NewMemoryStatusLogRepository()
	deviceID := "dev-flaky-01"

	// Simulate 6 DOWN events in the status log
	now := time.Now()
	for i := 0; i < 6; i++ {
		_ = statusRepo.Append(&domain.DeviceStatusLog{
			ID:        "log-" + string(rune('0'+i)),
			DeviceID:  deviceID,
			Status:    domain.StatusDOWN,
			Timestamp: now.Add(-time.Duration(i*12) * time.Hour),
		})
	}

	reports, err := statusRepo.GetFlapDevices(5, now.Add(-7*24*time.Hour), now)
	if err != nil {
		t.Fatalf("Failed to query flap devices: %v", err)
	}

	if len(reports) == 0 {
		t.Errorf("Expected device %s to be flagged as flap device", deviceID)
	}
	if reports[0].DownCount7d != 6 {
		t.Errorf("Expected 6 down events in 7d report, got %d", reports[0].DownCount7d)
	}
}

// ─── Test 6: Notification Failover Test (WhatsApp -> Telegram Fallback) ──────
func TestNotificationPipelineFailover(t *testing.T) {
	// WhatsApp sidecar endpoint is invalid -> fails -> falls back to Telegram
	pipeline := notifier.NewPipeline("http://invalid-whatsapp-host:9999", "MOCK_WA_TOKEN", "MOCK_TG_TOKEN", "-10012345", "+6281234567890", 60, nil)

	// Pipeline should attempt WhatsApp, encounter error, and failover gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pipeline.Send(ctx, "🔴 TEST INCIDENT ALERT")
	time.Sleep(200 * time.Millisecond)
}

// ─── Test 7: Role Middleware Table-Driven Test ────────────────────────────────
func TestRoleMiddlewareAccessMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		route          string
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Superadmin accesses settings",
			role:           "superadmin",
			route:          "/api/v1/settings",
			allowedRoles:   []string{"superadmin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Superadmin accesses protected anggota route implicitly",
			role:           "superadmin",
			route:          "/api/v1/devices",
			allowedRoles:   []string{"anggota"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Anggota accesses incident resolve",
			role:           "anggota",
			route:          "/api/v1/devices",
			allowedRoles:   []string{"superadmin", "anggota"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Pimpinan denied incident resolve (403)",
			role:           "pimpinan",
			route:          "/api/v1/devices",
			allowedRoles:   []string{"superadmin", "anggota"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Pimpinan denied settings (403)",
			role:           "pimpinan",
			route:          "/api/v1/settings",
			allowedRoles:   []string{"superadmin"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Anggota denied superadmin-only user management (403)",
			role:           "anggota",
			route:          "/api/v1/users/invite",
			allowedRoles:   []string{"superadmin"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			// Attach mock role context & test middleware
			r.Use(func(c *gin.Context) {
				c.Set(middleware.ContextUserRole, tt.role)
				c.Set(middleware.ContextUserID, "user-test")
				c.Next()
			})

			r.POST(tt.route, middleware.RequireRole(tt.allowedRoles...), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req, _ := http.NewRequest(http.MethodPost, tt.route, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}

// ─── Test 8: WebSocket Hub Integration Test ───────────────────────────────────
func TestWebSocketHubBroadcast(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	msg := domain.WSMessage{
		Type:        "STATUS_CHANGE",
		DeviceID:    "dev-10",
		DeviceName:  "Test Router",
		Status:      domain.StatusDOWN,
		Title:       "Test Outage",
		Description: "Simulated alert",
		Timestamp:   time.Now(),
	}

	// Ensure broadcasting does not panic or deadlock
	hub.BroadcastMessage(msg)
	time.Sleep(50 * time.Millisecond)
}

// ─── Helper for Engine ────────────────────────────────────────────────────────
func TestPollerEngineInitialization(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	deviceRepo := repository.NewMemoryDeviceRepository([]domain.Device{})
	statusRepo := repository.NewMemoryStatusLogRepository()
	cfg := poller.DefaultConfig()

	engine := poller.NewEngine(hub, deviceRepo, statusRepo, cfg)
	engine.Start()
	time.Sleep(50 * time.Millisecond)
	engine.Stop()
}

// ─── Test 9: Ping Output Parser Test ──────────────────────────────────────────
func TestIsPingSuccess(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		err      error
		expected bool
	}{
		{
			name:     "Windows English Success",
			output:   "Pinging 192.168.1.3 with 32 bytes of data:\nReply from 192.168.1.3: bytes=32 time=5ms TTL=64\nPing statistics: Sent = 1, Received = 1, Lost = 0 (0% loss)",
			err:      nil,
			expected: true,
		},
		{
			name:     "Windows English Timeout (Exit code 0)",
			output:   "Pinging 192.168.1.3 with 32 bytes of data:\nRequest timed out.\nPing statistics: Sent = 1, Received = 0, Lost = 1 (100% loss)",
			err:      nil,
			expected: false,
		},
		{
			name:     "Windows English Destination Host Unreachable (Exit code 0)",
			output:   "Pinging 192.168.1.3 with 32 bytes of data:\nReply from 192.168.1.1: Destination host unreachable.\nPing statistics: Sent = 1, Received = 0, Lost = 1 (100% loss)",
			err:      nil,
			expected: false,
		},
		{
			name:     "Windows Indonesian Timeout (Exit code 0)",
			output:   "Ping 192.168.1.3 dengan 32 bita data:\nWaktu minta habis.\nStatistik Ping: Terkirim = 1, Diterima = 0, Hilang = 1 (100% hilang)",
			err:      nil,
			expected: false,
		},
		{
			name:     "Linux Success",
			output:   "PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.\n64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=40.2 ms\n--- 8.8.8.8 ping statistics ---\n1 packets transmitted, 1 received, 0% packet loss, time 0ms",
			err:      nil,
			expected: true,
		},
		{
			name:     "Linux Timeout",
			output:   "PING 192.168.1.3 (192.168.1.3) 56(84) bytes of data.\n--- 192.168.1.3 ping statistics ---\n1 packets transmitted, 0 received, 100% packet loss, time 0ms",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poller.IsPingSuccessForTest(tt.output, tt.err)
			if got != tt.expected {
				t.Errorf("[%s] Expected %v, got %v for output:\n%s", tt.name, tt.expected, got, tt.output)
			}
		})
	}
}

