package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"sanoc/backend/internal/ai"
	"sanoc/backend/internal/config"
	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/handler"
	"sanoc/backend/internal/middleware"
	"sanoc/backend/internal/mailer"
	"sanoc/backend/internal/notifier"
	"sanoc/backend/internal/poller"
	"sanoc/backend/internal/repository"
	"sanoc/backend/internal/scheduler"
	"sanoc/backend/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
)

func main() {
	// ─── Configuration ────────────────────────────────────────────────────────
	cfg := config.LoadConfig()
	log.Printf("[SANOC] Environment configuration loaded. DB Host=%s, Redis Host=%s", cfg.DBHost, cfg.RedisHost)

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
	log.Println("[SANOC] Database migrations applied successfully.")


	redisClient, err := repository.InitRedis(cfg)
	if err != nil {
		log.Printf("[WARN] Redis connection issue: %v. Continuing with in-memory fallback cache.", err)
	}

	router := gin.Default()
	router.Use(middleware.SecurityHeaders())

	// ─── CORS ──────────────────────────────────────────────────────────────────
	corsOrigin := cfg.CORSAllowedOrigin
	if corsOrigin == "" {
		corsOrigin = "http://localhost:5173"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-CSRF-Token"},
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
				log.Printf("[SANOC] Seeded WhatsApp target from .env: %s", newTarget.PhoneNumber)
			}
		}
	}

	// Load Telegram config from DB or env fallback on startup
	var tgBotToken, tgChatID string
	var tgCfg handler.TelegramDBConfig
	if err := settingsRepo.GetJSON("telegram", &tgCfg); err == nil && tgCfg.BotToken != "" {
		tgBotToken = tgCfg.BotToken
		tgChatID = tgCfg.ChatID
		log.Printf("[SANOC] Loaded Telegram credentials from PostgreSQL: chat_id=%s", tgChatID)
	} else if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		tgBotToken = cfg.TelegramBotToken
		tgChatID = cfg.TelegramChatID
		_ = settingsRepo.SetJSON("telegram", handler.TelegramDBConfig{BotToken: tgBotToken, ChatID: tgChatID})
		log.Printf("[SANOC] Saved Telegram credentials from env to PostgreSQL: chat_id=%s", tgChatID)
	}

	// ─── Asynq Client Setup ──────────────────────────────────────────────────
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr()})
	defer asynqClient.Close()

	// Reset the rate-limiter's next_allowed_at Redis key to current time on startup
	if redisClient != nil && redisClient.Client != nil {
		err := redisClient.Client.Set(context.Background(), "notif:whatsapp:next_allowed_at", time.Now().Unix(), 0).Err()
		if err != nil {
			log.Printf("[Notifier] Failed to reset next_allowed_at in Redis: %v", err)
		} else {
			log.Println("[Notifier] Reset rate-limiter next_allowed_at in Redis to current time")
		}
	}

	// ─── Notification Pipeline & Queue ───────────────────────────────────────
	pipeline := notifier.NewPipeline(cfg.WhatsAppGatewayURL, cfg.WASidecarToken, tgBotToken, tgChatID, cfg.WhatsAppTargetNumber, 60, whatsappTargetRepo, redisClient, asynqClient)
	pipeline.SetNotifLogRepo(notifLogRepo)
	pipeline.SetIncidentRepo(incidentRepo)
	pipeline.SetSettingsRepo(settingsRepo)
	pipeline.SetDeviceRepo(deviceRepo)

	notifyQueue := notifier.NewNotifyQueue(pipeline, notifLogRepo)
	defer notifyQueue.Stop()
	pipeline.SetNotifyQueue(notifyQueue)


	// ─── Asynq Server Worker Setup ───────────────────────────────────────────
	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr()},
		asynq.Config{
			Concurrency: 1, // Sequential execution for strict spacing
			Queues: map[string]int{
				"whatsapp": 1,
			},
		},
	)
	asynqMux := asynq.NewServeMux()
	asynqMux.HandleFunc(notifier.TypeWhatsAppDispatch, notifier.HandleWhatsAppDispatchTask(pipeline))

	go func() {
		log.Printf("[Asynq Server] Starting worker server listening on queue 'whatsapp' (concurrency=1)...")
		if err := asynqServer.Run(asynqMux); err != nil {
			log.Fatalf("[Asynq Server] Fatal error: %v", err)
		}
	}()
	defer asynqServer.Shutdown()

	// Run one-time incident deduplication & stale incident cleanup on startup
	if db != nil {
		_ = repository.CleanupDuplicateAndStaleIncidents(db)
	}

	// Scheduled incident retention job (archives old resolved incidents)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if db != nil {
				var sysSettings domain.SystemSettings
				retentionDays := 90
				if settingsRepo != nil && settingsRepo.GetJSON("system_settings", &sysSettings) == nil && sysSettings.RetentionDays > 0 {
					retentionDays = sysSettings.RetentionDays
				}
				_, _ = repository.ArchiveOldResolvedIncidents(db, retentionDays)
			}
		}
	}()

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

	// ─── DHCP Auto-Sync Engine ──────────────────────────────────────────────────
	dhcpSyncWorker := poller.NewDHCPSyncWorker(cfg, deviceRepo)
	dhcpSyncWorker.Start()
	defer dhcpSyncWorker.Stop()


	log.Printf("[Settings] Loaded startup configuration:")
	log.Printf("[Settings]   • Polling Interval: %ds", pollCfg.IntervalSeconds)
	log.Printf("[Settings]   • Concurrency/Batch Size: %d", pollCfg.ConcurrencyBatchSize)
	log.Printf("[Settings]   • Default Debounce: %ds", pollCfg.DebounceSeconds)
	log.Printf("[Settings]   • Thresholds: %v", pollCfg.Thresholds)
	log.Printf("[Settings]   • Rate Limit: %d msgs/min", pipeline.GetRateLimitMaxMsgPerMin())

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
	h.SetWhatsAppTargetRepo(whatsappTargetRepo)
	h.SetMailer(mailer.NewMailer(cfg))
	// ─── AI Service (Google Gemini) ─────────────────────────────────────────────
	aiService := ai.NewService(cfg.GeminiAPIKey, cfg.GeminiModel, deviceRepo, incidentRepo, statusRepo)
	h.SetAIService(aiService)
	if aiService.IsConfigured() {
		log.Printf("[SANOC] Gemini AI Copilot initialized successfully (Model: %s)", aiService.GetModel())
	} else {
		log.Printf("[SANOC] Gemini AI Copilot: GEMINI_API_KEY not configured in .env (Copilot ready on key provision)")
	}
	
	// ─── Flap Report Job (Weekly Scheduler) ──────────────────────────────────
	flapReportJob := scheduler.NewFlapReportJob(statusRepo, deviceRepo, pipeline, aiService)
	flapReportJob.RunWeekly()

	importH := handler.NewImportHandler(deviceRepo, locationRepo)

	// Integrations handler — uses real sidecar proxy (URL + internal token from .env)
	integH := handler.NewIntegrationsHandler(hub, pipeline, settingsRepo, whatsappTargetRepo, cfg.WhatsAppGatewayURL, cfg.WASidecarToken)
	log.Printf("[SANOC] WhatsApp sidecar URL: %s (token auth: %v)", cfg.WhatsAppGatewayURL, cfg.WASidecarToken != "")

	// ─── Health, Static Files & Root Landing Route ────────────────────────────────
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!DOCTYPE html><html><body><h1>SANOC API</h1></body></html>`)
	})
	// Static uploaded files with security headers
	uploadsGroup := router.Group("/uploads")
	uploadsGroup.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Security-Policy", "default-src 'none'; img-src 'self'")
		c.Next()
	})
	uploadsGroup.Static("", "./uploads")

	// ─── WebSocket ────────────────────────────────────────────────────────────
	router.GET("/ws/live", hub.HandleWS)

	// ─── REST API v1 ───────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTMiddleware())
	v1.Use(middleware.CSRFProtection())

	// Auth routes
	v1.GET("/auth/me", h.GetMe)
	v1.PUT("/auth/profile", h.UpdateProfile)
	v1.POST("/auth/profile/send-reset-otp", h.SendProfileResetOTP)
	v1.POST("/auth/profile/reset-password-otp", h.ResetProfilePasswordByOTP)
	v1.POST("/auth/avatar", h.UploadAvatar)
	v1.POST("/auth/mfa/setup", h.SetupMFA)
	v1.POST("/auth/mfa/verify", h.VerifyMFA)
	v1.POST("/auth/mfa/disable", h.DisableMFA)

	router.POST("/api/v1/auth/login", middleware.RateLimitLogin(cfg.RateLimitLoginMax, 1*time.Minute), h.Login)
	router.POST("/api/v1/auth/verify-mfa", middleware.RateLimitLogin(cfg.RateLimitLoginMax, 1*time.Minute), h.VerifyLoginMFA)
	router.POST("/api/v1/auth/logout", h.Logout)
	router.POST("/api/v1/auth/forgot-password", h.ForgotPassword)
	router.POST("/api/v1/auth/reset-password", h.ResetPassword)
	router.POST("/api/v1/auth/send-verification-otp", h.SendAccountVerificationOTP)
	router.POST("/api/v1/auth/verify-account-otp", h.VerifyAccountOTP)
	router.GET("/api/v1/settings/branding", h.GetBranding)

	// User logs (accessible for activity log viewing)
	v1.GET("/user-logs", h.GetUserLogs)

	// Notifications log & audit
	notifH := handler.NewNotificationsHandler(incidentRepo)
	v1.GET("/notifications", notifH.GetNotifications)
	v1.GET("/notifications/logs", notifH.GetNotificationLogs)
	v1.PATCH("/notifications/read-all", notifH.MarkAllAsRead)
	v1.DELETE("/notifications", notifH.ClearNotifications)

	// Dashboard & Polling Control
	v1.GET("/dashboard/summary", h.GetSummary)
	v1.POST("/monitoring/refresh-now", h.RefreshNow)

	// Locations
	v1.GET("/locations", h.GetLocations)
	v1.POST("/locations", middleware.RequirePermission(permRepo, "settings.locations", "admin"), h.CreateLocation)
	v1.PUT("/locations/:id", middleware.RequirePermission(permRepo, "settings.locations", "admin"), h.UpdateLocation)
	v1.DELETE("/locations/:id", middleware.RequirePermission(permRepo, "settings.locations", "admin"), h.DeleteLocation)

	// Permissions Matrix
	v1.GET("/permissions", h.GetPermissions)
	v1.PUT("/permissions", middleware.RequireRole("admin"), h.UpdatePermissions)

	// Devices
	devices := v1.Group("/devices")
	{
		devices.GET("", middleware.RequirePermission(permRepo, "devices.view", "admin", "anggota", "pimpinan"), h.GetDevices)
		devices.GET("/:id", middleware.RequirePermission(permRepo, "devices.view", "admin", "anggota", "pimpinan"), h.GetDeviceByID)
		devices.GET("/:id/status-history", middleware.RequirePermission(permRepo, "devices.view", "admin", "anggota", "pimpinan"), h.GetStatusHistory)
		devices.GET("/:id/metrics", middleware.RequirePermission(permRepo, "devices.view", "admin", "anggota", "pimpinan"), h.GetDeviceMetrics)

		devices.POST("", middleware.RequirePermission(permRepo, "devices.create", "admin", "anggota"), h.CreateDevice)
		devices.PUT("/:id", middleware.RequirePermission(permRepo, "devices.edit", "admin", "anggota"), h.UpdateDevice)
		devices.DELETE("/:id", middleware.RequirePermission(permRepo, "devices.delete", "admin"), h.DeleteDevice)
		devices.PATCH("/bulk", middleware.RequirePermission(permRepo, "devices.bulk", "admin"), h.BulkDeviceAction)
		devices.POST("/import", middleware.RequirePermission(permRepo, "devices.import", "admin"), importH.ImportDevices)
		devices.GET("/auto-detect", middleware.RequirePermission(permRepo, "devices.view", "admin", "anggota"), h.AutoDetect)
	}


	// Incidents
	incidents := v1.Group("/incidents")
	{
		incidents.GET("", middleware.RequirePermission(permRepo, "incidents.view", "admin", "anggota", "pimpinan"), h.GetIncidents)
		incidents.GET("/:id", middleware.RequirePermission(permRepo, "incidents.view", "admin", "anggota", "pimpinan"), h.GetIncidentByID)
		incidents.GET("/:id/events", middleware.RequirePermission(permRepo, "incidents.view", "admin", "anggota", "pimpinan"), h.GetIncidentEvents)
	}

	// Reports
	reports := v1.Group("/reports")
	{
		reports.GET("", middleware.RequirePermission(permRepo, "reports.view", "admin", "anggota", "pimpinan"), h.GetReport)
		reports.GET("/flap-devices", middleware.RequirePermission(permRepo, "reports.view", "admin", "anggota", "pimpinan"), h.GetFlapDevices)
	}

	// Settings
	settings := v1.Group("/settings")
	{
		settings.GET("", h.GetSettings)
		settings.PUT("", h.UpdateSettings)
		settings.PUT("/thresholds", middleware.RequirePermission(permRepo, "settings.polling", "settings.thresholds", "admin"), h.UpdateThresholds)
		settings.PUT("/polling", middleware.RequirePermission(permRepo, "settings.polling", "admin"), h.UpdatePolling)
		settings.PUT("/rate-limit", middleware.RequirePermission(permRepo, "settings.notifications", "admin"), h.UpdateRateLimit)
		settings.PUT("/branding", middleware.RequirePermission(permRepo, "settings.branding", "admin"), h.UpdateBranding)
		settings.POST("/branding/upload", middleware.RequirePermission(permRepo, "settings.branding", "admin"), h.UploadBrandingAsset)
		settings.GET("/dhcp/logs", middleware.RequirePermission(permRepo, "settings.polling", "admin"), h.GetDHCPLogs)
	}

	// Diagnostics
	diagnostics := v1.Group("/diagnostics")
	{
		diagnostics.POST("/ping", middleware.RequirePermission(permRepo, "diagnostics.run", "admin", "anggota"), h.RunPing)
		diagnostics.POST("/traceroute", middleware.RequirePermission(permRepo, "diagnostics.run", "admin", "anggota"), h.RunTraceroute)
		diagnostics.POST("/port-probe", middleware.RequirePermission(permRepo, "diagnostics.run", "admin", "anggota"), h.RunPortProbe)
	}

	// AI Copilot
	aiGroup := v1.Group("/ai")
	{
		aiGroup.GET("/status", h.GetAIStatus)
		aiGroup.GET("/quick-prompts", h.GetAIQuickPrompts)
		aiGroup.POST("/chat", h.ChatAI)
		aiGroup.POST("/analyze-incident/:id", h.AnalyzeIncidentAI)
	}

	// Users
	users := v1.Group("/users")
	users.Use(middleware.RequirePermission(permRepo, "settings.users", "admin"))
	{
		users.GET("", h.GetUsers)
		users.GET("/logs", h.GetUserLogs)
		users.POST("", h.CreateUser)
		users.POST("/create", h.CreateUser)
		users.PUT("/:id/role", h.UpdateUserRole)
		users.PUT("/:id", h.UpdateUser)
		users.PUT("/:id/reset-password", h.ResetUserPassword)
		users.DELETE("/:id", h.DeleteUser)
	}

	// Integrations
	integrations := v1.Group("/integrations")
	integrations.Use(middleware.RequirePermission(permRepo, "settings.notifications", "admin"))
	{
		integrations.POST("/whatsapp/connect", integH.WhatsAppConnect)
		integrations.GET("/whatsapp/qr", integH.WhatsAppQR)
		integrations.GET("/whatsapp/status", integH.WhatsAppStatus)
		integrations.POST("/whatsapp/disconnect", integH.WhatsAppDisconnect)
		integrations.GET("/whatsapp/targets", integH.ListWhatsAppTargets)
		integrations.POST("/whatsapp/targets", integH.CreateWhatsAppTarget)
		integrations.PUT("/whatsapp/targets/:id", integH.UpdateWhatsAppTarget)
		integrations.DELETE("/whatsapp/targets/:id", integH.DeleteWhatsAppTarget)
		integrations.POST("/whatsapp/test", integH.WhatsAppTest)
		integrations.GET("/telegram/config", integH.TelegramGetConfig)
		integrations.POST("/telegram/config", integH.TelegramConfig)
		integrations.POST("/telegram/test", integH.TelegramTest)
	}


	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("[SANOC] Go Backend listening on :%s...", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[SANOC] Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("[SANOC] Server forced to shutdown: ", err)
	}

	// Close DB, Redis, and stop Poller Engine
	if db != nil {
		log.Println("[SANOC] Closing database connections...")
		db.Close()
	}
	if redisClient != nil {
		log.Println("[SANOC] Closing Redis connection...")
		redisClient.Client.Close()
	}
	if pollerEngine != nil {
		log.Println("[SANOC] Stopping Poller Engine...")
		pollerEngine.Stop()
	}
	if dhcpSyncWorker != nil {
		log.Println("[SANOC] Stopping DHCP Sync Engine...")
		dhcpSyncWorker.Stop()
	}

	log.Println("[SANOC] Server exiting")
}

func seedDevices() []domain.Device {
	return []domain.Device{
		{ID: "dev-local", Name: "Test - Localhost", Type: domain.Router, IP: "127.0.0.1", MAC: "00:00:00:00:00:01", Status: domain.StatusUP, AddressingMode: domain.AddressingStatic, Location: "Local Host Server", Rack: "Rack L-01", CheckedSecondsAgo: 2, LastChecked: "14:02:10", Uptime30d: 100.0, FailureThreshold: 2, SNMPEnabled: true, SNMPCommunity: "public", SNMPPort: 161},
		{ID: "dev-cloudflare", Name: "Test - Cloudflare DNS", Type: domain.AccessPoint, IP: "1.1.1.1", MAC: "00:00:00:00:00:02", Status: domain.StatusUP, AddressingMode: domain.AddressingStatic, Location: "Cloudflare Public DNS", Rack: "Edge 1", CheckedSecondsAgo: 5, LastChecked: "14:02:08", Uptime30d: 99.9, FailureThreshold: 3},
		{ID: "dev-google", Name: "Test - Google DNS", Type: domain.Switch, IP: "8.8.8.8", MAC: "00:00:00:00:00:03", Status: domain.StatusUP, AddressingMode: domain.AddressingStatic, Location: "Google Public DNS", Rack: "Edge 2", CheckedSecondsAgo: 1, LastChecked: "14:02:15", Uptime30d: 99.9, FailureThreshold: 3},
	}
}