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

	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/repository"

	"github.com/hibiken/asynq"
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
	IncidentIDs []string
	Channel     string
	Recipient   string
	Status      string // "delivered" | "failed"
	Error       string
	Timestamp   time.Time
}

type Pipeline struct {
	whatsApp            *WhatsAppClient
	telegram            *TelegramClient
	limiter             *TokenBucket
	logChan             chan NotificationLog
	whatsAppTargetRepo  repository.WhatsAppTargetRepository
	notifLogRepo        repository.NotificationLogRepository
	incidentRepo        repository.IncidentRepository
	settingsRepo        *repository.SettingsRepository
	deviceRepo          repository.DeviceRepository
	notifyQueue         *NotifyQueue
	defaultTargetNumber string
	redisClient         *repository.RedisClient
	asynqClient         *asynq.Client
}

func NewPipeline(whatsAppURL, whatsAppToken, telegramToken, telegramChatID, defaultTargetNumber string, maxMsgPerMin int, whatsAppTargetRepo repository.WhatsAppTargetRepository, redisClient *repository.RedisClient, asynqClient *asynq.Client) *Pipeline {
	p := &Pipeline{
		whatsApp:            NewWhatsAppClient(whatsAppURL, whatsAppToken),
		telegram:            NewTelegramClient(telegramToken, telegramChatID),
		limiter:             NewTokenBucket(maxMsgPerMin),
		logChan:             make(chan NotificationLog, 256),
		whatsAppTargetRepo:  whatsAppTargetRepo,
		defaultTargetNumber: defaultTargetNumber,
		redisClient:         redisClient,
		asynqClient:         asynqClient,
	}
	go p.drainLogs()
	return p
}

func (p *Pipeline) SetDeviceRepo(repo repository.DeviceRepository) {
	if p != nil {
		p.deviceRepo = repo
	}
}

func (p *Pipeline) SetNotifyQueue(nq *NotifyQueue) {
	if p != nil {
		p.notifyQueue = nq
	}
}


func (p *Pipeline) SetSettingsRepo(repo *repository.SettingsRepository) {
	if p != nil {
		p.settingsRepo = repo
	}
}

func (p *Pipeline) SetNotifLogRepo(repo repository.NotificationLogRepository) {
	if p != nil {
		p.notifLogRepo = repo
	}
}

