// Package notifier implements the WhatsApp + Telegram notification pipeline
// with token-bucket rate limiting, aggregation window, and fallback logic.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/repository"
)

// ─── Rate Limiter (Token Bucket) ─────────────────────────────────────────────

// TokenBucket implements a simple token-bucket rate limiter without external deps.
// For production, replace with golang.org/x/time/rate.Limiter.
type TokenBucket struct {
	mu        sync.Mutex
	tokens    float64
	maxTokens float64
	rate      float64 // tokens per second
	lastTime  time.Time
}

func NewTokenBucket(maxPerMinute int) *TokenBucket {
	rate := float64(maxPerMinute) / 60.0
	return &TokenBucket{
		tokens:    float64(maxPerMinute),
		maxTokens: float64(maxPerMinute),
		rate:      rate,
		lastTime:  time.Now(),
	}
}

// Wait blocks until a token is available.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		tb.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(tb.lastTime).Seconds()
		tb.tokens = min(tb.maxTokens, tb.tokens+elapsed*tb.rate)
		tb.lastTime = now
		if tb.tokens >= 1 {
			tb.tokens--
			tb.mu.Unlock()
			return nil
		}
		tb.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ─── Message Template ─────────────────────────────────────────────────────────

const incidentTmplStr = `🔴 INCIDENT — {{ len .Devices }} Device{{ if gt (len .Devices) 1 }}s{{ end }} Down
{{- if .RootDevice }}
Root cause: {{ .RootDevice.Name }} ({{ .RootDevice.Type }}) — {{ .RootDevice.IP }}
{{ end }}
{{- range .LocationGroups }}
📍 {{ .Location }} ({{ len .Devices }} device{{ if gt (len .Devices) 1 }}s{{ end }})
{{- range .Devices }}
  • {{ .Name }} — {{ .Type }} — {{ .IP }}
{{- end }}
{{ end }}
Detected: {{ .DetectedAt.Format "15:04:05" }} · Debounce passed: {{ .DebouncedAt.Format "15:04:05" }}`

var incidentTmpl = template.Must(template.New("incident").Parse(incidentTmplStr))

type LocationGroup struct {
	Location string
	Devices  []DeviceInfo
}

type DeviceInfo struct {
	Name string
	Type string
	IP   string
}

type IncidentTemplateData struct {
	Devices        []DeviceInfo
	RootDevice     *DeviceInfo
	LocationGroups []LocationGroup
	DetectedAt     time.Time
	DebouncedAt    time.Time
}

// RenderIncidentMessage formats the aggregated incident as a structured text message.
func RenderIncidentMessage(devices []DeviceInfo, root *DeviceInfo, detectedAt, debouncedAt time.Time) (string, error) {
	// Group by location
	groupMap := make(map[string][]DeviceInfo)
	for _, d := range devices {
		groupMap[d.Type] = append(groupMap[d.Type], d) // simplified — real grouping by location field
	}

	var groups []LocationGroup
	for loc, devs := range groupMap {
		groups = append(groups, LocationGroup{Location: loc, Devices: devs})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Location < groups[j].Location })

	data := IncidentTemplateData{
		Devices:        devices,
		RootDevice:     root,
		LocationGroups: groups,
		DetectedAt:     detectedAt,
		DebouncedAt:    debouncedAt,
	}

	var buf bytes.Buffer
	if err := incidentTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render incident template: %w", err)
	}
	return buf.String(), nil
}

// ─── WhatsApp Client (Baileys HTTP Sidecar) ───────────────────────────────────

type WhatsAppClient struct {
	sidecarURL string
	token      string
	httpClient *http.Client
}

