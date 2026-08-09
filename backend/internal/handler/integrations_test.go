package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"govmonitor-it/backend/internal/handler"
	"govmonitor-it/backend/internal/notifier"
	"govmonitor-it/backend/internal/repository"
	"govmonitor-it/backend/internal/ws"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *handler.IntegrationsHandler, *repository.SettingsRepository) {
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub()
	go hub.Run()

	whatsappTargetRepo := repository.NewPostgresWhatsAppTargetRepository(nil)
	pipeline := notifier.NewPipeline("http://localhost:3001", "mock-token", "", "", "+6281234567890", 60, whatsappTargetRepo, nil, nil)
	// Passing nil db creates in-memory fallback for settingsRepo
	settingsRepo := repository.NewSettingsRepository(nil)
	integH := handler.NewIntegrationsHandler(hub, pipeline, settingsRepo, whatsappTargetRepo, "http://sidecar-mock:3001", "mock-token")

	r := gin.New()
	v1 := r.Group("/api/v1/integrations")
	{
		v1.POST("/whatsapp/connect", integH.WhatsAppConnect)
		v1.GET("/whatsapp/qr", integH.WhatsAppQR)
		v1.GET("/whatsapp/status", integH.WhatsAppStatus)
		v1.POST("/whatsapp/disconnect", integH.WhatsAppDisconnect)
		v1.POST("/whatsapp/test", integH.WhatsAppTest)
		v1.GET("/telegram/config", integH.TelegramGetConfig)
		v1.POST("/telegram/config", integH.TelegramConfig)
		v1.POST("/telegram/test", integH.TelegramTest)
	}

	return r, integH, settingsRepo
}

// ─── Bug 1 Test: Telegram chat_id == bot_id validation ───────────────────────
func TestTelegramChatIDValidation(t *testing.T) {
	r, _, _ := setupTestRouter()

	// 1. Attempt to save chat_id that equals bot's numeric ID (8983148351)
	botToken := "8983148351:AAFfxQYzYTw08-na7RxNVCzpaaw4Ww6FZUw"
	invalidChatID := "8983148351"

	body, _ := json.Marshal(map[string]string{
		"botToken": botToken,
		"chatId":   invalidChatID,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integrations/telegram/config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request when chat_id == bot_id, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errMsg, _ := resp["error"].(string)
	if errMsg == "" || (!bytes.Contains([]byte(errMsg), []byte("matches your bot's own ID")) && !bytes.Contains([]byte(errMsg), []byte("ID Bot itu sendiri"))) {
		t.Errorf("Expected clear error message explaining chat_id matches bot ID, got: %s", errMsg)
	}

	// 2. Save with valid group/channel chat_id (e.g. -1001982736412)
	validChatID := "-1001982736412"
	bodyValid, _ := json.Marshal(map[string]string{
		"botToken": botToken,
		"chatId":   validChatID,
	})

	reqValid, _ := http.NewRequest(http.MethodPost, "/api/v1/integrations/telegram/config", bytes.NewBuffer(bodyValid))
	reqValid.Header.Set("Content-Type", "application/json")
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for valid chat_id, got %d. Body: %s", wValid.Code, wValid.Body.String())
	}
}

// ─── Bug 2 Test: Telegram config persistence and re-test ────────────────────
func TestTelegramConfigPersistenceAndRetest(t *testing.T) {
	r, _, _ := setupTestRouter()

	// 1. Save initial Telegram config
	token1 := "7129847123:AAH3k891k1zL0P921"
	chat1 := "-1001982736412"

	body1, _ := json.Marshal(map[string]string{
		"botToken": token1,
		"chatId":   chat1,
	})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/integrations/telegram/config", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("Failed to save initial Telegram config: %d", w1.Code)
	}

	// 2. Fetch saved config via GET /telegram/config to verify persistence
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/integrations/telegram/config", nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)

	var getResp map[string]string
	_ = json.Unmarshal(wGet.Body.Bytes(), &getResp)
	if getResp["chatId"] != chat1 {
		t.Errorf("Expected fetched chatId %s, got %s", chat1, getResp["chatId"])
	}

	// 3. Edit config to new chat ID
	chat2 := "-1009876543210"
	body2, _ := json.Marshal(map[string]string{
		"botToken": token1,
		"chatId":   chat2,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/integrations/telegram/config", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Failed to update Telegram config: %d", w2.Code)
	}

	// 4. Verify GET returns updated chat ID
	reqGet2, _ := http.NewRequest(http.MethodGet, "/api/v1/integrations/telegram/config", nil)
	wGet2 := httptest.NewRecorder()
	r.ServeHTTP(wGet2, reqGet2)

	_ = json.Unmarshal(wGet2.Body.Bytes(), &getResp)
	if getResp["chatId"] != chat2 {
		t.Errorf("Expected updated chatId %s, got %s", chat2, getResp["chatId"])
	}
}

// ─── WhatsApp Integration Test ───────────────────────────────────────────────
func TestWhatsAppQRLifecycle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/connect" || r.URL.Path == "/status" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":       "connected",
				"linkedNumber": "6281290008888",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	gin.SetMode(gin.TestMode)
	hub := ws.NewHub()
	go hub.Run()

	whatsappTargetRepo := repository.NewPostgresWhatsAppTargetRepository(nil)
	pipeline := notifier.NewPipeline(ts.URL, "mock-token", "", "", "+6281234567890", 60, whatsappTargetRepo, nil, nil)
	settingsRepo := repository.NewSettingsRepository(nil)
	integH := handler.NewIntegrationsHandler(hub, pipeline, settingsRepo, whatsappTargetRepo, ts.URL, "mock-token")

	r := gin.New()
	v1 := r.Group("/api/v1/integrations")
	{
		v1.POST("/whatsapp/connect", integH.WhatsAppConnect)
		v1.GET("/whatsapp/status", integH.WhatsAppStatus)
	}

	reqConn, _ := http.NewRequest(http.MethodPost, "/api/v1/integrations/whatsapp/connect", nil)
	wConn := httptest.NewRecorder()
	r.ServeHTTP(wConn, reqConn)

	if wConn.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on connect, got %d", wConn.Code)
	}

	reqStatus, _ := http.NewRequest(http.MethodGet, "/api/v1/integrations/whatsapp/status", nil)
	wStatus := httptest.NewRecorder()
	r.ServeHTTP(wStatus, reqStatus)

	var statusResp map[string]interface{}
	_ = json.Unmarshal(wStatus.Body.Bytes(), &statusResp)

	if statusResp["status"] != "connected" {
		t.Errorf("Expected WhatsApp status 'connected', got %v", statusResp["status"])
	}
	if statusResp["linkedNumber"] != "6281290008888" {
		t.Errorf("Expected linked number '6281290008888', got %v", statusResp["linkedNumber"])
	}
}
