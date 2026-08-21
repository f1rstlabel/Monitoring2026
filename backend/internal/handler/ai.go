package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sanoc/backend/internal/ai"
	"sanoc/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type AIChatRequest struct {
	Prompt  string           `json:"prompt" binding:"required"`
	History []ai.ChatMessage `json:"history"`
}

// GetAIStatus checks whether the Gemini AI service is ready and configured in backend .env
func (h *Handler) GetAIStatus(c *gin.Context) {
	configured := h.aiService != nil && h.aiService.IsConfigured()
	model := ""
	if h.aiService != nil {
		model = h.aiService.GetModel()
	}

	c.JSON(http.StatusOK, gin.H{
		"configured": configured,
		"model":      model,
		"provider":   "Google Gemini",
	})
}

// ChatAI processes an interactive question from the user with full live network context
func (h *Handler) ChatAI(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Gemini AI Copilot belum dikonfigurasi. Silakan isi variabel GEMINI_API_KEY pada file backend/.env",
		})
		return
	}

	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt tidak boleh kosong"})
		return
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt tidak boleh kosong"})
		return
	}

	reply, err := h.aiService.Chat(c.Request.Context(), prompt, req.History)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Gagal menghubungi AI Copilot: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":     reply,
		"timestamp": time.Now().Format("2006-01-02 15:04:05 WIB"),
		"model":     h.aiService.GetModel(),
	})
}

// AnalyzeIncidentAI performs automated Root Cause Analysis and troubleshooting for a specific incident
func (h *Handler) AnalyzeIncidentAI(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Gemini AI Copilot belum dikonfigurasi. Silakan isi variabel GEMINI_API_KEY pada file backend/.env",
		})
		return
	}

	incidentID := c.Param("id")
	if incidentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incident ID is required"})
		return
	}

	analysis, err := h.aiService.AnalyzeIncident(c.Request.Context(), incidentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Gagal menganalisis insiden: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incidentId": incidentID,
		"analysis":   analysis,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05 WIB"),
	})
}

// GetAIQuickPrompts returns smart contextual suggestions based on current network health
func (h *Handler) GetAIQuickPrompts(c *gin.Context) {
	var prompts []string

	// Check if there are down devices
	var downDevices []string
	if h.deviceRepo != nil {
		if devices, err := h.deviceRepo.GetAll("", "", ""); err == nil {
			for _, d := range devices {
				if d.Status == domain.StatusDOWN {
					downDevices = append(downDevices, d.Name)
				}
			}
		}
	}

	if len(downDevices) > 0 {
		prompts = append(prompts, fmt.Sprintf("Kenapa %s mengalami down?", downDevices[0]))
		prompts = append(prompts, "Rangkum seluruh perangkat yang down hari ini")
		if len(downDevices) > 1 {
			prompts = append(prompts, fmt.Sprintf("Apakah ada korelasi antara %s dan %s?", downDevices[0], downDevices[1]))
		}
	} else {
		prompts = append(prompts, "Bagaimana status ketersediaan dan SLA jaringan hari ini?")
		prompts = append(prompts, "Apakah ada perangkat yang mengalami flapping minggu ini?")
	}

	prompts = append(prompts, "Buatkan draf format laporan insiden WhatsApp untuk pimpinan")
	prompts = append(prompts, "Tampilkan daftar rekomendasi pengecekan preventif perangkat core")

	c.JSON(http.StatusOK, gin.H{
		"prompts": prompts,
	})
}
