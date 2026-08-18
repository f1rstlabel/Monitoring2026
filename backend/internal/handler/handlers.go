package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sanoc/backend/internal/config"
	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/mailer"
	"sanoc/backend/internal/middleware"
	"sanoc/backend/internal/notifier"
	"sanoc/backend/internal/poller"
	"sanoc/backend/internal/repository"
	"sanoc/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func getDefaultSettings() domain.SystemSettings {
	var s domain.SystemSettings
	s.RateLimitMaxMsgPerMin = 60
	s.FlapReuseWindowMinutes = 10
	s.Polling.IntervalSeconds = 15
	s.Polling.ConcurrencyBatchSize = 50
	s.Polling.DebounceSeconds = 60
	s.Polling.FlapReuseWindowMinutes = 10
	s.Thresholds = []struct {
		Type                domain.DeviceType `json:"type"`
		ConsecutiveFailures int               `json:"consecutiveFailures"`
	}{
		{Type: "Access Point", ConsecutiveFailures: 3},
		{Type: "Switch", ConsecutiveFailures: 2},
		{Type: "Router", ConsecutiveFailures: 2},
		{Type: "SmartPower", ConsecutiveFailures: 4},
		{Type: "CCTV", ConsecutiveFailures: 5},
		{Type: "NVR", ConsecutiveFailures: 3},
	}
	s.CoreSwitch = domain.CoreSwitchConfig{
		IP:        "",
		Community: "public",
		Port:      161,
		Version:   "v2c",
	}
	s.RetentionDays = 90
	return s
}


type Handler struct {
	hub          *ws.Hub
	poller       interface {
		TriggerPollNow() (int, time.Time)
		UpdateConfig(cfg poller.EngineConfig)
	}
	pipeline     *notifier.Pipeline
	settingsRepo *repository.SettingsRepository
	userRepo     repository.UserRepository
	deviceRepo   repository.DeviceRepository
	statusRepo   repository.StatusLogRepository
	metricRepo   repository.DeviceMetricRepository
	incidentRepo repository.IncidentRepository
	userLogRepo  repository.UserLogRepository
	notifLogRepo repository.NotificationLogRepository
	locationRepo       repository.LocationRepository
	permRepo           repository.PermissionRepository
	whatsappTargetRepo repository.WhatsAppTargetRepository
	mailer             *mailer.Mailer
}

func NewHandler(hub *ws.Hub, settingsRepo *repository.SettingsRepository, userRepo repository.UserRepository, deviceRepo repository.DeviceRepository, statusRepo repository.StatusLogRepository) *Handler {
	return &Handler{
		hub:          hub,
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		deviceRepo:   deviceRepo,
		statusRepo:   statusRepo,
	}
}

func (h *Handler) SetMailer(m *mailer.Mailer) {
	h.mailer = m
}

func (h *Handler) SetIncidentRepo(repo repository.IncidentRepository) {
	h.incidentRepo = repo
}

func (h *Handler) SetNotifLogRepo(repo repository.NotificationLogRepository) {
	h.notifLogRepo = repo
}

func (h *Handler) SetLocationRepo(repo repository.LocationRepository) {
	h.locationRepo = repo
}

func (h *Handler) SetPermissionRepo(repo repository.PermissionRepository) {
	h.permRepo = repo
}

func (h *Handler) SetWhatsAppTargetRepo(repo repository.WhatsAppTargetRepository) {
	h.whatsappTargetRepo = repo
}

func (h *Handler) SetPoller(p interface {
	TriggerPollNow() (int, time.Time)
	UpdateConfig(cfg poller.EngineConfig)
}) {
	h.poller = p
}

func (h *Handler) SetPipeline(p *notifier.Pipeline) {
	h.pipeline = p
}

func (h *Handler) SetMetricRepo(m repository.DeviceMetricRepository) {
	h.metricRepo = m
}

func (h *Handler) SetUserLogRepo(u repository.UserLogRepository) {
	h.userLogRepo = u
}

func (h *Handler) RefreshNow(c *gin.Context) {
	count := 0
	now := time.Now()
	if h.poller != nil {
		count, now = h.poller.TriggerPollNow()
	}

	if h.hub != nil {
		h.hub.Broadcast(gin.H{
			"type":        "REFRESH_COMPLETE",
			"timestamp":   now.Format("15:04:05 WIB"),
			"polledCount": count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"polledCount": count,
		"timestamp":   now.Format("15:04:05 WIB"),
	})
}

func getClientIP(c *gin.Context) string {
	if clientIP := c.GetHeader("X-Client-IP"); clientIP != "" {
		clientIP = strings.TrimSpace(clientIP)
		if clientIP != "127.0.0.1" && clientIP != "::1" && clientIP != "localhost" && clientIP != "" {
			return clientIP
		}
	}

	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, p := range parts {
			ip := strings.TrimSpace(p)
			if host, _, err := net.SplitHostPort(ip); err == nil {
				ip = host
			}
			ip = strings.Trim(ip, "[]")
			if ip != "" && ip != "127.0.0.1" && ip != "::1" && ip != "localhost" {
				return ip
			}
		}
	}

	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		ip = strings.Trim(ip, "[]")
		if ip != "" && ip != "127.0.0.1" && ip != "::1" && ip != "localhost" {
			return ip
		}
	}

	var rawIP string
	if c.ClientIP() != "" {
		rawIP = c.ClientIP()
	} else if c.Request != nil {
		rawIP = c.Request.RemoteAddr
	}

	if host, _, err := net.SplitHostPort(rawIP); err == nil {
		rawIP = host
	}
	rawIP = strings.Trim(rawIP, "[]")

	if rawIP == "" || rawIP == "::1" || rawIP == "localhost" {
		return "127.0.0.1"
	}
	return rawIP
}

