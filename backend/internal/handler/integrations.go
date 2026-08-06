package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/notifier"
	"govmonitor-it/backend/internal/repository"
	"govmonitor-it/backend/internal/ws"

	"github.com/gin-gonic/gin"
)

// ─── Sidecar Proxy Client ─────────────────────────────────────────────────────

type sidecarClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func newSidecarClient(url, token string) *sidecarClient {
	return &sidecarClient{
		baseURL: strings.TrimRight(url, "/"),
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *sidecarClient) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
		req.Header.Set("X-Internal-Token", s.token)
	}
	return s.http.Do(req)
}

func (s *sidecarClient) get(path string) (*http.Response, error) {
	return s.do(http.MethodGet, path, nil)
}

func (s *sidecarClient) post(path string, body interface{}) (*http.Response, error) {
	return s.do(http.MethodPost, path, body)
}

// ─── Handler Types ────────────────────────────────────────────────────────────

type IntegrationsHandler struct {
	hub               *ws.Hub
	pipeline          *notifier.Pipeline
	settingsRepo      *repository.SettingsRepository
	whatsappTargetRepo repository.WhatsAppTargetRepository // New: dedicated repo for targets
	sidecar           *sidecarClient

	mu            sync.Mutex
	savedBotToken string
	savedChatID   string
}

// TelegramConfig stored in DB
type TelegramDBConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

// WhatsAppDBConfig stored in DB — tracks last known state from sidecar
type WhatsAppDBConfig struct {
	Status       string `json:"status"`
	LinkedNumber string `json:"linked_number"`
}

func NewIntegrationsHandler(hub *ws.Hub, pipeline *notifier.Pipeline, settingsRepo *repository.SettingsRepository, whatsappTargetRepo repository.WhatsAppTargetRepository, sidecarURL, sidecarToken string) *IntegrationsHandler {
	h := &IntegrationsHandler{
		hub:               hub,
		pipeline:          pipeline,
		settingsRepo:      settingsRepo,
		whatsappTargetRepo: whatsappTargetRepo,
		sidecar:           newSidecarClient(sidecarURL, sidecarToken),
	}

	// Load telegram config from DB on startup
	var tgCfg TelegramDBConfig
	if settingsRepo != nil {
		if err := settingsRepo.GetJSON("telegram", &tgCfg); err == nil && tgCfg.BotToken != "" {
			h.savedBotToken = tgCfg.BotToken
			h.savedChatID = tgCfg.ChatID
			log.Printf("[Integrations] Loaded Telegram config from DB: chat_id=%s", tgCfg.ChatID)
		}
	}

	return h
}

// ─── Telegram Helpers ─────────────────────────────────────────────────────────

func cleanChatID(chatID string) string {
	cleaned := strings.TrimSpace(chatID)
	cleaned = strings.Trim(cleaned, "()[]\"'")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return cleaned
}

func SendTelegramMessage(botToken, chatId, text string) error {
	cleanID := cleanChatID(chatId)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{
		"chat_id":    cleanID,
		"text":       text,
		"parse_mode": "HTML",
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Network error: %v", err)
	}
	defer resp.Body.Close()

	var tgResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("Failed to parse response from Telegram API")
	}

	if !tgResp.OK {
		if strings.Contains(strings.ToLower(tgResp.Description), "chat not found") {
			return fmt.Errorf("Telegram API Error: Bad Request: chat not found.\n\n" +
				"Cara Perbaiki:\n" +
				"1. Pastikan Anda sudah mengirim pesan /start ke Bot di Telegram.\n" +
				"2. Jika target adalah Grup/Channel, pastikan Bot SUDAH ditambahkan sebagai Administrator di grup/channel tersebut.\n" +
				"3. Pastikan Chat ID benar (Gunakan ID dari @userinfobot untuk chat personal, atau ID grup/channel seperti -1001982736412). " +
				"PENTING: Jangan gunakan ID Bot (%s) sebagai Chat ID!", cleanID)
		}
		return fmt.Errorf("Telegram API Error: %s", tgResp.Description)
	}
	return nil
}

