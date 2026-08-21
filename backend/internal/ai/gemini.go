package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/repository"
)

type ChatMessage struct {
	Role    string `json:"role"`    // "user" | "model"
	Content string `json:"content"`
}

type Service struct {
	apiKey       string
	model        string
	deviceRepo   repository.DeviceRepository
	incidentRepo repository.IncidentRepository
	statusRepo   repository.StatusLogRepository
	httpClient   *http.Client
}

func NewService(
	apiKey string,
	model string,
	deviceRepo repository.DeviceRepository,
	incidentRepo repository.IncidentRepository,
	statusRepo repository.StatusLogRepository,
) *Service {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &Service{
		apiKey:       strings.TrimSpace(apiKey),
		model:        model,
		deviceRepo:   deviceRepo,
		incidentRepo: incidentRepo,
		statusRepo:   statusRepo,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (s *Service) IsConfigured() bool {
	return s.apiKey != ""
}

func (s *Service) GetModel() string {
	return s.model
}

// BuildTelemetryContext extracts live network state to inject into the AI prompt
func (s *Service) BuildTelemetryContext(query string) string {
	var sb strings.Builder

	sb.WriteString("[TELEMETRI SANOC LIVE]:\n")
	sb.WriteString(fmt.Sprintf("- Waktu: %s\n", time.Now().Format("15:04:05 WIB")))

	// Fetch all devices
	if s.deviceRepo != nil {
		devices, err := s.deviceRepo.GetAll("", "", "")
		if err == nil && len(devices) > 0 {
			var upCount, downCount int
			var downDevices []string

			for _, d := range devices {
				switch d.Status {
				case domain.StatusUP:
					upCount++
				case domain.StatusDOWN:
					downCount++
					downDevices = append(downDevices, fmt.Sprintf("%s (IP: %s, Lokasi: %s)", d.Name, d.IP, d.Location))
				}
			}

			sb.WriteString(fmt.Sprintf("- Total: %d perangkat (UP: %d, DOWN: %d)\n", len(devices), upCount, downCount))

			if len(downDevices) > 0 {
				limit := 10
				if len(downDevices) < limit {
					limit = len(downDevices)
				}
				sb.WriteString(fmt.Sprintf("- Ringkasan DOWN (%d unit): %s\n", len(downDevices), strings.Join(downDevices[:limit], "; ")))
			} else {
				sb.WriteString("- Status DOWN: Tidak ada (Semua perangkat UP 100%)\n")
			}

			// Targeted search if query mentions an IP or device name
			lowerQ := strings.ToLower(query)
			for _, d := range devices {
				if (d.IP != "" && strings.Contains(lowerQ, strings.ToLower(d.IP))) ||
					(d.Name != "" && len(d.Name) > 2 && strings.Contains(lowerQ, strings.ToLower(d.Name))) {
					sb.WriteString(fmt.Sprintf("- TARGET PERANGKAT DITANYAKAN: Name=%q, IP=%s, Tipe=%s, Status=%s, Lokasi=%s, Rack=%s, Checked=%s (%ds lalu)\n",
						d.Name, d.IP, d.Type, d.Status, d.Location, d.Rack, d.LastChecked, d.CheckedSecondsAgo))
					if d.SNMPSysName != "" {
						sb.WriteString(fmt.Sprintf("  SNMP Info: SysName=%q, SysDescr=%q, Uptime=%s\n", d.SNMPSysName, d.SNMPSysDescr, d.SNMPSysUpTime))
					}
					break
				}
			}
		}
	}

	// Fetch active incidents (compact)
	if s.incidentRepo != nil {
		incidents, err := s.incidentRepo.GetActive()
		if err == nil && len(incidents) > 0 {
			var incList []string
			for _, inc := range incidents {
				incList = append(incList, fmt.Sprintf("#%s %s (%s)", inc.ID, inc.DeviceName, inc.Status))
			}
			sb.WriteString(fmt.Sprintf("- Insiden Aktif (%d): %s\n", len(incidents), strings.Join(incList, ", ")))
		}
	}

	return sb.String()
}

// Chat executes a conversation with Google Gemini with injected context
func (s *Service) Chat(ctx context.Context, prompt string, history []ChatMessage) (string, error) {
	cleanKey := strings.TrimSpace(strings.Trim(s.apiKey, "\"'"))
	if cleanKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not configured in backend .env")
	}

	telemetryContext := s.BuildTelemetryContext(prompt)

	systemInstruction := `Anda adalah SANOC AI Copilot, asisten cerdas ahli jaringan & infrastruktur (Senior SANOC Infrastructure Specialist) untuk sistem pemantauan Sanditel Jabar.

PANDUAN KOMUNIKASI & RESPON:
1. IDENTITAS: Anda bernama SANOC AI Copilot, asisten cerdas resmi sistem monitoring SANOC (Smart Agent and Network Operations Center) Sanditel Jabar.
2. GAYA BAHASA: Ramah, profesional, cerdas, solutif, menggunakan Bahasa Indonesia yang baik (istilah teknis networking tetap dipertahankan seperti ping, latency, gateway, ARP, switch, VLAN, dll).
3. KELUWESAN PERCAKAPAN (VERSATILITY):
   - Jika pengguna menyapa (misal "halo", "hai", "selamat pagi"), balaslah dengan hangat dan tanyakan apa yang bisa dibantu untuk pemantauan SANOC.
   - Jika pengguna bertanya hal umum (konsep jaringan, cara routing OSPF/BGP, subnetting IP, konfigurasi Mikrotik/Cisco, perintah Linux/Bash, database, atau obrolan santai), jawablah dengan jelas dan ramah seperti rekan kerja senior IT.
   - Jika pengguna bertanya spesifik mengenai status perangkat/jaringan SANOC atau insiden, manfaatkan data telemetri live yang disertakan di atas.
4. FORMAT PENULISAN (MARKDOWN RAPI):
   - Gunakan heading terstruktur (### Judul Bagian) untuk analisis panjang.
   - Gunakan bullet points (-) atau penomoran (1., 2.) agar mudah dibaca sekilas oleh teknisi SANOC.
   - Tebalkan (**kata kunci**, **IP**, **Nama Perangkat**).
   - Gunakan code block (seperti perintah 'ping 10.11.x.x', 'traceroute', dll) untuk instruksi terminal.
5. KECEPATAN & KEPADATAN: Berikan jawaban yang padat, akurat, dan langsung ke poin (tidak bertele-tele).`

	rawModel := strings.TrimSpace(strings.Trim(s.model, "\"'"))
	rawModel = strings.TrimPrefix(rawModel, "models/")
	if rawModel == "" {
		rawModel = "gemini-2.0-flash"
	}

	// Model candidates with fallback support
	modelList := []string{rawModel}
	for _, fb := range []string{"gemini-2.0-flash", "gemini-1.5-flash-latest", "gemini-1.5-pro", "gemini-1.5-flash", "gemini-2.5-flash"} {
		if fb != rawModel {
			modelList = append(modelList, fb)
		}
	}

	// Build contents array for Gemini API
	var contents []map[string]interface{}

	// Add history
	for _, h := range history {
		role := h.Role
		if role == "assistant" || role == "ai" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{"text": h.Content},
			},
		})
	}

	// Add current user prompt augmented with telemetry context
	userMessageWithContext := fmt.Sprintf("%s\n\n[USER QUESTION / REQUEST]:\n%s", telemetryContext, prompt)
	contents = append(contents, map[string]interface{}{
		"role": "user",
		"parts": []map[string]interface{}{
			{"text": userMessageWithContext},
		},
	})

	reqBody := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemInstruction},
			},
		},
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     0.3,
			"topP":            0.8,
			"maxOutputTokens": 2048,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error

	// Try candidate models and API versions until success or definite non-404 error
	for _, candidateModel := range modelList {
		for _, apiVer := range []string{"v1beta", "v1"} {
			url := fmt.Sprintf("https://generativelanguage.googleapis.com/%s/models/%s:generateContent?key=%s", apiVer, candidateModel, cleanKey)

			httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
			if err != nil {
				return "", fmt.Errorf("failed to create request: %w", err)
			}
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := s.httpClient.Do(httpReq)
			if err != nil {
				return "", fmt.Errorf("Gemini API connection failed: %w", err)
			}

			respBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return "", fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode == http.StatusOK {
				var geminiResp struct {
					Candidates []struct {
						Content struct {
							Parts []struct {
								Text string `json:"text"`
							} `json:"parts"`
						} `json:"content"`
						FinishReason string `json:"finishReason"`
					} `json:"candidates"`
				}

				if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
					return "", fmt.Errorf("failed to parse Gemini response: %w", err)
				}

				if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
					return "Maaf, AI tidak dapat menghasilkan jawaban untuk permintaan ini.", nil
				}

				s.model = candidateModel
				return geminiResp.Candidates[0].Content.Parts[0].Text, nil
			}

			// If it's a 404 (model not found on this version), try next model/version
			if resp.StatusCode == http.StatusNotFound {
				lastErr = fmt.Errorf("Gemini API error (HTTP 404): %s", string(respBytes))
				continue
			}

			// For any other error (e.g. 400 Bad Request, 403 Invalid API Key, 429 Rate Limit), return immediately
			log.Printf("[AI Gemini Error] HTTP %d with model %s: %s", resp.StatusCode, candidateModel, string(respBytes))
			return "", fmt.Errorf("Gemini API error (HTTP %d): %s", resp.StatusCode, string(respBytes))
		}
	}

	return "", lastErr
}