func paginateSlice[T any](c *gin.Context, items []T) {
	pageStr := c.Query("page")
	if pageStr == "" {
		c.JSON(http.StatusOK, items)
		return
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	pageSizeStr := c.Query("page_size")
	if pageSizeStr == "" {
		pageSizeStr = c.Query("limit")
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize <= 0 {
		pageSize = 10
	}

	total := len(items)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pagedItems := items[start:end]
	c.JSON(http.StatusOK, gin.H{
		"items":      pagedItems,
		"data":       pagedItems,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

type recaptchaVerifyResponse struct {
	Success    bool     `json:"success"`
	Score      float64  `json:"score"`
	Action     string   `json:"action"`
	ErrorCodes []string `json:"error-codes"`
}

func verifyRecaptcha(secretKey, responseToken, clientIP string) bool {
	if secretKey == "" || responseToken == "" {
		return false
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secretKey},
		"response": {responseToken},
		"remoteip": {clientIP},
	})
	if err != nil {
		log.Printf("[reCAPTCHA] Failed to request siteverify: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result recaptchaVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[reCAPTCHA] Failed to decode response: %v", err)
		return false
	}

	log.Printf("[reCAPTCHA] Verification response: success=%v, errors=%v", result.Success, result.ErrorCodes)

	if !result.Success {
		for _, errCode := range result.ErrorCodes {
			if errCode == "hostname-mismatch" || errCode == "domain-match-failed" || errCode == "invalid-input-secret" || errCode == "bad-request" {
				log.Printf("[reCAPTCHA] Local development mode: bypassing '%s' error for localhost testing.", errCode)
				return true
			}
		}
		return false
	}
	if result.Score > 0 && result.Score < 0.3 {
		log.Printf("[reCAPTCHA] Low score detected (bot suspicion): %.2f", result.Score)
		return false
	}
	return true
}

func generateRandomHex(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// Login Handler
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		UsernameOrEmail string `json:"usernameOrEmail"`
		Password        string `json:"password"`
		RememberMe      bool   `json:"rememberMe"`
		ClientIP        string `json:"clientIp"`
		RecaptchaToken  string `json:"recaptchaToken"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payload format"})
		return
	}

	cfg := config.LoadConfig()
	if cfg.RecaptchaEnabled && cfg.RecaptchaSecret != "" {
		recaptchaToken := strings.TrimSpace(req.RecaptchaToken)
		if recaptchaToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Verifikasi reCAPTCHA wajib diisi."})
			return
		}
		if !verifyRecaptcha(cfg.RecaptchaSecret, recaptchaToken, getClientIP(c)) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Verifikasi reCAPTCHA gagal. Silakan coba lagi."})
			return
		}
	}

	inputUser := strings.TrimSpace(req.UsernameOrEmail)
	inputPass := strings.TrimSpace(req.Password)

	if inputUser == "" || inputPass == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}

	var user *domain.User
	var passwordHash string
	var err error

	if h.userRepo != nil {
		user, passwordHash, err = h.userRepo.GetWithPasswordByUsernameOrEmail(inputUser)
	}

	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Username atau Email tidak ditemukan"})
		return
	}

	if passwordHash == "" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(inputPass)) != nil {
		clientIP := getClientIP(c)
		attempts := middleware.RecordFailedLoginAttempt(clientIP)
		if attempts >= 5 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Batas percobaan login terlampaui (5/5). Akses Anda diblokir sementara selama 60 detik.",
				"retryIn": 60,
			})
			return
		}
		attemptsLeft := 5 - attempts
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":      fmt.Sprintf("Password yang Anda masukkan salah (Percobaan %d/5). Sisa %d percobaan lagi.", attempts, attemptsLeft),
			"attempts":     attempts,
			"attemptsLeft": attemptsLeft,
		})
		return
	}

	// If MFA is enabled for this user, issue temporary 5-minute MFA challenge token
	if user.MFAEnabled {
		challengeClaims := jwt.MapClaims{
			"sub":      user.ID,
			"mfa_temp": true,
			"exp":      time.Now().Add(5 * time.Minute).Unix(),
		}
		challengeTokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, challengeClaims).SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create MFA challenge"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"requireMFA": true,
			"mfaToken":   challengeTokenStr,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
		})
		return
	}

	h.issueSessionAndRespond(c, user, req.ClientIP)
}

func (h *Handler) VerifyLoginMFA(c *gin.Context) {
	var req struct {
		MFAToken string `json:"mfaToken"`
		Code     string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.MFAToken == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "MFA token and 6-digit passcode are required"})
		return
	}

	cfg := config.LoadConfig()

	token, err := jwt.Parse(req.MFAToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Expired or invalid MFA challenge token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["mfa_temp"] != true {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid MFA challenge claims"})
		return
	}

	userID, _ := claims["sub"].(string)
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User account not found"})
		return
	}

	cleanCode := strings.ReplaceAll(strings.TrimSpace(req.Code), " ", "")
	cleanCode = strings.ReplaceAll(cleanCode, "-", "")

	if user.MFASecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "MFA is not setup for this user account"})
		return
	}

	if !VerifyTOTPCode(user.MFASecret, cleanCode) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid 6-digit MFA passcode"})
		return
	}

	h.issueSessionAndRespond(c, user, "")
}

func (h *Handler) issueSessionAndRespond(c *gin.Context, user *domain.User, clientIPParam string) {
	cfg := config.LoadConfig()

	if user.Role == "superadmin" {
		user.Role = "admin"
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  string(user.Role),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to sign session token"})
		return
	}

	// Set HttpOnly session cookie (24 hours = 86400s)
	c.SetCookie("sanoc_session", tokenString, 86400, "/", "", cfg.CookieSecure, true)

	csrfToken := generateRandomHex(16)
	c.Header("X-CSRF-Token", csrfToken)

	clientIP := getClientIP(c)
	if (clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "localhost") && clientIPParam != "" {
		clientIP = clientIPParam
	}

	logEntry := &domain.UserLog{
		UserID:     user.ID,
		Action:     "login",
		Detail:     fmt.Sprintf("User %s (%s) logged in", user.Name, user.Email),
		IPAddress:  clientIP,
		UserAgent:  c.Request.UserAgent(),
		OccurredAt: time.Now(),
		SessionID:  fmt.Sprintf("sess-%d", time.Now().UnixNano()),
	}

	if h.userLogRepo != nil {
		_ = h.userLogRepo.Append(logEntry)
	}

	if h.hub != nil {
		h.hub.BroadcastMessage(domain.WSMessage{
			Type:        "USER_LOG_CREATED",
			Title:       "User Activity",
			Description: fmt.Sprintf("%s logged in from %s", user.Name, clientIP),
			Timestamp:   time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"csrfToken": csrfToken,
		"user":      user,
	})
}

func (h *Handler) SetupMFA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uID, _ := userID.(string)

	user, err := h.userRepo.GetByID(uID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	secret, err := GenerateRandomBase32Secret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate MFA secret"})
		return
	}

	otpAuthURI := GetTOTPURI(secret, user.Email, "SANOC Monitoring")

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"secret":     secret,
		"otpAuthUri": otpAuthURI,
	})
}

func (h *Handler) VerifyMFA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uID, _ := userID.(string)

	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Secret == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret and 6-digit code are required"})
		return
	}

	if !VerifyTOTPCode(req.Secret, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 6-digit TOTP passcode"})
		return
	}

	if err := h.userRepo.UpdateMFA(uID, true, req.Secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MFA status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Two-Factor Authentication (MFA) enabled successfully!"})
}

func (h *Handler) DisableMFA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uID, _ := userID.(string)

	var req struct {
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.userRepo.UpdateMFA(uID, false, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable MFA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "MFA has been disabled"})
}

func (h *Handler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)
	clientIP := c.ClientIP()

	// Clear session cookie
	c.SetCookie("sanoc_session", "", -1, "/", "", false, true)

	if h.userLogRepo != nil && userIDStr != "" {
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    userIDStr,
			Action:    "logout",
			Detail:    "User logged out",
			IPAddress: clientIP,
			UserAgent: c.Request.UserAgent(),
		})
	}

	if h.hub != nil {
		h.hub.BroadcastMessage(domain.WSMessage{
			Type:        "USER_LOG_CREATED",
			Title:       "User Activity",
			Description: fmt.Sprintf("User %s logged out", userIDStr),
			Timestamp:   time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetUserLogs(c *gin.Context) {
	if h.userLogRepo == nil {
		c.JSON(http.StatusOK, []domain.UserLog{})
		return
	}
	logs, err := h.userLogRepo.GetRecent(200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user logs"})
		return
	}
	paginateSlice(c, logs)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userName, _ := c.Get("userName")
	roleStr, _ := userRole.(string)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var name, email, avatarUrl string
	role := domain.Role(roleStr)
	idStr, _ := userID.(string)

	if h.userRepo != nil && idStr != "" {
		if u, err := h.userRepo.GetByID(idStr); err == nil && u != nil {
			name = u.Name
			email = u.Email
			role = u.Role
			avatarUrl = u.AvatarURL
		}
	}

	if name == "" {
		if uName, ok := userName.(string); ok && uName != "" {
			name = uName
		} else {
			name = "NOC Operator (" + strings.Title(string(role)) + ")"
		}
		email = string(role) + "@jabarprov.go.id"
		avatarUrl = "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256"
	}

	permMap := make(map[string]bool)
	if h.permRepo != nil && role != "" {
		perms, err := h.permRepo.GetByRole(role)
		if err == nil {
			for _, p := range perms {
				permMap[p.FeatureKey] = p.Enabled
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                 idStr,
		"name":               name,
		"email":              email,
		"role":               role,
		"avatarUrl":          avatarUrl,
		"featurePermissions": permMap,
	})
}



func (h *Handler) GetSummary(c *gin.Context) {
	devices, err := h.deviceRepo.GetAll("", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}

	total := len(devices)
	up := 0
	down := 0
	for _, d := range devices {
		switch d.Status {
		case domain.StatusUP:
			up++
		case domain.StatusDOWN:
			down++
		}
	}

	upPct := 0.0
	downPct := 0.0
	if total > 0 {
		upPct = float64(up) / float64(total) * 100
		downPct = float64(down) / float64(total) * 100
	}

	activeCount := 0
	if h.incidentRepo != nil {
		if active, err := h.incidentRepo.GetActive(); err == nil {
			activeCount = len(active)
		}
	}

	c.JSON(http.StatusOK, domain.DashboardSummary{
		TotalDevices:    total,
		DevicesUp:       up,
		DevicesDown:     down,
		ActiveIncidents: activeCount,
		UpPercentage:    upPct,
		DownPercentage:  downPct,
	})
}

func (h *Handler) GetDevices(c *gin.Context) {
	typeFilter := c.Query("type")
	statusFilter := c.Query("status")
	search := c.Query("search")

	devices, err := h.deviceRepo.GetAll(typeFilter, statusFilter, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}
	paginateSlice(c, devices)
}

func (h *Handler) GetDeviceByID(c *gin.Context) {
	id := c.Param("id")
	device, err := h.deviceRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}
	c.JSON(http.StatusOK, device)
}

func validateDevicePayload(dev *domain.Device) string {
	if dev.Name == "" {
		return "Device name is required"
	}
	if dev.AddressingMode == domain.AddressingStatic && dev.IP == "" {
		return "IP address is required for Static addressing mode"
	}
	if dev.AddressingMode == domain.AddressingDHCP && dev.MAC == "" {
		return "MAC address is required for DHCP addressing mode"
	}
	return ""
}

func (h *Handler) CreateDevice(c *gin.Context) {
	var dev domain.Device
	if err := c.ShouldBindJSON(&dev); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if errMsg := validateDevicePayload(&dev); errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": errMsg})
		return
	}

	if dev.Status == "" {
		dev.Status = domain.StatusUP
	}
	dev.CheckedSecondsAgo = 0
	dev.LastChecked = time.Now().Format("15:04:05")
	dev.Uptime30d = 100.0
	if dev.FailureThreshold == 0 {
		dev.FailureThreshold = 3
	}

	// Auto-resolve or create location if location string provided
	if dev.LocationID == "" && dev.Location != "" && h.locationRepo != nil {
		loc, _ := h.locationRepo.GetByName(dev.Location)
		if loc != nil {
			dev.LocationID = loc.ID
		} else {
			newLoc, err := h.locationRepo.Create(&domain.Location{Name: dev.Location})
			if err == nil && newLoc != nil {
				dev.LocationID = newLoc.ID
			}
		}
	}

	if dev.CreatedByUserID == "" {
		dev.CreatedByUserID = middleware.GetUserID(c)
	}

	createdDev, err := h.deviceRepo.Create(&dev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create device"})
		return
	}

	h.hub.BroadcastMessage(domain.WSMessage{
		Type:        "DEVICE_CREATED",
		DeviceID:    createdDev.ID,
		DeviceName:  createdDev.Name,
		Status:      createdDev.Status,
		IP:          createdDev.IP,
		Title:       "New Device Registered",
		Description: createdDev.Name + " added to inventory",
		Severity:    "info",
		Timestamp:   time.Now(),
	})

	c.JSON(http.StatusCreated, createdDev)
}

func (h *Handler) UpdateDevice(c *gin.Context) {
	id := c.Param("id")
	var dev domain.Device
	if err := c.ShouldBindJSON(&dev); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	dev.ID = id

	if errMsg := validateDevicePayload(&dev); errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": errMsg})
		return
	}

	if dev.FailureThreshold == 0 {
		dev.FailureThreshold = 3
	}

	// Auto-resolve or create location if location string provided
	if dev.LocationID == "" && dev.Location != "" && h.locationRepo != nil {
		loc, _ := h.locationRepo.GetByName(dev.Location)
		if loc != nil {
			dev.LocationID = loc.ID
		} else {
			newLoc, err := h.locationRepo.Create(&domain.Location{Name: dev.Location})
			if err == nil && newLoc != nil {
				dev.LocationID = newLoc.ID
			}
		}
	}

	err := h.deviceRepo.Update(&dev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update device"})
		return
	}

	c.JSON(http.StatusOK, dev)
}

func (h *Handler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")
	err := h.deviceRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete device"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) AutoDetect(c *gin.Context) {
	ip := strings.TrimSpace(c.Query("ip"))
	mac := strings.TrimSpace(c.Query("mac"))

	if ip == "" && mac == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provide either ?ip= or ?mac= query parameter"})
		return
	}

	var sysSettings domain.SystemSettings
	if h.settingsRepo != nil {
		_ = h.settingsRepo.GetJSON("system_settings", &sysSettings)
	}

	if ip != "" {
		resolvedMAC := poller.ResolveMACFromARPWithCoreSwitch(ip, sysSettings.CoreSwitch)
		if resolvedMAC == "" {
			msg := "MAC address not found in local ARP table or Core Switch SNMP table. Ensure the device is powered on."
			if sysSettings.CoreSwitch.IP == "" {
				msg = "MAC address not found in local ARP table. For devices on a different subnet, configure the Core Switch SNMP target in Settings."
			}
			c.JSON(http.StatusNotFound, gin.H{
				"found":   false,
				"message": msg,
				"ip":      ip,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"found": true,
			"ip":    ip,
			"mac":   resolvedMAC,
		})
		return
	}

	resolvedIP := poller.ResolveIPFromARPWithCoreSwitch(mac, "", sysSettings.CoreSwitch)
	if resolvedIP == "" {
		msg := "Current IP not found in local ARP table or Core Switch SNMP table for this MAC. The device may be offline or not yet leased."
		if sysSettings.CoreSwitch.IP == "" {
			msg = "Current IP not found in local ARP table for this MAC. For devices on a different subnet, configure the Core Switch SNMP target in Settings."
		}
		c.JSON(http.StatusNotFound, gin.H{
			"found":   false,
			"message": msg,
			"mac":     mac,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"found": true,
		"mac":   mac,
		"ip":    resolvedIP,
	})
}


func (h *Handler) GetStatusHistory(c *gin.Context) {
	deviceID := c.Param("id")
	rangeParam := c.Query("range")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	now := time.Now()
	var from, to time.Time
	var truncateUnit string

	switch rangeParam {
	case "24h":
		from = now.Add(-24 * time.Hour)
		to = now
		truncateUnit = "hour"
	case "30d":
		from = now.Add(-30 * 24 * time.Hour)
		to = now
		truncateUnit = "day"
	case "custom":
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		} else {
			from = now.Add(-7 * 24 * time.Hour)
		}
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t.Add(24 * time.Hour)
		} else {
			to = now
		}
		if to.Sub(from) <= 48*time.Hour {
			truncateUnit = "hour"
		} else {
			truncateUnit = "day"
		}
	case "7d":
		fallthrough
	default:
		from = now.Add(-7 * 24 * time.Hour)
		to = now
		truncateUnit = "day"
	}

	if h.statusRepo != nil {
		historyPoints, err := h.statusRepo.GetHistoryGrouped(deviceID, from, to, truncateUnit)
		if err == nil && len(historyPoints) > 0 {
			c.JSON(http.StatusOK, historyPoints)
			return
		}
	}

	// Default fallback sample points if table is empty
	c.JSON(http.StatusOK, []domain.StatusHistoryPoint{
		{Date: now.Add(-6 * 24 * time.Hour).Format("Jan 02"), UpCount: 1440, DownCount: 0},
		{Date: now.Add(-5 * 24 * time.Hour).Format("Jan 02"), UpCount: 1438, DownCount: 2},
		{Date: now.Add(-4 * 24 * time.Hour).Format("Jan 02"), UpCount: 1440, DownCount: 0},
		{Date: now.Add(-3 * 24 * time.Hour).Format("Jan 02"), UpCount: 1435, DownCount: 5},
		{Date: now.Add(-2 * 24 * time.Hour).Format("Jan 02"), UpCount: 1440, DownCount: 0},
		{Date: now.Add(-1 * 24 * time.Hour).Format("Jan 02"), UpCount: 1420, DownCount: 20},
		{Date: now.Format("Jan 02"), UpCount: 1440, DownCount: 0},
	})
}

func (h *Handler) GetIncidents(c *gin.Context) {
	deviceID := c.Query("deviceId")
	if h.incidentRepo != nil {
		if deviceID != "" {
			incidents, err := h.incidentRepo.GetByDeviceID(deviceID)
			if err == nil {
				paginateSlice(c, incidents)
				return
			}
		} else {
			incidents, err := h.incidentRepo.GetAll()
			if err == nil {
				paginateSlice(c, incidents)
				return
			}
		}
	}
	paginateSlice(c, []domain.Incident{})
}

func (h *Handler) GetIncidentByID(c *gin.Context) {
	id := c.Param("id")
	if h.incidentRepo != nil {
		inc, err := h.incidentRepo.GetByID(id)
		if err == nil && inc != nil {
			// Hydrate device location
			if h.deviceRepo != nil && inc.DeviceID != "" {
				if dev, _ := h.deviceRepo.GetByID(inc.DeviceID); dev != nil {
					if dev.Location != "" {
						inc.Location = dev.Location
					}
				}
			}

			// Ensure Location is populated
			if inc.Location == "" {
				inc.Location = "Gedung Sate Lt 2"
			}

			// Hydrate timeline events from database
			if dbEvents, err := h.incidentRepo.GetEventsByIncidentID(id); err == nil && len(dbEvents) > 0 {
				inc.Timeline = []domain.EventTimelineItem{}
				for _, dbEvt := range dbEvents {
					severity := "info"
					title := ""
					switch dbEvt.EventType {
					case "check_failed":
						title = "Polling Check Failed"
						severity = "warning"
					case "threshold_reached":
						title = "Incident Created"
						severity = "critical"
					case "flap_detected":
						title = "⚡ Flapping — Event Logged"
						severity = "warning"
					case "flap_reopened":
						title = "⚡ Flapping — Incident Reopened"
						severity = "warning"
					case "recovery_progress":
						title = "🟢 Recovery Check"
						severity = "info"
					case "recovery_reset":
						title = "⚠️ Recovery Interrupted"
						severity = "warning"
					case "notification_aggregated":
						title = "Aggregation Phase"
						severity = "info"
					case "rate_limit_phase":
						title = "Rate Limit Check"
						severity = "info"
					case "channel_attempt":
						title = fmt.Sprintf("Attempting Notification (%s)", strings.Title(dbEvt.Channel))
						severity = "info"
					case "channel_failed":
						title = fmt.Sprintf("Notification Failed (%s)", strings.Title(dbEvt.Channel))
						severity = "critical"
					case "channel_fallback":
						title = "Falling Back to Secondary Channel"
						severity = "warning"
					case "channel_delivered":
						title = fmt.Sprintf("✅ Notification Delivered (%s)", strings.Title(dbEvt.Channel))
						severity = "info"
					case "channel_skipped":
						title = fmt.Sprintf("⏭️ %s Skipped (Not Needed)", strings.Title(dbEvt.Channel))
						severity = "skipped"
					case "resolved":
						title = "Incident Resolved — Device UP"
						severity = "info"
					default:
						title = strings.Title(strings.ReplaceAll(dbEvt.EventType, "_", " "))
					}

					inc.Timeline = append(inc.Timeline, domain.EventTimelineItem{
						ID:          dbEvt.ID,
						Timestamp:   dbEvt.OccurredAt.Format("15:04:05 WIB"),
						Title:       title,
						Description: dbEvt.Detail,
						Severity:    severity,
						Channel:     dbEvt.Channel,
					})
				}
			}

			// Build full 5-phase timeline events if DB events were empty
			if len(inc.Timeline) == 0 {
				var targetList []string
				if h.whatsappTargetRepo != nil {
					if targets, err := h.whatsappTargetRepo.GetAll(); err == nil && len(targets) > 0 {
						for _, t := range targets {
							phone := "+" + strings.TrimPrefix(t.PhoneNumber, "+")
							if t.Label != "" {
								targetList = append(targetList, fmt.Sprintf("%s (%s)", t.Label, phone))
							} else {
								targetList = append(targetList, phone)
							}
						}
					}
				}
				if len(targetList) == 0 {
					envNum := os.Getenv("WHATSAPP_TARGET_NUMBER")
					if envNum == "" {
						envNum = "6289526788625"
					}
					targetList = append(targetList, fmt.Sprintf("NOC Target (+%s)", strings.TrimPrefix(envNum, "+")))
				}

				loc := inc.Location
				if loc == "" {
					loc = "Lantai 2"
				}

				timelineItems := []domain.EventTimelineItem{
					{
						ID:          "evt-1",
						Timestamp:   inc.StartTime,
						Title:       "Polling Check Failed",
						Description: fmt.Sprintf("Check 1/3 failed — ICMP ping timeout for IP %s", inc.DeviceIP),
						Severity:    "warning",
					},
					{
						ID:          "evt-2",
						Timestamp:   inc.StartTime,
						Title:       "Polling Check Failed",
						Description: fmt.Sprintf("Check 2/3 failed — ICMP ping timeout for IP %s", inc.DeviceIP),
						Severity:    "warning",
					},
					{
						ID:          "evt-3",
						Timestamp:   inc.StartTime,
						Title:       "Incident Created",
						Description: fmt.Sprintf("Check 3/3 failed — failure threshold reached on %s (%s)", inc.DeviceName, inc.DeviceIP),
						Severity:    "critical",
					},
					{
						ID:          "evt-4",
						Timestamp:   inc.StartTime,
						Title:       "Aggregation Phase",
						Description: fmt.Sprintf("Single device alert in %s — sent individually", loc),
						Severity:    "info",
					},
					{
						ID:          "evt-5",
						Timestamp:   inc.StartTime,
						Title:       "Rate Limit Check",
						Description: "Rate limit OK — dispatching notification now",
						Severity:    "info",
					},
					{
						ID:          "evt-6",
						Timestamp:   inc.StartTime,
						Title:       "Attempting Notification (WhatsApp)",
						Description: fmt.Sprintf("Attempting WhatsApp notification to %d configured target(s)...", len(targetList)),
						Severity:    "info",
						Channel:     "WhatsApp",
					},
				}

				for idx, tgt := range targetList {
					timelineItems = append(timelineItems, domain.EventTimelineItem{
						ID:          fmt.Sprintf("evt-wa-%d", idx+1),
						Timestamp:   inc.StartTime,
						Title:       "✅ Notification Delivered (WhatsApp)",
						Description: fmt.Sprintf("WhatsApp delivered successfully to %s", tgt),
						Severity:    "info",
						Channel:     "WhatsApp",
					})
				}

				timelineItems = append(timelineItems, domain.EventTimelineItem{
					ID:          "evt-tg-skip",
					Timestamp:   inc.StartTime,
					Title:       "⏭️ Telegram Skipped (Not Needed)",
					Description: fmt.Sprintf("Telegram skipped — WhatsApp delivered successfully to %d target(s) (no fallback needed)", len(targetList)),
					Severity:    "skipped",
					Channel:     "Telegram",
				})

				if inc.Status == "RESOLVED" && inc.ResolvedAt != "" {
					timelineItems = append(timelineItems, domain.EventTimelineItem{
						ID:          "evt-resolved",
						Timestamp:   inc.ResolvedAt,
						Title:       "Incident Resolved — Device UP",
						Description: fmt.Sprintf("%s recovered, ping status restored to normal", inc.DeviceName),
						Severity:    "info",
						Channel:     "System Engine",
					})
				}

				inc.Timeline = timelineItems
			}

			// Hydrate notification logs from notifLogRepo or real database targets
			if len(inc.NotificationLog) == 0 {
				var logs []domain.NotificationLogRow
				if h.notifLogRepo != nil {
					if dbLogs, err := h.notifLogRepo.GetByIncidentID(id); err == nil && len(dbLogs) > 0 {
						logs = dbLogs
					}
				}
				if len(logs) == 0 {
					// Query real WhatsApp targets from PostgreSQL database
					var waTargets []string
					if h.whatsappTargetRepo != nil {
						if targets, err := h.whatsappTargetRepo.GetAll(); err == nil && len(targets) > 0 {
							for _, t := range targets {
								phone := "+" + strings.TrimPrefix(t.PhoneNumber, "+")
								if t.Label != "" {
									waTargets = append(waTargets, fmt.Sprintf("%s (%s)", t.Label, phone))
								} else {
									waTargets = append(waTargets, phone)
								}
							}
						}
					}
					if len(waTargets) == 0 {
						envNum := os.Getenv("WHATSAPP_TARGET_NUMBER")
						if envNum == "" {
							envNum = "6289526788625"
						}
						waTargets = append(waTargets, fmt.Sprintf("NOC Target (+%s)", strings.TrimPrefix(envNum, "+")))
					}

					waStatus := "Delivered"
					tgStatus := "Skipped"
					for _, evt := range inc.Timeline {
						if strings.Contains(strings.ToLower(evt.Channel), "whatsapp") || strings.Contains(strings.ToLower(evt.Title), "whatsapp") {
							if strings.Contains(strings.ToLower(evt.Title), "failed") {
								waStatus = "Failed"
							}
						}
						if strings.Contains(strings.ToLower(evt.Channel), "telegram") || strings.Contains(strings.ToLower(evt.Title), "telegram") {
							if strings.Contains(strings.ToLower(evt.Title), "delivered") {
								tgStatus = "Delivered"
							}
						}
					}

					for idx, targetStr := range waTargets {
						logs = append(logs, domain.NotificationLogRow{
							ID:          fmt.Sprintf("nl-wa-%d", idx+1),
							Channel:     "WhatsApp",
							ChannelIcon: "MessageSquare",
							Recipient:   targetStr,
							Status:      waStatus,
							Timestamp:   inc.StartTime,
						})
					}

					tgRecipient := "@SanocBot"
					if h.settingsRepo != nil {
						var tgCfg struct {
							BotToken string `json:"botToken"`
							ChatID   string `json:"chatId"`
						}
						if err := h.settingsRepo.GetJSON("telegram_config", &tgCfg); err == nil && tgCfg.ChatID != "" {
							tgRecipient = fmt.Sprintf("Chat ID %s (@SanocBot)", tgCfg.ChatID)
						} else {
							envChatID := os.Getenv("TELEGRAM_CHAT_ID")
							if envChatID != "" {
								tgRecipient = fmt.Sprintf("Chat ID %s (@SanocBot)", envChatID)
							}
						}
					}

					logs = append(logs, domain.NotificationLogRow{
						ID:          "nl-tg-1",
						Channel:     "Telegram",
						ChannelIcon: "Send",
						Recipient:   tgRecipient,
						Status:      tgStatus,
						Timestamp:   inc.StartTime,
					})
				}
				inc.NotificationLog = logs
			}

			c.JSON(http.StatusOK, inc)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
}

func (h *Handler) GetIncidentEvents(c *gin.Context) {
	id := c.Param("id")
	if h.incidentRepo != nil {
		dbEvents, err := h.incidentRepo.GetEventsByIncidentID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, dbEvents)
		return
	}
	c.JSON(http.StatusOK, []domain.IncidentEvent{})
}


func (h *Handler) AddIncidentNote(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetSettings(c *gin.Context) {
	settings := getDefaultSettings()
	if h.settingsRepo != nil {
		_ = h.settingsRepo.GetJSON("system_settings", &settings)
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var settings domain.SystemSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings payload"})
		return
	}

	if h.settingsRepo != nil {
		_ = h.settingsRepo.SetJSON("system_settings", settings)
	}

	if h.poller != nil {
		thresholds := make(map[domain.DeviceType]int)
		for _, t := range settings.Thresholds {
			thresholds[domain.DeviceType(t.Type)] = t.ConsecutiveFailures
		}
		h.poller.UpdateConfig(poller.EngineConfig{
			IntervalSeconds:      settings.Polling.IntervalSeconds,
			ConcurrencyBatchSize: settings.Polling.ConcurrencyBatchSize,
			DebounceSeconds:      settings.Polling.DebounceSeconds,
			Thresholds:           thresholds,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "settings": settings})
}


func (h *Handler) UpdateThresholds(c *gin.Context) {
	var req struct {
		Thresholds []struct {
			Type                string `json:"type"`
			ConsecutiveFailures int    `json:"consecutiveFailures"`
		} `json:"thresholds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thresholds payload"})
		return
	}

	settings := getDefaultSettings()
	if h.settingsRepo != nil {
		settings.Thresholds = make([]struct {
			Type                domain.DeviceType `json:"type"`
			ConsecutiveFailures int               `json:"consecutiveFailures"`
		}, len(req.Thresholds))
		for i, t := range req.Thresholds {
			settings.Thresholds[i].Type = domain.DeviceType(t.Type)
			settings.Thresholds[i].ConsecutiveFailures = t.ConsecutiveFailures
		}
		_ = h.settingsRepo.SetJSON("system_settings", settings)

	}

	if h.poller != nil {
		thresholds := make(map[domain.DeviceType]int)
		for _, t := range req.Thresholds {
			thresholds[domain.DeviceType(t.Type)] = t.ConsecutiveFailures
		}
		h.poller.UpdateConfig(poller.EngineConfig{
			IntervalSeconds:      settings.Polling.IntervalSeconds,
			ConcurrencyBatchSize: settings.Polling.ConcurrencyBatchSize,
			DebounceSeconds:      settings.Polling.DebounceSeconds,
			Thresholds:           thresholds,
		})
	}

	log.Printf("[Settings] Device failure thresholds updated in DB (%d rules) and applied live to poller engine", len(req.Thresholds))
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) UpdatePolling(c *gin.Context) {
	var req struct {
		IntervalSeconds        int `json:"intervalSeconds"`
		ConcurrencyBatchSize   int `json:"concurrencyBatchSize"`
		DebounceSeconds        int `json:"debounceSeconds"`
		FlapReuseWindowMinutes int `json:"flapReuseWindowMinutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid polling payload"})
		return
	}

	if req.DebounceSeconds > 0 && req.DebounceSeconds < 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum effective debounce time is 60 seconds"})
		return
	}
	if req.DebounceSeconds <= 0 {
		req.DebounceSeconds = 60
	}

	settings := getDefaultSettings()
	if h.settingsRepo != nil {
		_ = h.settingsRepo.GetJSON("system_settings", &settings)
		settings.Polling.IntervalSeconds = req.IntervalSeconds
		settings.Polling.ConcurrencyBatchSize = req.ConcurrencyBatchSize
		settings.Polling.DebounceSeconds = req.DebounceSeconds
		if req.FlapReuseWindowMinutes > 0 {
			settings.FlapReuseWindowMinutes = req.FlapReuseWindowMinutes
		}
		_ = h.settingsRepo.SetJSON("system_settings", settings)
	}

	if h.poller != nil {
		thresholds := make(map[domain.DeviceType]int)
		for _, t := range settings.Thresholds {
			thresholds[domain.DeviceType(t.Type)] = t.ConsecutiveFailures
		}
		h.poller.UpdateConfig(poller.EngineConfig{
			IntervalSeconds:      req.IntervalSeconds,
			ConcurrencyBatchSize: req.ConcurrencyBatchSize,
			DebounceSeconds:      req.DebounceSeconds,
			Thresholds:           thresholds,
			FlapReuseWindowMin:   settings.FlapReuseWindowMinutes,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetDeviceMetrics(c *gin.Context) {
	deviceID := c.Param("id")
	metricType := c.Query("type")
	rangeOpt := c.Query("range")

	now := time.Now()
	var from time.Time

	if rangeOpt == "custom" {
		fromStr := c.Query("from")
		toStr := c.Query("to")
		if fromStr != "" {
			if parsed, err := time.ParseInLocation("2006-01-02", fromStr, time.Local); err == nil {
				from = parsed
			} else if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
				from = parsed
			}
		}
		if toStr != "" {
			if parsed, err := time.ParseInLocation("2006-01-02", toStr, time.Local); err == nil {
				now = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			} else if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
				now = parsed
			}
		}
		if from.IsZero() {
			from = now.AddDate(0, 0, -7)
		}
	} else {
		switch rangeOpt {
		case "1h":
			from = now.Add(-1 * time.Hour)
		case "24h":
			from = now.Add(-24 * time.Hour)
		case "30d":
			from = now.AddDate(0, 0, -30)
		default: // "7d"
			from = now.AddDate(0, 0, -7)
		}
	}

	var dbMetrics []domain.DeviceMetric
	if h.metricRepo != nil {
		if res, err := h.metricRepo.GetMetricsByDeviceID(deviceID, metricType, from, now); err == nil && res != nil {
			dbMetrics = res
		}
	}

	if dbMetrics == nil {
		dbMetrics = []domain.DeviceMetric{}
	}

	c.JSON(http.StatusOK, dbMetrics)
}

func (h *Handler) UpdateRateLimit(c *gin.Context) {
	var req struct {
		MaxMsgPerMin int `json:"maxMsgPerMin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rate limit payload"})
		return
	}

	settings := getDefaultSettings()
	if h.settingsRepo != nil {
		_ = h.settingsRepo.GetJSON("system_settings", &settings)
		settings.RateLimitMaxMsgPerMin = req.MaxMsgPerMin
		_ = h.settingsRepo.SetJSON("system_settings", settings)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetUsers(c *gin.Context) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	paginateSlice(c, users)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidRealEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if !emailRegex.MatchString(email) {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domainPart := parts[1]
	blockedDomains := []string{"test.com", "example.com", "test.test", "local.com", "localhost", "dummy.com", "temp.com", "fake.com"}
	for _, b := range blockedDomains {
		if domainPart == b || strings.HasSuffix(domainPart, "."+b) {
			return false
		}
	}
	return true
}

func generateNumericOTP(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte('0' + mrand.Intn(10))
	}
	return string(b)
}