// validateChatID returns an error if chatID matches the bot's own numeric ID (prefix of botToken).
func validateChatID(botToken, chatID string) error {
	cleanID := cleanChatID(chatID)
	if botToken == "" || cleanID == "" {
		return nil
	}
	parts := strings.SplitN(botToken, ":", 2)
	if len(parts) < 2 {
		return nil
	}
	botNumericID := parts[0]
	strippedChatID := strings.TrimPrefix(cleanID, "-")
	if strippedChatID == botNumericID {
		return fmt.Errorf("Chat ID '%s' adalah ID Bot itu sendiri! Bot tidak dapat mengirim pesan ke dirinya sendiri. Gunakan Chat ID personal Anda (dapatkan dari @userinfobot) atau ID Grup/Channel (contoh: -1001982736412). Jika menggunakan grup/channel, pastikan bot sudah ditambahkan sebagai Administrator!", chatID)
	}
	return nil
}

// ─── Telegram Endpoints ───────────────────────────────────────────────────────

// POST /api/v1/integrations/telegram/config
func (h *IntegrationsHandler) TelegramConfig(c *gin.Context) {
	var req struct {
		BotToken string `json:"botToken"`
		ChatID   string `json:"chatId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration payload"})
		return
	}

	cleanToken := strings.TrimSpace(req.BotToken)
	cleanID := cleanChatID(req.ChatID)

	if cleanToken == "" || cleanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both Bot Token and Chat/Channel ID are required"})
		return
	}

	if err := validateChatID(cleanToken, cleanID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update in-memory
	h.mu.Lock()
	h.savedBotToken = cleanToken
	h.savedChatID = cleanID
	h.mu.Unlock()

	if h.pipeline != nil {
		h.pipeline.UpdateTelegramConfig(cleanToken, cleanID)
	}

	// Persist to PostgreSQL
	if h.settingsRepo != nil {
		dbCfg := TelegramDBConfig{BotToken: cleanToken, ChatID: cleanID}
		if err := h.settingsRepo.SetJSON("telegram", dbCfg); err != nil {
			log.Printf("[Integrations] Failed to persist Telegram config to DB: %v", err)
		} else {
			log.Printf("[Integrations] Telegram config saved to PostgreSQL: chat_id=%s", cleanID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Telegram Bot token and Chat ID saved securely",
		"chatId":  cleanID,
	})
}

// GET /api/v1/integrations/telegram/config
func (h *IntegrationsHandler) TelegramGetConfig(c *gin.Context) {
	// Always prefer DB value over in-memory
	var tgCfg TelegramDBConfig
	if h.settingsRepo != nil {
		if err := h.settingsRepo.GetJSON("telegram", &tgCfg); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"botToken": tgCfg.BotToken,
				"chatId":   tgCfg.ChatID,
			})
			return
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"botToken": h.savedBotToken,
		"chatId":   h.savedChatID,
	})
}

// POST /api/v1/integrations/telegram/test
func (h *IntegrationsHandler) TelegramTest(c *gin.Context) {
	var req struct {
		BotToken string `json:"botToken"`
		ChatID   string `json:"chatId"`
	}
	_ = c.ShouldBindJSON(&req)

	token := strings.TrimSpace(req.BotToken)
	chatID := cleanChatID(req.ChatID)

	// Prefer DB values if request fields are empty
	if token == "" || chatID == "" {
		var tgCfg TelegramDBConfig
		if h.settingsRepo != nil {
			if err := h.settingsRepo.GetJSON("telegram", &tgCfg); err == nil {
				if token == "" {
					token = tgCfg.BotToken
				}
				if chatID == "" {
					chatID = tgCfg.ChatID
				}
			}
		}
		h.mu.Lock()
		if token == "" {
			token = h.savedBotToken
		}
		if chatID == "" {
			chatID = h.savedChatID
		}
		h.mu.Unlock()
	}

	token = strings.TrimSpace(token)
	chatID = cleanChatID(chatID)

	if token == "" || chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Both Bot Token and Chat/Channel ID are required",
		})
		return
	}

	if err := validateChatID(token, chatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	msg := fmt.Sprintf("🟢 <b>GovMonitor IT — Test Connection</b>\n\nTelegram Bot integration is active and operating normally.\nTarget Channel ID: <code>%s</code>\nTime: %s", chatID, time.Now().Format("15:04:05 WIB"))

	err := SendTelegramMessage(token, chatID, msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Persist verified config to PostgreSQL DB
	if h.settingsRepo != nil {
		dbCfg := TelegramDBConfig{BotToken: token, ChatID: chatID}
		_ = h.settingsRepo.SetJSON("telegram", dbCfg)
		log.Printf("[Integrations] Telegram test config saved to PostgreSQL: chat_id=%s", chatID)
	}

	h.mu.Lock()
	h.savedBotToken = token
	h.savedChatID = chatID
	h.mu.Unlock()

	if h.pipeline != nil {
		h.pipeline.UpdateTelegramConfig(token, chatID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Test alert delivered to Telegram Channel %s and saved to database", chatID),
	})
}

// ─── WhatsApp Endpoints (real sidecar proxy) ──────────────────────────────────
// All WhatsApp state lives in the wa-sidecar. The Go backend:
// 1. Proxies requests to the sidecar
// 2. Caches the resulting status in PostgreSQL (for startup persistence)
// 3. Broadcasts status changes over WebSocket

// POST /api/v1/integrations/whatsapp/connect
func (h *IntegrationsHandler) WhatsAppConnect(c *gin.Context) {
	resp, err := h.sidecar.post("/connect", nil)
	if err != nil {
		log.Printf("[WhatsApp] Sidecar unreachable at %s: %v", h.sidecar.baseURL, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("WhatsApp sidecar is not running (tried %s). Start it with: cd backend/wa-sidecar && node index.js", h.sidecar.baseURL),
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	// If sidecar already connected, persist state to DB
	if s, _ := result["status"].(string); s == "connected" {
		num, _ := result["linkedNumber"].(string)
		h.persistWAState("connected", num)
		if h.hub != nil {
			h.hub.Broadcast(gin.H{"type": "whatsapp_status", "status": "connected", "linkedNumber": num})
		}
	}

	c.JSON(resp.StatusCode, result)
}

// GET /api/v1/integrations/whatsapp/qr
// Returns the exact QR string from the sidecar — never re-encoded.
func (h *IntegrationsHandler) WhatsAppQR(c *gin.Context) {
	resp, err := h.sidecar.get("/qr")
	if err != nil {
		log.Printf("[WhatsApp] Sidecar unreachable at %s/qr: %v", h.sidecar.baseURL, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "disconnected",
			"qr":     "",
			"error":  "WhatsApp sidecar is not running",
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	// Log QR prefix for verification that it's the exact Baileys string
	if qr, ok := result["qr"].(string); ok && qr != "" {
		prefix := qr
		if len(prefix) > 30 {
			prefix = prefix[:30]
		}
		log.Printf("[WhatsApp] QR proxy: status=%v qr_len=%d qr_prefix=%s...", result["status"], len(qr), prefix)
	}

	// Update DB cache when status transitions
	if s, _ := result["status"].(string); s == "connected" {
		num, _ := result["linkedNumber"].(string)
		h.persistWAState("connected", num)
		if h.hub != nil {
			h.hub.Broadcast(gin.H{"type": "whatsapp_status", "status": "connected", "linkedNumber": num})
		}
	}

	c.JSON(resp.StatusCode, result)
}

// GET /api/v1/integrations/whatsapp/status
func (h *IntegrationsHandler) WhatsAppStatus(c *gin.Context) {
	resp, err := h.sidecar.get("/status")
	if err != nil {
		// Sidecar offline — return cached DB state so UI shows last-known status
		log.Printf("[WhatsApp] Sidecar offline, returning cached DB state")
		var waCfg WhatsAppDBConfig
		if h.settingsRepo != nil {
			_ = h.settingsRepo.GetJSON("whatsapp", &waCfg)
		}
		c.JSON(http.StatusOK, gin.H{
			"status":       waCfg.Status,
			"linkedNumber": waCfg.LinkedNumber,
			"lastSync":     "sidecar offline",
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	// Sync DB cache with sidecar truth
	if s, _ := result["status"].(string); s != "" {
		num, _ := result["linkedNumber"].(string)
		h.persistWAState(s, num)
		if s == "connected" && h.hub != nil {
			h.hub.Broadcast(gin.H{"type": "whatsapp_status", "status": "connected", "linkedNumber": num})
		}
	}

	// Enrich with lastSync timestamp
	result["lastSync"] = time.Now().Format("15:04:05 WIB")
	c.JSON(resp.StatusCode, result)
}

// POST /api/v1/integrations/whatsapp/disconnect
func (h *IntegrationsHandler) WhatsAppDisconnect(c *gin.Context) {
	resp, err := h.sidecar.post("/disconnect", nil)
	if err != nil {
		log.Printf("[WhatsApp] Sidecar unreachable for disconnect: %v", err)
		// Even if sidecar is unreachable, clear our cached state
		h.persistWAState("disconnected", "")
		if h.hub != nil {
			h.hub.Broadcast(gin.H{"type": "whatsapp_status", "status": "disconnected"})
		}
		c.JSON(http.StatusOK, gin.H{"status": "disconnected", "message": "Local state cleared (sidecar was unreachable)"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	h.persistWAState("disconnected", "")
	if h.hub != nil {
		h.hub.Broadcast(gin.H{"type": "whatsapp_status", "status": "disconnected"})
	}

	c.JSON(resp.StatusCode, result)
}

// ─── WhatsApp Target Endpoints ────────────────────────────────────────────────
// Target numbers are stored in PostgreSQL and used by the pipeline to dispatch notifications.

// GET /api/v1/integrations/whatsapp/targets
func (h *IntegrationsHandler) ListWhatsAppTargets(c *gin.Context) {
	targets, err := h.whatsappTargetRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve WhatsApp targets"})
		return
	}
	c.JSON(http.StatusOK, targets)
}

// POST /api/v1/integrations/whatsapp/targets
func (h *IntegrationsHandler) CreateWhatsAppTarget(c *gin.Context) {
	var req struct {
		Label       string `json:"label"`
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target payload"})
		return
	}

	if req.Label == "" || req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Label and Phone Number are required"})
		return
	}

	// Normalize phone number to JID
	// Remove all non-digit characters
	phoneNumberDigits := strings.ReplaceAll(req.PhoneNumber, " ", "")
	phoneNumberDigits = strings.ReplaceAll(phoneNumberDigits, "+", "")
	phoneNumberDigits = strings.ReplaceAll(phoneNumberDigits, "-", "")
	// Assuming country code is already part of the number, or handling if '0' prefix
	// Example: +62812... -> 62812...
	if strings.HasPrefix(phoneNumberDigits, "0") && len(phoneNumberDigits) > 1 {
		phoneNumberDigits = "62" + phoneNumberDigits[1:] // Indonesian specific
	}
	if !strings.HasPrefix(phoneNumberDigits, "62") { // Basic validation: ensure it has country code prefix
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number must include country code (e.g., +62...) or start with '62'"})
		return
	}

	jid := phoneNumberDigits + "@s.whatsapp.net"

	newTarget := &domain.WhatsAppTarget{
		Label:       req.Label,
		PhoneNumber: req.PhoneNumber,
		JID:         jid,
	}

	createdTarget, err := h.whatsappTargetRepo.Create(newTarget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create WhatsApp target"})
		return
	}
	c.JSON(http.StatusCreated, createdTarget)
}

// PUT /api/v1/integrations/whatsapp/targets/:id
func (h *IntegrationsHandler) UpdateWhatsAppTarget(c *gin.Context) {
	targetID := c.Param("id")
	var req struct {
		Label       string `json:"label"`
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target payload"})
		return
	}

	if req.Label == "" || req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Label and Phone Number are required"})
		return
	}

	// Normalize phone number to JID
	phoneNumberDigits := strings.ReplaceAll(req.PhoneNumber, " ", "")
	phoneNumberDigits = strings.ReplaceAll(phoneNumberDigits, "+", "")
	phoneNumberDigits = strings.ReplaceAll(phoneNumberDigits, "-", "")
	if strings.HasPrefix(phoneNumberDigits, "0") && len(phoneNumberDigits) > 1 {
		phoneNumberDigits = "62" + phoneNumberDigits[1:]
	}
	if !strings.HasPrefix(phoneNumberDigits, "62") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number must include country code (e.g., +62...) or start with '62'"})
		return
	}

	jid := phoneNumberDigits + "@s.whatsapp.net"

	target, err := h.whatsappTargetRepo.GetByID(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve target"})
		return
	}
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "WhatsApp target not found"})
		return
	}

	target.Label = req.Label
	target.PhoneNumber = req.PhoneNumber
	target.JID = jid

	if err := h.whatsappTargetRepo.Update(target); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update WhatsApp target"})
		return
	}
	c.JSON(http.StatusOK, target)
}

// DELETE /api/v1/integrations/whatsapp/targets/:id
func (h *IntegrationsHandler) DeleteWhatsAppTarget(c *gin.Context) {
	targetID := c.Param("id")
	if err := h.whatsappTargetRepo.Delete(targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete WhatsApp target"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// POST /api/v1/integrations/whatsapp/test
func (h *IntegrationsHandler) WhatsAppTest(c *gin.Context) {
	var req struct {
		TargetID string `json:"targetId"` // New: send to specific target
		Message  string `json:"message"`
	}
	_ = c.ShouldBindJSON(&req)

	// First check sidecar is alive and connected
	statusResp, err := h.sidecar.get("/status")
	if err != nil {
		log.Printf("[WhatsApp] Sidecar unreachable at %s: %v", h.sidecar.baseURL, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   fmt.Sprintf("WhatsApp sidecar is not running (tried %s). Start it with: cd backend/wa-sidecar && node index.js", h.sidecar.baseURL),
		})
		return
	}
	defer statusResp.Body.Close()

	var status map[string]interface{}
	_ = json.NewDecoder(statusResp.Body).Decode(&status)

	if s, _ := status["status"].(string); s != "connected" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("WhatsApp is not connected (status: %s). Please scan QR code first.", s),
		})
		return
	}

	var target *domain.WhatsAppTarget
	if req.TargetID != "" {
		target, err = h.whatsappTargetRepo.GetByID(req.TargetID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve target for test"})
			return
		}
		if target == nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Test target not found"})
			return
		}
	} else {
		// Fallback to first target if no specific ID provided, or error if none
		targets, err := h.whatsappTargetRepo.GetAll()
		if err != nil || len(targets) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No WhatsApp targets configured to send test message"})
			return
		}
		target = &targets[0] // Use the first target found
	}


	msg := req.Message
	if msg == "" {
		linkedNum, _ := status["linkedNumber"].(string)
		msg = fmt.Sprintf("🟢 *GovMonitor IT — WhatsApp Test*\n\nBaileys gateway is connected and operational.\nLinked Number: %s\nTarget: %s\nTime: %s", linkedNum, target.PhoneNumber, time.Now().Format("15:04:05 WIB"))
	} else {
		msg = notifier.FormatForWhatsApp(msg)
	}

	sendResp, err := h.sidecar.post("/send", map[string]string{
		"recipient": target.JID,
		"message":   msg,
	})
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Failed to communicate with WhatsApp sidecar: " + err.Error(),
		})
		return
	}
	defer sendResp.Body.Close()

	var sendResult map[string]interface{}
	_ = json.NewDecoder(sendResp.Body).Decode(&sendResult)

	if sendResp.StatusCode != http.StatusOK {
		errMsg, _ := sendResult["error"].(string)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Test WhatsApp message sent to %s", target.PhoneNumber),
		"recipient": target.PhoneNumber,
	})
}

// ─── Internal Helpers ─────────────────────────────────────────────────────────

func (h *IntegrationsHandler) persistWAState(status, linkedNumber string) {
	if h.settingsRepo == nil {
		return
	}
	waCfg := WhatsAppDBConfig{Status: status, LinkedNumber: linkedNumber}
	if err := h.settingsRepo.SetJSON("whatsapp", waCfg); err != nil {
		log.Printf("[WhatsApp] Failed to persist state to DB: %v", err)
	} else {
		log.Printf("[WhatsApp] State persisted: status=%s linkedNumber=%s", status, linkedNumber)
	}
}