// AnalyzeIncident performs deep RCA and step-by-step resolution checklist for an incident
func (s *Service) AnalyzeIncident(ctx context.Context, incidentID string) (string, error) {
	if !s.IsConfigured() {
		return "", fmt.Errorf("GEMINI_API_KEY is not configured in backend .env")
	}

	if s.incidentRepo == nil {
		return "", fmt.Errorf("incident repository is nil")
	}

	inc, err := s.incidentRepo.GetByID(incidentID)
	if err != nil || inc == nil {
		return "", fmt.Errorf("incident not found: %s", incidentID)
	}

	events, _ := s.incidentRepo.GetEventsByIncidentID(incidentID)

	prompt := fmt.Sprintf(`Lakukan analisis mendalam (Root Cause Analysis - RCA), penilaian dampak keparahan, dan panduan investigasi langkah-demi-langkah untuk insiden berikut:

[DATA INSIDEN]:
- Incident ID: #%s
- Perangkat: %s (Tipe: %s, IP: %s)
- Lokasi: %s
- Status Insiden: %s
- Waktu Mulai: %s (Durasi: %s)
- Jumlah Perangkat Terdampak: %d
- Timeline Kejadian:
%s

Berikan output terstruktur dalam format Markdown rapi yang mencakup:
1. 🔍 **Ringkasan & Diagnosis Masalah**
2. ⚠️ **Analisis Akar Masalah (Root Cause Analysis)**
3. 🛠️ **Checklist Tindakan Perbaikan Teknis SANOC (Step-by-Step)**
4. 📱 **Draf Format Laporan WhatsApp Tim SANOC / Pimpinan**`,
		inc.ID, inc.DeviceName, inc.DeviceType, inc.DeviceIP, inc.Location, inc.Status,
		inc.StartedAt.Format("2006-01-02 15:04:05 WIB"), inc.Duration, inc.AffectedDevicesCount,
		formatEvents(events),
	)

	return s.Chat(ctx, prompt, nil)
}

func formatEvents(events []domain.IncidentEvent) string {
	if len(events) == 0 {
		return "  (Belum ada catatan event tambahan)"
	}
	var sb strings.Builder
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("  • [%s] %s: %s\n", e.OccurredAt.Format("15:04:05"), e.EventType, e.Detail))
	}
	return sb.String()
}
