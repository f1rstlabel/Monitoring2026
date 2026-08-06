package main

import (
	"log"
	"net/url"
	"strings"
	"time"

	"govmonitor-it/backend/internal/config"
	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/handler"
	"govmonitor-it/backend/internal/middleware"
	"govmonitor-it/backend/internal/notifier"
	"govmonitor-it/backend/internal/poller"
	"govmonitor-it/backend/internal/repository"
	"govmonitor-it/backend/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// ─── Configuration ────────────────────────────────────────────────────────
	cfg := config.LoadConfig()
	log.Printf("[GovMonitor IT] Environment configuration loaded. DB Host=%s, Redis Host=%s", cfg.DBHost, cfg.RedisHost)

	// ─── Database & Cache Connections ─────────────────────────────────────────
	db, err := repository.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Could not connect to PostgreSQL: %v", err)
	}

	// ─── Run Migrations ───────────────────────────────────────────────────────
	// Construct connection string for migrate, ensuring it has 'sslmode=disable' if not already present
	// to avoid issues with older postgres versions or dev environments.
	parsedURL, err := url.Parse(cfg.PostgresDSN()) // Corrected: cfg.PostgresDSN()
	if err != nil {
		log.Fatalf("[FATAL] Failed to parse database URL for migrations: %v", err)
	}
	q := parsedURL.Query()
	if q.Get("sslmode") == "" {
		q.Add("sslmode", "disable")
		parsedURL.RawQuery = q.Encode()
	}
	dbURL := parsedURL.String()

	m, err := migrate.New(
		"file://./migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize database migrations: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("[FATAL] Failed to run database migrations: %v", err)
	}
	log.Println("[GovMonitor IT] Database migrations applied successfully.")


	redisClient, err := repository.InitRedis(cfg)
	if err != nil {
		log.Printf("[WARN] Redis connection issue: %v. Continuing with in-memory fallback cache.", err)
	}

	router := gin.Default()

	// ─── CORS ──────────────────────────────────────────────────────────────────
	corsOrigin := cfg.CORSAllowedOrigin
	if corsOrigin == "" {
		corsOrigin = "http://localhost:5173"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ─── WebSocket Hub ─────────────────────────────────────────────────────────
	hub := ws.NewHub()
	go hub.Run()
	metricRepo := repository.NewPostgresDeviceMetricRepository(db)

	// ─── Repositories (PostgreSQL + Redis Caching) ──────────────────────────────
	var deviceRepo repository.DeviceRepository
	if db != nil {
		deviceRepo = repository.NewPostgresDeviceRepository(db, redisClient)
	} else {
		deviceRepo = repository.NewMemoryDeviceRepository(seedDevices())
	}
	var statusRepo repository.StatusLogRepository
	if db != nil {
		statusRepo = repository.NewPostgresStatusLogRepository(db)
	} else {
		statusRepo = repository.NewMemoryStatusLogRepository()
	}
	userRepo := repository.NewPostgresUserRepository(db)
	whatsappTargetRepo := repository.NewPostgresWhatsAppTargetRepository(db)
	incidentRepo := repository.NewPostgresIncidentRepository(db)
	notifLogRepo := repository.NewPostgresNotificationLogRepository(db)
	userLogRepo := repository.NewPostgresUserLogRepository(db)
	locationRepo := repository.NewPostgresLocationRepository(db)
	permRepo := repository.NewPostgresPermissionRepository(db)

	// ─── Settings Repository (Postgres-backed, auto-creates table) ─────────────
	settingsRepo := repository.NewSettingsRepository(db)

	// Seed WHATSAPP_TARGET_NUMBER from .env if table is empty
	if cfg.WhatsAppTargetNumber != "" {
		targets, err := whatsappTargetRepo.GetAll()
		if err != nil {
			log.Printf("[WARN] Failed to read WhatsApp targets for seeding: %v", err)
		} else if len(targets) == 0 {
			// Normalize phone number to JID
			phoneNumberDigits := strings.ReplaceAll(cfg.WhatsAppTargetNumber, " ", "")
			phoneNumberDigits = strings.ReplaceAll(phoneNumberDigits, "+", "")
			phoneNumberDigits = strings.ReplaceAll(phoneNumberDigits, "-", "")
			if strings.HasPrefix(phoneNumberDigits, "0") && len(phoneNumberDigits) > 1 {
				phoneNumberDigits = "62" + phoneNumberDigits[1:]
			}

			jid := phoneNumberDigits + "@s.whatsapp.net"
			newTarget := &domain.WhatsAppTarget{
				Label:       "Default from .env",
				PhoneNumber: cfg.WhatsAppTargetNumber,
				JID:         jid,
			}
			_, err := whatsappTargetRepo.Create(newTarget)
			if err != nil {
				log.Printf("[WARN] Failed to seed WhatsApp target from .env: %v", err)
			} else {
				log.Printf("[GovMonitor IT] Seeded WhatsApp target from .env: %s", newTarget.PhoneNumber)
			}
		}
	}

	// Load Telegram config from DB or env fallback on startup
	var tgBotToken, tgChatID string
	var tgCfg handler.TelegramDBConfig
	if err := settingsRepo.GetJSON("telegram", &tgCfg); err == nil && tgCfg.BotToken != "" {
		tgBotToken = tgCfg.BotToken
		tgChatID = tgCfg.ChatID
		log.Printf("[GovMonitor IT] Loaded Telegram credentials from PostgreSQL: chat_id=%s", tgChatID)
	} else if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		tgBotToken = cfg.TelegramBotToken
		tgChatID = cfg.TelegramChatID
		_ = settingsRepo.SetJSON("telegram", handler.TelegramDBConfig{BotToken: tgBotToken, ChatID: tgChatID})
		log.Printf("[GovMonitor IT] Saved Telegram credentials from env to PostgreSQL: chat_id=%s", tgChatID)
	}

	// ─── Notification Pipeline & Queue ───────────────────────────────────────
	pipeline := notifier.NewPipeline(
		cfg.WhatsAppGatewayURL,
		cfg.WASidecarToken,
		tgBotToken,
		tgChatID,
		cfg.WhatsAppTargetNumber,
		60,
		whatsappTargetRepo,
	)
	pipeline.SetNotifLogRepo(notifLogRepo)
	notifyQueue := notifier.NewNotifyQueue(pipeline, notifLogRepo)
	defer notifyQueue.Stop()

	// ─── Polling Engine ─────────────────────────────────────────────────────────
	pollCfg := poller.DefaultConfig()
	pollerEngine := poller.NewEngine(hub, deviceRepo, statusRepo, pollCfg)
	pollerEngine.SetPipeline(pipeline)
	pollerEngine.SetNotifyQueue(notifyQueue)
	pollerEngine.SetSettingsRepo(settingsRepo)
	pollerEngine.SetMetricRepo(metricRepo)
	pollerEngine.SetIncidentRepo(incidentRepo)
	pollerEngine.Start()
	defer pollerEngine.Stop()

	// ─── Handlers ──────────────────────────────────────────────────────────────
	h := handler.NewHandler(hub, settingsRepo, userRepo, deviceRepo, statusRepo)
	h.SetPoller(pollerEngine)
	h.SetPipeline(pipeline)
	h.SetMetricRepo(metricRepo)
	h.SetIncidentRepo(incidentRepo)
	h.SetUserLogRepo(userLogRepo)
	h.SetNotifLogRepo(notifLogRepo)
	h.SetLocationRepo(locationRepo)
	h.SetPermissionRepo(permRepo)
	importH := handler.NewImportHandler(deviceRepo)

	// Integrations handler — uses real sidecar proxy (URL + internal token from .env)
	integH := handler.NewIntegrationsHandler(hub, pipeline, settingsRepo, whatsappTargetRepo, cfg.WhatsAppGatewayURL, cfg.WASidecarToken)
	log.Printf("[GovMonitor IT] WhatsApp sidecar URL: %s (token auth: %v)", cfg.WhatsAppGatewayURL, cfg.WASidecarToken != "")

	// ─── Health & Root Landing Route ───────────────────────────────────────────
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!DOCTYPE html><html><body><h1>GovMonitor IT API</h1></body></html>`)
	})

	// ─── WebSocket ────────────────────────────────────────────────────────────
	router.GET("/ws/live", hub.HandleWS)

	// ─── REST API v1 ───────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTMiddleware())

	// Auth routes
	v1.GET("/auth/me", h.GetMe)
	router.POST("/api/v1/auth/login", h.Login)
	router.POST("/api/v1/auth/logout", h.Logout)
	router.POST("/api/v1/auth/forgot-password", h.ForgotPassword)
	router.POST("/api/v1/auth/reset-password", h.ResetPassword)

	// Dashboard & Polling Control
	v1.GET("/dashboard/summary", h.GetSummary)
	v1.POST("/monitoring/refresh-now", h.RefreshNow)

	// Locations
	v1.GET("/locations", h.GetLocations)
	v1.POST("/locations", middleware.RequirePermission(permRepo, "devices.create", "superadmin", "anggota"), h.CreateLocation)

	// Permissions Matrix
	v1.GET("/permissions", h.GetPermissions)
	v1.PUT("/permissions", middleware.RequireRole("superadmin"), h.UpdatePermissions)

	// Devices
	devices := v1.Group("/devices")
	{
		devices.GET("", h.GetDevices)
		devices.GET("/:id", h.GetDeviceByID)
		devices.GET("/:id/status-history", h.GetStatusHistory)
		devices.GET("/:id/metrics", h.GetDeviceMetrics)

		devices.POST("", middleware.RequirePermission(permRepo, "devices.create", "superadmin", "anggota"), h.CreateDevice)
		devices.PUT("/:id", middleware.RequirePermission(permRepo, "devices.edit", "superadmin", "anggota"), h.UpdateDevice)
		devices.DELETE("/:id", middleware.RequirePermission(permRepo, "devices.delete", "superadmin"), h.DeleteDevice)
		devices.POST("/import", middleware.RequirePermission(permRepo, "devices.import", "superadmin"), importH.ImportDevices)
		devices.GET("/auto-detect", h.AutoDetect)
	}

	// Incidents
	incidents := v1.Group("/incidents")
	{
		incidents.GET("", h.GetIncidents)
		incidents.GET("/:id", h.GetIncidentByID)
	}

	// Reports
	reports := v1.Group("/reports")
	{
		reports.GET("", h.GetReport)
		reports.GET("/flap-devices", h.GetFlapDevices)
	}

	// Settings
	settings := v1.Group("/settings")
	settings.Use(middleware.RequireRole("superadmin"))
	{
		settings.GET("", h.GetSettings)
		settings.PUT("/thresholds", h.UpdateThresholds)
		settings.PUT("/polling", h.UpdatePolling)
		settings.PUT("/rate-limit", h.UpdateRateLimit)
	}

	// Users
	users := v1.Group("/users")
	users.Use(middleware.RequireRole("superadmin"))
	{
		users.GET("", h.GetUsers)
		users.GET("/logs", h.GetUserLogs)
		users.POST("/invite", h.InviteUser)
		users.PUT("/:id/role", h.UpdateUserRole)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}

	// Integrations
	integrations := v1.Group("/integrations")
	{
		integrations.POST("/whatsapp/connect", integH.WhatsAppConnect)
		integrations.GET("/whatsapp/qr", integH.WhatsAppQR)
		integrations.GET("/whatsapp/status", integH.WhatsAppStatus)
		integrations.POST("/whatsapp/disconnect", integH.WhatsAppDisconnect)
		integrations.GET("/whatsapp/targets", integH.ListWhatsAppTargets) // New endpoint
		integrations.POST("/whatsapp/targets", integH.CreateWhatsAppTarget) // New endpoint
		integrations.PUT("/whatsapp/targets/:id", integH.UpdateWhatsAppTarget) // New endpoint
		integrations.DELETE("/whatsapp/targets/:id", integH.DeleteWhatsAppTarget) // New endpoint
		integrations.POST("/whatsapp/test", integH.WhatsAppTest)
		integrations.GET("/telegram/config", integH.TelegramGetConfig)
		integrations.POST("/telegram/config", integH.TelegramConfig)
		integrations.POST("/telegram/test", integH.TelegramTest)
	}

	// Notifications
	notifH := handler.NewNotificationsHandler(incidentRepo)
	v1.GET("/notifications", notifH.GetNotifications)
	v1.PATCH("/notifications/read-all", notifH.MarkAllAsRead)
	v1.DELETE("/notifications", notifH.ClearNotifications)


	log.Printf("[GovMonitor IT] Go Backend listening on :%s...", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func seedDevices() []domain.Device {
	return []domain.Device{
		{ID: "dev-local", Name: "Test - Localhost", Type: domain.Router, IP: "127.0.0.1", MAC: "00:00:00:00:00:01", Status: domain.StatusUP, AddressingMode: domain.AddressingStatic, Location: "Local Host Server", Rack: "Rack L-01", CheckedSecondsAgo: 2, LastChecked: "14:02:10", Uptime30d: 100.0, FailureThreshold: 2, SNMPEnabled: true, SNMPCommunity: "public", SNMPPort: 161},
		{ID: "dev-cloudflare", Name: "Test - Cloudflare DNS", Type: domain.AccessPoint, IP: "1.1.1.1", MAC: "00:00:00:00:00:02", Status: domain.StatusUP, AddressingMode: domain.AddressingStatic, Location: "Cloudflare Public DNS", Rack: "Edge 1", CheckedSecondsAgo: 5, LastChecked: "14:02:08", Uptime30d: 99.9, FailureThreshold: 3},
		{ID: "dev-google", Name: "Test - Google DNS", Type: domain.Switch, IP: "8.8.8.8", MAC: "00:00:00:00:00:03", Status: domain.StatusUP, AddressingMode: domain.AddressingStatic, Location: "Google Public DNS", Rack: "Edge 2", CheckedSecondsAgo: 1, LastChecked: "14:02:15", Uptime30d: 99.9, FailureThreshold: 3},
	}
}