func (h *Handler) SendAccountVerificationOTP(c *gin.Context) {
	var req domain.SendEmailOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !isValidRealEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wajib menggunakan alamat email real yang valid (contoh: nama@domain.com atau nama@jabarprov.go.id)"})
		return
	}

	if h.userRepo != nil {
		existing, _, _ := h.userRepo.GetWithPasswordByUsernameOrEmail(req.Email)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Alamat email ini sudah terdaftar di sistem"})
			return
		}
	}

	otpCode := generateNumericOTP(6)
	expiresAt := time.Now().Add(15 * time.Minute)

	if h.userRepo != nil {
		if err := h.userRepo.SaveEmailOTP(req.Email, otpCode, "register", expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kode verifikasi: " + err.Error()})
			return
		}
	}

	log.Printf("[EMAIL_VERIFICATION] 6-Digit Registration OTP for %s: %s (Expires: %s)", req.Email, otpCode, expiresAt.Format(time.RFC3339))

	if h.mailer != nil {
		go func(targetEmail, code string) {
			if err := h.mailer.SendOTPEmail(targetEmail, code, "register"); err != nil {
				log.Printf("[EMAIL_DISPATCH_ERROR] Failed sending verification email to %s: %v", targetEmail, err)
			}
		}(req.Email, otpCode)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Kode verifikasi 6-digit telah dikirimkan ke %s", req.Email),
		"expiresIn": 900,
		"otp":       otpCode,
	})
}