func NewWhatsAppClient(sidecarURL, token string) *WhatsAppClient {
	return &WhatsAppClient{
		sidecarURL: sidecarURL,
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends a message via the Baileys Node.js sidecar.
// Returns error if the sidecar is unreachable or returns a non-2xx status.
func (w *WhatsAppClient) Send(recipient, message string) error {
	payload := map[string]string{
		"recipient": recipient,
		"message":   message,
	}
	body, _ := json.Marshal(payload)

	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		req, err := http.NewRequest("POST", w.sidecarURL+"/send", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if w.token != "" {
			req.Header.Set("Authorization", "Bearer "+w.token)
			req.Header.Set("X-Internal-Token", w.token)
		}

		resp, err := w.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("whatsapp sidecar unreachable: %w", err)
			if attempt < 2 {
				time.Sleep(1 * time.Second)
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("whatsapp sidecar returned %d", resp.StatusCode)
			if resp.StatusCode == 503 && attempt < 2 {
				time.Sleep(1 * time.Second)
				continue
			}
			return lastErr
		}
		return nil
	}
	return lastErr
}

// IsHealthy checks the sidecar's /health endpoint.
func (w *WhatsAppClient) IsHealthy() bool {
	req, err := http.NewRequest("GET", w.sidecarURL+"/health", nil)
	if err != nil {
		return false
	}
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
		req.Header.Set("X-Internal-Token", w.token)
	}
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// ─── Telegram Client ──────────────────────────────────────────────────────────

type TelegramClient struct {
	botToken string
	chatID   string
	httpClient *http.Client
}

func NewTelegramClient(botToken, chatID string) *TelegramClient {
	return &TelegramClient{
		botToken:   botToken,
		chatID:     chatID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TelegramClient) Send(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       message,
		"parse_mode": "HTML",
	}
	body, _ := json.Marshal(payload)
	resp, err := t.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram API error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

// ─── Aggregation Buffer ───────────────────────────────────────────────────────

// AggregationBuffer collects device-down events within a window and emits
// a single batched notification after the window expires.
type AggregationBuffer struct {
	mu          sync.Mutex
	window      time.Duration
	buckets     map[string]*bucket // keyed by parentDeviceID or "orphan"
	onEmit      func(devices []DeviceInfo, rootKey string, detectedAt time.Time)
}

type bucket struct {
	devices    []DeviceInfo
	detectedAt time.Time
	timer      *time.Timer
}

func NewAggregationBuffer(window time.Duration, onEmit func(devices []DeviceInfo, rootKey string, detectedAt time.Time) ) *AggregationBuffer {
	return &AggregationBuffer{
		window:  window,
		buckets: make(map[string]*bucket),
		onEmit:  onEmit,
	}
}

// Add adds a device-down event. The first event starts the aggregation window timer.
func (ab *AggregationBuffer) Add(parentKey string, dev DeviceInfo) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	b, exists := ab.buckets[parentKey]
	if !exists {
		detectedAt := time.Now()
		b = &bucket{
			detectedAt: detectedAt,
			devices:    []DeviceInfo{dev},
		}
		b.timer = time.AfterFunc(ab.window, func() {
			ab.mu.Lock()
			defer ab.mu.Unlock()
			if bkt, ok := ab.buckets[parentKey]; ok {
				devs := bkt.devices
				at := bkt.detectedAt
				delete(ab.buckets, parentKey)
				go ab.onEmit(devs, parentKey, at)
			}
		})
		ab.buckets[parentKey] = b
	} else {
		b.devices = append(b.devices, dev)
	}
}

// ─── Notification Pipeline ────────────────────────────────────────────────────

type NotificationLog struct {
	Channel   string
	Recipient string
	Status    string // "delivered" | "failed"
	Error     string
	Timestamp time.Time
}

type Pipeline struct {
	whatsApp            *WhatsAppClient
	telegram            *TelegramClient
	limiter             *TokenBucket
	logChan             chan NotificationLog
	whatsAppTargetRepo  repository.WhatsAppTargetRepository
	notifLogRepo        repository.NotificationLogRepository
	defaultTargetNumber string
}

func NewPipeline(whatsAppURL, whatsAppToken, telegramToken, telegramChatID, defaultTargetNumber string, maxMsgPerMin int, whatsAppTargetRepo repository.WhatsAppTargetRepository) *Pipeline {
	p := &Pipeline{
		whatsApp:            NewWhatsAppClient(whatsAppURL, whatsAppToken),
		telegram:            NewTelegramClient(telegramToken, telegramChatID),
		limiter:             NewTokenBucket(maxMsgPerMin),
		logChan:             make(chan NotificationLog, 256),
		whatsAppTargetRepo:  whatsAppTargetRepo,
		defaultTargetNumber: defaultTargetNumber,
	}
	go p.drainLogs()
	return p
}

func (p *Pipeline) SetNotifLogRepo(repo repository.NotificationLogRepository) {
	if p != nil {
		p.notifLogRepo = repo
	}
}

func (p *Pipeline) SetDefaultTargetNumber(num string) {
	if p == nil {
		return
	}
	p.defaultTargetNumber = num
}

func (p *Pipeline) UpdateTelegramConfig(botToken, chatID string) {
	if p == nil {
		return
	}
	p.telegram = NewTelegramClient(botToken, chatID)
}

// Send dispatches a message: tries WhatsApp first to all configured targets, falls back to Telegram.
// Rate-limited by the token bucket.
func (p *Pipeline) Send(ctx context.Context, message string) { // Recipient removed from signature
	if err := p.limiter.Wait(ctx); err != nil {
		log.Printf("[Notifier] Rate limit wait cancelled: %v", err)
		return
	}

	// Fetch all configured WhatsApp targets if repository is available
	var targets []domain.WhatsAppTarget
	if p.whatsAppTargetRepo != nil {
		var err error
		targets, err = p.whatsAppTargetRepo.GetAll()
		if err != nil {
			log.Printf("[Notifier] Failed to fetch WhatsApp targets: %v", err)
			targets = []domain.WhatsAppTarget{} // Continue with no WhatsApp targets
		}
	}

	// Fall back to defaultTargetNumber (.env or system setting) if DB targets table is empty
	if len(targets) == 0 && p.defaultTargetNumber != "" {
		phone := strings.TrimSpace(p.defaultTargetNumber)
		numOnly := strings.ReplaceAll(strings.ReplaceAll(phone, "+", ""), " ", "")
		jid := numOnly
		if !strings.HasSuffix(jid, "@s.whatsapp.net") {
			jid = jid + "@s.whatsapp.net"
		}
		targets = append(targets, domain.WhatsAppTarget{
			ID:          "default-env-target",
			Label:       "Default Target (.env)",
			PhoneNumber: phone,
			JID:         jid,
		})
	}

	whatsAppSent := false
	if p.whatsApp.IsHealthy() {
		if len(targets) == 0 {
			log.Println("[Notifier] No WhatsApp targets configured, skipping WhatsApp send.")
		}
		waMessage := FormatForWhatsApp(message)
		for _, target := range targets {
			err := p.whatsApp.Send(target.JID, waMessage) // Send formatted WA message (*bold*)
			if err == nil {
				whatsAppSent = true
				p.logChan <- NotificationLog{
					Channel:   "WhatsApp",
					Recipient: target.PhoneNumber, // Log with actual phone number
					Status:    "delivered",
					Timestamp: time.Now(),
				}
				log.Printf("[Notifier] WhatsApp message delivered to %s", target.PhoneNumber)
			} else {
				log.Printf("[Notifier] WhatsApp failed for %s: %v", target.PhoneNumber, err)
				p.logChan <- NotificationLog{
					Channel:   "WhatsApp",
					Recipient: target.PhoneNumber,
					Status:    "failed",
					Error:     err.Error(),
					Timestamp: time.Now(),
				}
			}
		}
	}

	if !whatsAppSent { // Fallback to Telegram only if no WhatsApp message was successfully sent
		log.Println("[Notifier] Falling back to Telegram.")
		// Telegram fallback: format with HTML <b>bold</b> tags
		tgMessage := FormatForTelegram(message)
		err := p.telegram.Send(tgMessage)
		status := "delivered"
		errStr := ""
		if err != nil {
			log.Printf("[Notifier] Telegram also failed: %v", err)
			status = "failed"
			errStr = err.Error()
		}
		p.logChan <- NotificationLog{
			Channel:   "Telegram",
			Recipient: p.telegram.chatID, // Use the configured Telegram chat ID for logging
			Status:    status,
			Error:     errStr,
			Timestamp: time.Now(),
		}
	}
}

// FormatForWhatsApp converts HTML tags (<b>, <code>) or double asterisks (**bold**) into WhatsApp native markdown (*bold*, `code`).
func FormatForWhatsApp(msg string) string {
	res := msg

	// Replace HTML tags with WhatsApp syntax
	res = strings.ReplaceAll(res, "<b>", "*")
	res = strings.ReplaceAll(res, "</b>", "*")
	res = strings.ReplaceAll(res, "<strong>", "*")
	res = strings.ReplaceAll(res, "</strong>", "*")
	res = strings.ReplaceAll(res, "<code>", "`")
	res = strings.ReplaceAll(res, "</code>", "`")
	res = strings.ReplaceAll(res, "<i>", "_")
	res = strings.ReplaceAll(res, "</i>", "_")

	// Convert double asterisks **bold** to single asterisk *bold* for WhatsApp
	for strings.Contains(res, "**") {
		res = strings.Replace(res, "**", "*", 1)
	}

	return res
}

// FormatForTelegram converts WhatsApp markdown (*bold*) or double asterisks into Telegram HTML tags (<b>bold</b>, <code>code</code>).
func FormatForTelegram(msg string) string {
	res := msg

	// If message already contains HTML tags like <b> and <code>, preserve them cleanly
	if strings.Contains(res, "<b>") || strings.Contains(res, "<code>") {
		res = strings.ReplaceAll(res, "&", "&amp;")
		res = strings.ReplaceAll(res, "&amp;amp;", "&amp;")
		return res
	}

	// Convert **bold** -> <b>bold</b>
	for strings.Contains(res, "**") {
		res = strings.Replace(res, "**", "<b>", 1)
		if strings.Contains(res, "**") {
			res = strings.Replace(res, "**", "</b>", 1)
		} else {
			res += "</b>"
		}
	}

	// Convert single *bold* (WhatsApp style) -> <b>bold</b>
	if strings.Contains(res, "*") {
		for strings.Contains(res, "*") {
			res = strings.Replace(res, "*", "<b>", 1)
			if strings.Contains(res, "*") {
				res = strings.Replace(res, "*", "</b>", 1)
			} else {
				res += "</b>"
			}
		}
	}

	// Convert `code` -> <code>code</code>
	if strings.Contains(res, "`") {
		for strings.Contains(res, "`") {
			res = strings.Replace(res, "`", "<code>", 1)
			if strings.Contains(res, "`") {
				res = strings.Replace(res, "`", "</code>", 1)
			} else {
				res += "</code>"
			}
		}
	}

	res = strings.ReplaceAll(res, "&", "&amp;")
	res = strings.ReplaceAll(res, "&amp;amp;", "&amp;")
	return res
}

func (p *Pipeline) drainLogs() {
	for entry := range p.logChan {
		log.Printf("[Notifier] %s → %s | status=%s | err=%s",
			entry.Channel, entry.Recipient, entry.Status, entry.Error)
		if p.notifLogRepo != nil {
			_ = p.notifLogRepo.Append(&domain.NotificationLogRow{
				Channel:     entry.Channel,
				Recipient:   entry.Recipient,
				Status:      entry.Status,
				ChannelIcon: entry.Error,
			})
		}
	}
}
