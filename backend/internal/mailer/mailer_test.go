package mailer_test

import (
	"testing"

	"sanoc/backend/internal/config"
	"sanoc/backend/internal/mailer"
)

func TestSendOTPEmailToTarget(t *testing.T) {
	cfg := config.LoadConfig()
	m := mailer.NewMailer(cfg)

	targetEmail := "candrasaputraagung@gmail.com"
	t.Logf("Dispatching test OTP to %s using configured SMTP...", targetEmail)

	err := m.SendOTPEmail(targetEmail, "889900", "register")
	if err != nil {
		t.Fatalf("Failed to send OTP to %s: %v", targetEmail, err)
	}
	t.Logf("SUCCESS: OTP email successfully delivered to %s", targetEmail)
}