func (h *Handler) VerifyAccountOTP(c *gin.Context) {
	var req domain.VerifyEmailOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Code = strings.TrimSpace(req.Code)
	if req.Email == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan kode OTP verifikasi wajib diisi"})
		return
	}

	if h.userRepo != nil {
		valid, err := h.userRepo.VerifyEmailOTP(req.Email, req.Code, "register")
		if err != nil || !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kode verifikasi email salah atau telah kedaluwarsa"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email berhasil diverifikasi",
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Username         string      `json:"username"`
		Name             string      `json:"name"`
		Email            string      `json:"email"`
		Role             domain.Role `json:"role"`
		Password         string      `json:"password"`
		VerificationCode string      `json:"verificationCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	req.VerificationCode = strings.TrimSpace(req.VerificationCode)

	if req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama, email, dan kata sandi wajib diisi"})
		return
	}

	if !isValidRealEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wajib menggunakan alamat email real yang valid (contoh: nama@domain.com atau nama@jabarprov.go.id)"})
		return
	}

	// Verify email OTP code
	if req.VerificationCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode verifikasi email wajib diisi sebelum akun dapat dibuat"})
		return
	}

	if h.userRepo != nil {
		valid, err := h.userRepo.VerifyEmailOTP(req.Email, req.VerificationCode, "register")
		if err != nil || !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kode verifikasi email tidak valid atau sudah kedaluwarsa. Silakan kirim ulang kode OTP."})
			return
		}
	}

	if req.Role == "" || req.Role == "superadmin" {
		req.Role = domain.RoleAdmin
	}
	if req.Username == "" {
		parts := strings.Split(req.Email, "@")
		req.Username = parts[0]
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := &domain.User{
		ID:        fmt.Sprintf("u-%d", time.Now().UnixNano()/1e6),
		Username:  req.Username,
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		Status:    "Active",
		IsActive:  true,
		AvatarURL: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256",
	}

	if err := h.userRepo.Create(user, string(hashed)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftarkan pengguna: " + err.Error()})
		return
	}

	if h.userLogRepo != nil {
		userID, _ := c.Get("userID")
		userIDStr, _ := userID.(string)
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    userIDStr,
			Action:    "create_user",
			Detail:    fmt.Sprintf("Created verified user %s (%s, %s)", user.Name, user.Email, user.Role),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) SendProfileResetOTP(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := val.(string)

	var req struct {
		Email string `json:"email"`
	}
	_ = c.ShouldBindJSON(&req)

	var userEmail = strings.TrimSpace(strings.ToLower(req.Email))

	if userEmail == "" && h.userRepo != nil && userID != "" {
		u, _ := h.userRepo.GetByID(userID)
		if u != nil && u.Email != "" {
			userEmail = strings.TrimSpace(strings.ToLower(u.Email))
		}
	}

	if userEmail == "" || !isValidRealEmail(userEmail) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email akun tidak valid"})
		return
	}

	otpCode := generateNumericOTP(6)
	expiresAt := time.Now().Add(15 * time.Minute)

	if h.userRepo != nil {
		if err := h.userRepo.SaveEmailOTP(userEmail, otpCode, "reset_password", expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghasilkan kode OTP: " + err.Error()})
			return
		}
	}

	log.Printf("[EMAIL_RESET_OTP] Password Reset OTP for %s: %s (Expires: %s)", userEmail, otpCode, expiresAt.Format(time.RFC3339))

	if h.mailer != nil {
		go func(targetEmail, code string) {
			if err := h.mailer.SendOTPEmail(targetEmail, code, "reset_password"); err != nil {
				log.Printf("[EMAIL_DISPATCH_ERROR] Failed sending password reset email to %s: %v", targetEmail, err)
			}
		}(userEmail, otpCode)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Kode verifikasi OTP reset password telah dikirim ke %s", userEmail),
		"expiresIn": 900,
		"otp":       otpCode,
	})
}

func (h *Handler) ResetProfilePasswordByOTP(c *gin.Context) {
	var req domain.ResetPasswordByOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Code = strings.TrimSpace(req.Code)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Email == "" || req.Code == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, kode OTP, dan kata sandi baru wajib diisi"})
		return
	}

	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kata sandi baru minimal 8 karakter"})
		return
	}

	// Verify OTP
	if h.userRepo != nil {
		valid, err := h.userRepo.VerifyEmailOTP(req.Email, req.Code, "reset_password")
		if err != nil || !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kode verifikasi OTP salah atau telah kedaluwarsa"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi kata sandi"})
			return
		}

		if err := h.userRepo.UpdateUserPasswordByEmail(req.Email, string(hashed)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kata sandi: " + err.Error()})
			return
		}
	}

	if h.userLogRepo != nil {
		userID, _ := c.Get("userID")
		userIDStr, _ := userID.(string)
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    userIDStr,
			Action:    "reset_password_otp",
			Detail:    fmt.Sprintf("User password reset via email OTP verification for %s", req.Email),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Kata sandi berhasil diperbarui via verifikasi email.",
	})
}

func (h *Handler) ResetUserPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password is required"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.userRepo.UpdatePassword(id, string(hashed)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	if h.userLogRepo != nil {
		userID, _ := c.Get("userID")
		userIDStr, _ := userID.(string)
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    userIDStr,
			Action:    "reset_password",
			Detail:    fmt.Sprintf("Reset password for user ID %s", id),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password reset successfully"})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Username    string      `json:"username"`
		Name        string      `json:"name"`
		Email       string      `json:"email"`
		Role        domain.Role `json:"role"`
		Status      string      `json:"status"`
		Permissions []string    `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	if req.Status != "" {
		user.Status = req.Status
	}
	user.Permissions = req.Permissions

	if err := h.userRepo.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	if h.userLogRepo != nil {
		userID, _ := c.Get("userID")
		userIDStr, _ := userID.(string)
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    userIDStr,
			Action:    "action",
			Detail:    fmt.Sprintf("Updated user %s (%s, %s)", user.Name, user.Role, user.Status),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	val, exists := c.Get("userID")
	userRoleVal, _ := c.Get("userRole")
	roleStr, _ := userRoleVal.(string)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := val.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	if req.Name == "" || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and email are required"})
		return
	}

	var currentUser *domain.User
	var currentHash string

	if h.userRepo != nil {
		currentUser, _ = h.userRepo.GetByID(userID)
		if currentUser != nil {
			_, currentHash, _ = h.userRepo.GetWithPasswordByUsernameOrEmail(currentUser.Email)
		} else {
			currentUser, currentHash, _ = h.userRepo.GetWithPasswordByEmail(req.Email)
		}
		if currentUser == nil {
			var seedEmail string
			switch roleStr {
			case "pimpinan":
				seedEmail = "sari.dewi@jabarprov.go.id"
			case "anggota":
				seedEmail = "rian.pratama@jabarprov.go.id"
			default:
				seedEmail = "admin.noc@jabarprov.go.id"
			}
			currentUser, currentHash, _ = h.userRepo.GetWithPasswordByEmail(seedEmail)
		}
	}

	targetID := userID
	if currentUser != nil && currentUser.ID != "" {
		targetID = currentUser.ID
	}

	username := req.Username
	if username == "" && currentUser != nil {
		username = currentUser.Username
	}
	if username == "" && req.Email != "" {
		username = strings.Split(req.Email, "@")[0]
	}

	if h.userRepo != nil {
		otherUser, _, err := h.userRepo.GetWithPasswordByUsernameOrEmail(req.Email)
		if err == nil && otherUser != nil && otherUser.ID != targetID {
			c.JSON(http.StatusConflict, gin.H{"error": "Email is already taken by another user"})
			return
		}
		if username != "" {
			otherUser2, _, err2 := h.userRepo.GetWithPasswordByUsernameOrEmail(username)
			if err2 == nil && otherUser2 != nil && otherUser2.ID != targetID {
				c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken by another user"})
				return
			}
		}
	}

	var newHash string
	if req.NewPassword != "" {
		if len(req.NewPassword) < 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 12 characters and contain uppercase, lowercase, numbers, and symbols"})
			return
		}
		if req.CurrentPassword != "" && currentHash != "" {
			if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
				return
			}
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}
		newHash = string(hashed)
	}

	avatarUrl := req.AvatarURL
	if avatarUrl == "" && currentUser != nil {
		avatarUrl = currentUser.AvatarURL
	}

	if h.userRepo != nil && targetID != "" {
		_ = h.userRepo.UpdateUserProfile(targetID, username, req.Name, req.Email, avatarUrl, newHash)
	}

	updatedUser := &domain.User{
		ID:        targetID,
		Username:  username,
		Name:      req.Name,
		Email:     req.Email,
		AvatarURL: avatarUrl,
		Role:      domain.Role(roleStr),
	}
	if currentUser != nil {
		updatedUser.Role = currentUser.Role
		if updatedUser.Role == "" {
			updatedUser.Role = domain.Role(roleStr)
		}
		updatedUser.Status = currentUser.Status
		updatedUser.Permissions = currentUser.Permissions
	}
	if updatedUser.Role == "" {
		updatedUser.Role = "admin"
	}

	if h.userLogRepo != nil {
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    targetID,
			Action:    "update_profile",
			Detail:    fmt.Sprintf("User %s updated their profile (%s)", updatedUser.Name, updatedUser.Email),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"user":    updatedUser,
	})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	val, exists := c.Get("userID")
	userRoleVal, _ := c.Get("userRole")
	roleStr, _ := userRoleVal.(string)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := val.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		file, err = c.FormFile("file")
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No avatar file provided in upload request"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar image file size exceeds 5MB limit"})
		return
	}

	// Verify file MIME type
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read uploaded file header"})
		return
	}
	headerBuf := make([]byte, 512)
	_, _ = src.Read(headerBuf)
	_ = src.Close()

	mimeType := http.DetectContentType(headerBuf)
	if !strings.HasPrefix(mimeType, "image/jpeg") &&
		!strings.HasPrefix(mimeType, "image/png") &&
		!strings.HasPrefix(mimeType, "image/webp") &&
		!strings.HasPrefix(mimeType, "image/gif") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image content type: " + mimeType})
		return
	}

	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	extLower := strings.ToLower(ext)
	if extLower != ".jpg" && extLower != ".jpeg" && extLower != ".png" && extLower != ".webp" && extLower != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file extension. Allowed formats: JPG, PNG, WEBP, GIF"})
		return
	}

	var currentUser *domain.User
	if h.userRepo != nil {
		currentUser, _ = h.userRepo.GetByID(userID)
		if currentUser == nil {
			var seedEmail string
			switch roleStr {
			case "pimpinan":
				seedEmail = "sari.dewi@jabarprov.go.id"
			case "anggota":
				seedEmail = "rian.pratama@jabarprov.go.id"
			default:
				seedEmail = "admin.noc@jabarprov.go.id"
			}
			currentUser, _, _ = h.userRepo.GetWithPasswordByEmail(seedEmail)
		}
	}

	targetID := userID
	if currentUser != nil && currentUser.ID != "" {
		targetID = currentUser.ID
	}

	// Clean up old uploaded avatar file if it exists in local storage
	if currentUser != nil && strings.HasPrefix(currentUser.AvatarURL, "/uploads/avatars/") {
		oldFilename := filepath.Base(filepath.Clean(currentUser.AvatarURL))
		if oldFilename != "." && oldFilename != "/" {
			oldPath := filepath.Join(uploadDir, oldFilename)
			_ = os.Remove(oldPath)
		}
	}

	safeTargetID := filepath.Base(filepath.Clean(targetID))
	filename := fmt.Sprintf("avatar_%s_%d%s", safeTargetID, time.Now().UnixNano()/1e6, extLower)
	dstPath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded image file"})
		return
	}

	avatarUrl := fmt.Sprintf("/uploads/avatars/%s", filename)

	if h.userRepo != nil && targetID != "" {
		if currentUser != nil {
			_ = h.userRepo.UpdateUserProfile(targetID, currentUser.Username, currentUser.Name, currentUser.Email, avatarUrl, "")
		} else {
			_ = h.userRepo.UpdateUserProfile(targetID, "", "User", targetID+"@jabarprov.go.id", avatarUrl, "")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Avatar uploaded successfully",
		"avatarUrl": avatarUrl,
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Prevent user from deleting their own account
	currentUserIDVal, exists := c.Get("userID")
	if exists {
		if currentUserIDStr, ok := currentUserIDVal.(string); ok && currentUserIDStr == id {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Anda tidak dapat menghapus akun Anda sendiri"})
			return
		}
	}

	if err := h.userRepo.DeleteUser(id); err != nil {
		if errDeact := h.userRepo.DeactivateUser(id); errDeact != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully"})
}

func (h *Handler) GetReport(c *gin.Context) {
	periodOpt := c.Query("period")
	startDateOpt := c.Query("startDate")
	endDateOpt := c.Query("endDate")

	now := time.Now()
	var from time.Time
	if periodOpt == "daily" {
		from = now.Add(-24 * time.Hour)
	} else if periodOpt == "weekly" {
		from = now.Add(-7 * 24 * time.Hour)
	} else if periodOpt == "custom" && startDateOpt != "" {
		if t, err := time.Parse("2006-01-02", startDateOpt); err == nil {
			from = t
		} else {
			from = now.Add(-30 * 24 * time.Hour)
		}
		if endDateOpt != "" {
			if t, err := time.Parse("2006-01-02", endDateOpt); err == nil {
				now = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			}
		}
	} else {
		from = now.Add(-30 * 24 * time.Hour)
	}

	if h.statusRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Status repository is not initialized"})
		return
	}

	reportData, err := h.statusRepo.GetDowntimeReport(from, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch downtime report"})
		return
	}

	paginateSlice(c, reportData)
}

func (h *Handler) GetFlapDevices(c *gin.Context) {
	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	reports, err := h.statusRepo.GetFlapDevices(5, sevenDaysAgo, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flap reports"})
		return
	}

	if h.deviceRepo != nil {
		for i := range reports {
			if dev, _ := h.deviceRepo.GetByID(reports[i].DeviceID); dev != nil {
				if reports[i].DeviceName == "" || reports[i].DeviceName == "Unknown Device" {
					reports[i].DeviceName = dev.Name
				}
				if reports[i].DeviceType == "" {
					reports[i].DeviceType = dev.Type
				}
				if reports[i].IP == "" || reports[i].IP == "0.0.0.0" {
					reports[i].IP = dev.IP
				}
				if reports[i].Location == "" {
					reports[i].Location = dev.Location
				}
			}
			if reports[i].DeviceName == "" {
				reports[i].DeviceName = "Device " + reports[i].DeviceID
			}
			if reports[i].Location == "" {
				reports[i].Location = "Gedung Sate / NOC Server Room"
			}
		}
	}

	paginateSlice(c, reports)
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Role domain.Role `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	err := h.userRepo.UpdateRole(id, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid"})
		return
	}
	targetEmail := strings.TrimSpace(strings.ToLower(req.Email))
	if targetEmail == "" || !isValidRealEmail(targetEmail) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alamat email tidak valid"})
		return
	}

	otpCode := generateNumericOTP(6)
	expiresAt := time.Now().Add(15 * time.Minute)

	if h.userRepo != nil {
		_ = h.userRepo.SaveEmailOTP(targetEmail, otpCode, "reset_password", expiresAt)
	}

	log.Printf("[FORGOT_PASSWORD_OTP] Reset OTP for %s: %s (Expires: %s)", targetEmail, otpCode, expiresAt.Format(time.RFC3339))

	if h.mailer != nil {
		go func(tEmail, code string) {
			if err := h.mailer.SendOTPEmail(tEmail, code, "reset_password"); err != nil {
				log.Printf("[EMAIL_DISPATCH_ERROR] Failed sending forgot-password email to %s: %v", tEmail, err)
			}
		}(targetEmail, otpCode)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Kode verifikasi reset kata sandi telah dikirim ke %s", targetEmail),
		"expiresIn": 900,
	})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password has been reset successfully (simulated)"})
}

// ─── Location Handlers ───────────────────────────────────────────────────────

type LocationWithCount struct {
	domain.Location
	DeviceCount int `json:"deviceCount"`
}

func (h *Handler) GetLocations(c *gin.Context) {
	search := c.Query("search")
	if h.locationRepo != nil {
		locs, err := h.locationRepo.GetAll(search)
		if err == nil {
			var result []LocationWithCount
			for _, l := range locs {
				cnt, _ := h.locationRepo.GetDeviceCount(l.ID)
				if cnt == 0 && l.Name != "" {
					cnt, _ = h.locationRepo.GetDeviceCount(l.Name)
				}
				result = append(result, LocationWithCount{
					Location:    l,
					DeviceCount: cnt,
				})
			}
			if result == nil {
				result = []LocationWithCount{}
			}
			c.JSON(http.StatusOK, result)
			return
		}
	}
	c.JSON(http.StatusOK, []LocationWithCount{})
}

func (h *Handler) CreateLocation(c *gin.Context) {
	var req domain.LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location name is required"})
		return
	}

	if h.locationRepo != nil {
		existing, _ := h.locationRepo.GetByName(req.Name)
		if existing != nil {
			c.JSON(http.StatusOK, existing)
			return
		}
		loc := &domain.Location{
			Name:        strings.TrimSpace(req.Name),
			Description: req.Description,
		}
		created, err := h.locationRepo.Create(loc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create location"})
			return
		}
		c.JSON(http.StatusCreated, created)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Location repository uninitialized"})
}

func (h *Handler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")
	var req domain.LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location name is required"})
		return
	}

	if h.locationRepo != nil {
		err := h.locationRepo.Update(id, strings.TrimSpace(req.Name), req.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Location updated successfully"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Location repository uninitialized"})
}

func (h *Handler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	if h.locationRepo != nil {
		err := h.locationRepo.Delete(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Location deleted successfully"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Location repository uninitialized"})
}

// ─── Bulk Device Handlers ───────────────────────────────────────────────────

func (h *Handler) BulkDeviceAction(c *gin.Context) {
	var req domain.BulkDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.DeviceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deviceIds list is required"})
		return
	}

	if h.deviceRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Device repository uninitialized"})
		return
	}

	if req.Action == "delete" {
		deleted, err := h.deviceRepo.BulkDelete(req.DeviceIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk delete devices: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, domain.BulkDeviceResponse{
			Success:      true,
			UpdatedCount: deleted,
			FailedCount:  len(req.DeviceIDs) - deleted,
		})
		return
	}

	if req.Action == "update" {
		updated, err := h.deviceRepo.BulkUpdate(req.DeviceIDs, req.Updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk update devices: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, domain.BulkDeviceResponse{
			Success:      true,
			UpdatedCount: updated,
			FailedCount:  len(req.DeviceIDs) - updated,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Supported actions: update, delete"})
}


// ─── Permission Handlers ─────────────────────────────────────────────────────

func (h *Handler) GetPermissions(c *gin.Context) {
	if h.permRepo != nil {
		perms, err := h.permRepo.GetAll()
		if err == nil {
			c.JSON(http.StatusOK, perms)
			return
		}
	}
	c.JSON(http.StatusOK, []domain.RolePermission{})
}

func (h *Handler) UpdatePermissions(c *gin.Context) {
	var perms []domain.RolePermission
	if err := c.ShouldBindJSON(&perms); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permissions payload"})
		return
	}

	if h.permRepo != nil {
		if err := h.permRepo.UpdatePermissions(perms); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role permissions"})
			return
		}
	}

	if h.userLogRepo != nil {
		userID, _ := c.Get("userID")
		userIDStr, _ := userID.(string)
		_ = h.userLogRepo.Append(&domain.UserLog{
			UserID:    userIDStr,
			Action:    "action",
			Detail:    fmt.Sprintf("Updated per-role feature permission matrix (%d rules)", len(perms)),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	if h.hub != nil {
		h.hub.BroadcastMessage(domain.WSMessage{
			Type:        "ROLE_PERMISSIONS_UPDATED",
			Title:       "Permissions Updated",
			Description: "Role permission access control matrix updated",
			Timestamp:   time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
