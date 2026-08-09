package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/notifier"
	"govmonitor-it/backend/internal/repository"
)

const flapThreshold = 5 // downs in 7 days before flagging as recurring issue

// FlapReportJob scans the status log for flapping devices and sends a summary.
type FlapReportJob struct {
	statusRepo repository.StatusLogRepository
	deviceRepo repository.DeviceRepository
	pipeline   *notifier.Pipeline
}

func NewFlapReportJob(
	statusRepo repository.StatusLogRepository,
	deviceRepo repository.DeviceRepository,
	pipeline *notifier.Pipeline,
) *FlapReportJob {
	return &FlapReportJob{
		statusRepo: statusRepo,
		deviceRepo: deviceRepo,
		pipeline:   pipeline,
	}
}

// RunDaily starts a goroutine that fires daily at 07:00 WIB (UTC+7 = UTC 00:00).
func (j *FlapReportJob) RunDaily() {
	go func() {
		for {
			now := time.Now()
			// Calculate next 07:00 local time
			next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())
			if now.After(next) {
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
	log.Println("[FlapReport] Running daily flap detection scan...")

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
