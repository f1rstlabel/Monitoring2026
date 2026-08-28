package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"sanoc/backend/internal/ai"
	"sanoc/backend/internal/domain"
	"sanoc/backend/internal/notifier"
	"sanoc/backend/internal/repository"
)

const flapThreshold = 5 // downs in 7 days before flagging as recurring issue

// FlapReportJob scans the status log for flapping devices and sends a summary.
type FlapReportJob struct {
	statusRepo repository.StatusLogRepository
	deviceRepo repository.DeviceRepository
	pipeline   *notifier.Pipeline
	aiService  *ai.Service
}

func NewFlapReportJob(
	statusRepo repository.StatusLogRepository,
	deviceRepo repository.DeviceRepository,
	pipeline *notifier.Pipeline,
	aiService *ai.Service,
) *FlapReportJob {
	return &FlapReportJob{
		statusRepo: statusRepo,
		deviceRepo: deviceRepo,
		pipeline:   pipeline,
		aiService:  aiService,
	}
}

// RunWeekly starts a goroutine that fires weekly on Monday at 07:00 WIB.
func (j *FlapReportJob) RunWeekly() {
	go func() {
		for {
			now := time.Now()
			// Calculate next 07:00 local time
			next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())
			
			// Advance to next Monday if today is not Monday, or if it is Monday but past 07:00
			for next.Weekday() != time.Monday || now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			
			wait := next.Sub(now)
			log.Printf("[FlapReport] Next run in %v (at %s)", wait.Round(time.Minute), next.Format("2006-01-02 15:04"))
			time.Sleep(wait)
			j.Execute()
		}
	}()
}

// Execute runs the flap report immediately (also callable manually).
func (j *FlapReportJob) Execute() {
	log.Println("[FlapReport] Running weekly flap detection scan...")

	to := time.Now()
	from := to.Add(-7 * 24 * time.Hour)

	reports, err := j.statusRepo.GetFlapDevices(flapThreshold, from, to)
	if err != nil {
		log.Printf("[FlapReport] Error scanning status log: %v", err)
		return
	}

	if len(reports) == 0 {
		log.Println("[FlapReport] No recurring issues detected — no notification sent")
		return
	}

	msg := buildFlapSummary(reports)

	if j.aiService != nil && j.aiService.IsConfigured() {
		log.Println("[FlapReport] Analyzing reports with AI...")
		aiAnalysis, err := j.aiService.AnalyzeFlapReport(context.Background(), reports)
		if err == nil && aiAnalysis != "" {
			msg += "\n\n🤖 **Analisis AI Pintar:**\n" + aiAnalysis
		} else {
			log.Printf("[FlapReport] AI Analysis failed: %v", err)
		}
	}

	log.Printf("[FlapReport] Found %d recurring devices — sending summary", len(reports))
	j.pipeline.Send(context.Background(), msg, nil)
}

func buildFlapSummary(reports []domain.FlapReport) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚠️ <b>RECURRING ISSUE REPORT</b> — %d devices with frequent outages:\n", len(reports)))
	for _, r := range reports {
		sb.WriteString(fmt.Sprintf("\n• <b>%s</b> (%s)\n  IP: %s | Location: %s\n  Down %d times in 7d, total %d min downtime.\n",
			r.DeviceName, r.DeviceType, r.IP, r.Location, r.DownCount7d, r.TotalDowntimeMinutes))
	}
	return sb.String()
}