func (p *Pipeline) SetIncidentRepo(repo repository.IncidentRepository) {
	if p != nil {
		p.incidentRepo = repo
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

func (p *Pipeline) writeEvent(incidentID, eventType, channel, detail string) {
	if p.incidentRepo != nil && incidentID != "" {
		_ = p.incidentRepo.CreateEvent(&domain.IncidentEvent{
			IncidentID: incidentID,
			EventType:  eventType,
			Channel:    channel,
			Detail:     detail,
			OccurredAt: time.Now(),
		})
	}
}

// Send dispatches a message via the Asynq scheduled task queue to enforce the configured rate limit.
func (p *Pipeline) GetRateLimitMaxMsgPerMin() int {
	if p.settingsRepo != nil {
		var sysSettings struct {
			RateLimitMaxMsgPerMin int `json:"rateLimitMaxMsgPerMin"`
		}
		if err := p.settingsRepo.GetJSON("system_settings", &sysSettings); err == nil && sysSettings.RateLimitMaxMsgPerMin > 0 {
			return sysSettings.RateLimitMaxMsgPerMin
		}
	}
	return 60
}

func (p *Pipeline) Send(ctx context.Context, message string, incidentIDs []string) {
	p.SendExt(ctx, message, incidentIDs, false)
}

func (p *Pipeline) SendExt(ctx context.Context, message string, incidentIDs []string, direct bool) {
	if p.asynqClient == nil || p.redisClient == nil || p.redisClient.Client == nil {
		log.Println("[Notifier] Asynq client or Redis is not initialized. Sending alert synchronously...")
		_ = p.dispatchActualWithErr(ctx, message, incidentIDs)
		return
	}

	spacingSec := int64(60)
	if p.settingsRepo != nil {
		var sysSettings struct {
			RateLimitMaxMsgPerMin int `json:"rateLimitMaxMsgPerMin"`
		}
		if err := p.settingsRepo.GetJSON("system_settings", &sysSettings); err == nil && sysSettings.RateLimitMaxMsgPerMin > 0 {
			spacingSec = int64(sysSettings.RateLimitMaxMsgPerMin)
		}
	}

	// Calculate the earliest allowed processing time using the Lua script (dynamic spacing)
	scheduledTime, err := GetNextAllowedTime(ctx, p.redisClient.Client, spacingSec)
	if err != nil {
		log.Printf("[Notifier] Failed to calculate scheduled time via Redis: %v. Sending synchronously...", err)
		_ = p.dispatchActualWithErr(ctx, message, incidentIDs)
		return
	}

	for _, incID := range incidentIDs {
		var detail string
		if scheduledTime.After(time.Now().Add(1000 * time.Millisecond)) {
			detail = fmt.Sprintf("Rate limited — notification scheduled for %s", scheduledTime.Format("15:04:05 WIB"))
		} else {
			detail = "Rate limit OK — dispatching notification now"
		}
		p.writeEvent(incID, "rate_limit_phase", "", detail)
	}

	// Determine task type (DOWN or RECOVERED) based on message content
	taskType := "DOWN"
	msgLower := strings.ToLower(message)
	if strings.Contains(msgLower, "recovered") || strings.Contains(msgLower, "up") || strings.Contains(msgLower, "online") ||
		strings.Contains(message, "🟢") || strings.Contains(message, "✅") || strings.Contains(message, "✨") {
		taskType = "RECOVERED"
	}

	// Prepare payload
	pld := WhatsAppDispatchPayload{
		Message:     message,
		IncidentIDs: incidentIDs,
		CreatedAt:   time.Now(),
		TaskType:    taskType,
		Direct:      direct,
	}
	bytesPayload, err := json.Marshal(pld)
	if err != nil {
		log.Printf("[Notifier] Failed to marshal task payload: %v. Sending synchronously...", err)
		_ = p.dispatchActualWithErr(ctx, message, incidentIDs)
		return
	}

	task := asynq.NewTask(TypeWhatsAppDispatch, bytesPayload, asynq.MaxRetry(3), asynq.Queue("whatsapp"))

	// Enqueue the task with ProcessAt
	info, err := p.asynqClient.Enqueue(task, asynq.ProcessAt(scheduledTime))
	if err != nil {
		log.Printf("[Notifier] Failed to enqueue task: %v. Sending synchronously...", err)
		_ = p.dispatchActualWithErr(ctx, message, incidentIDs)
		return
	}

	log.Printf("[Notifier] Task %s enqueued successfully with spacing %ds. Job ID: %s. Scheduled for: %v. Direct=%t", TypeWhatsAppDispatch, spacingSec, info.ID, scheduledTime.Format("15:04:05.000"), direct)
}


// dispatchActualWithErr performs the actual delivery attempt and records log data, returning an error if all targets fail.
func (p *Pipeline) dispatchActualWithErr(ctx context.Context, message string, incidentIDs []string) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
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
	var lastWAErr error

	for _, incID := range incidentIDs {
		p.writeEvent(incID, "channel_attempt", "whatsapp", "Attempting WhatsApp notification...")
	}

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
					IncidentIDs: incidentIDs,
					Channel:     "WhatsApp",
					Recipient:   target.PhoneNumber, // Log with actual phone number
					Status:      "delivered",
					Timestamp:   time.Now(),
				}
				log.Printf("[Notifier] WhatsApp message delivered to %s", target.PhoneNumber)
				for _, incID := range incidentIDs {
					p.writeEvent(incID, "channel_delivered", "whatsapp", fmt.Sprintf("WhatsApp delivered successfully to %s", target.PhoneNumber))
					// Also record Telegram as skipped — always show both channels in the timeline
					if p.telegram != nil && p.telegram.chatID != "" {
						p.writeEvent(incID, "channel_skipped", "telegram", fmt.Sprintf("Telegram skipped — WhatsApp delivered successfully to %s (no fallback needed)", target.PhoneNumber))
					}
				}
			} else {
				log.Printf("[Notifier] WhatsApp failed for %s: %v", target.PhoneNumber, err)
				lastWAErr = err
				p.logChan <- NotificationLog{
					IncidentIDs: incidentIDs,
					Channel:     "WhatsApp",
					Recipient:   target.PhoneNumber,
					Status:      "failed",
					Error:       err.Error(),
					Timestamp:   time.Now(),
				}
				for _, incID := range incidentIDs {
					p.writeEvent(incID, "channel_failed", "whatsapp", fmt.Sprintf("WhatsApp failed: %s", err.Error()))
				}
			}
		}

		if whatsAppSent {
			tgRecipient := "@SanocBot"
			if p.telegram != nil && p.telegram.chatID != "" {
				tgRecipient = p.telegram.chatID
			}
			p.logChan <- NotificationLog{
				IncidentIDs: incidentIDs,
				Channel:     "Telegram",
				Recipient:   tgRecipient,
				Status:      "Skipped",
				Error:       "Skipped — WhatsApp delivered successfully (no fallback needed)",
				Timestamp:   time.Now(),
			}
		}
	} else {
		lastWAErr = fmt.Errorf("whatsapp sidecar is not healthy")
		for _, incID := range incidentIDs {
			p.writeEvent(incID, "channel_failed", "whatsapp", "WhatsApp failed: whatsapp sidecar is not healthy")
		}
	}

	if !whatsAppSent { // Fallback to Telegram only if no WhatsApp message was successfully sent
		log.Println("[Notifier] WhatsApp send failed. Falling back to Telegram.")
		for _, incID := range incidentIDs {
			p.writeEvent(incID, "channel_fallback", "telegram", "Falling back to Telegram...")
			p.writeEvent(incID, "channel_attempt", "telegram", "Attempting Telegram notification...")
		}

		// Telegram fallback: format with HTML <b>bold</b> tags
		tgMessage := FormatForTelegram(message)
		err := p.telegram.Send(tgMessage)
		status := "delivered"
		errStr := ""
		if err != nil {
			log.Printf("[Notifier] Telegram also failed: %v", err)
			status = "failed"
			errStr = err.Error()
			p.logChan <- NotificationLog{
				IncidentIDs: incidentIDs,
				Channel:     "Telegram",
				Recipient:   p.telegram.chatID, // Use the configured Telegram chat ID for logging
				Status:      status,
				Error:       errStr,
				Timestamp:   time.Now(),
			}
			for _, incID := range incidentIDs {
				p.writeEvent(incID, "channel_failed", "telegram", fmt.Sprintf("Telegram failed: %s", err.Error()))
			}
			return fmt.Errorf("whatsapp failed (%v) and telegram fallback failed (%v)", lastWAErr, err)
		}
		p.logChan <- NotificationLog{
			IncidentIDs: incidentIDs,
			Channel:     "Telegram",
			Recipient:   p.telegram.chatID,
			Status:      status,
			Timestamp:   time.Now(),
		}
		for _, incID := range incidentIDs {
			p.writeEvent(incID, "channel_delivered", "telegram", fmt.Sprintf("Telegram delivered successfully to channel %s", p.telegram.chatID))
		}
	}
	return nil
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
			if len(entry.IncidentIDs) > 0 {
				for _, incID := range entry.IncidentIDs {
					_ = p.notifLogRepo.Append(&domain.NotificationLogRow{
						IncidentID:  incID,
						Channel:     entry.Channel,
						Recipient:   entry.Recipient,
						Status:      entry.Status,
						ChannelIcon: entry.Error,
					})
				}
			} else {
				_ = p.notifLogRepo.Append(&domain.NotificationLogRow{
					Channel:     entry.Channel,
					Recipient:   entry.Recipient,
					Status:      entry.Status,
					ChannelIcon: entry.Error,
				})
			}
		}
	}
}
