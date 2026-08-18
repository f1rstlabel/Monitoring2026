package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"sanoc/backend/internal/config"
)

type Mailer struct {
	cfg *config.Config
}

func NewMailer(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendOTPEmail(recipientEmail, otpCode, purpose string) error {
	recipientEmail = strings.TrimSpace(recipientEmail)
	if recipientEmail == "" {
		return fmt.Errorf("recipient email is required")
	}

	// Always ensure fresh configuration
	if m.cfg == nil || m.cfg.SMTPUser == "" || m.cfg.SMTPPassword == "" {
		freshCfg := config.LoadConfig()
		if freshCfg != nil {
			m.cfg = freshCfg
		}
	}

	subject := "Kode Verifikasi Pendaftaran Akun SANOC"
	headerTitle := "Verifikasi Pendaftaran Akun"
	descText := "Gunakan kode verifikasi berikut untuk menyelesaikan pendaftaran akun Anda di sistem SANOC Jabar Monitoring."
	if purpose == "reset_password" {
		subject = "Kode OTP Reset Kata Sandi Akun SANOC"
		headerTitle = "Reset Kata Sandi Akun"
		descText = "Kami menerima permintaan untuk mereset kata sandi akun SANOC Anda. Gunakan kode verifikasi di bawah ini:"
	}

	from := m.cfg.SMTPFrom
	if from == "" {
		from = "SANOC Monitoring <noreply@jabarprov.go.id>"
	}

	// Build HTML email body
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <title>%s</title>
</head>
<body style="margin: 0; padding: 0; background-color: #0A0A0B; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #E4E4E7;">
  <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #0A0A0B; padding: 30px 15px;">
    <tr>
      <td align="center">
        <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 540px; background-color: #151517; border: 1px solid #26262A; border-radius: 16px; overflow: hidden; box-shadow: 0 10px 25px rgba(0,0,0,0.5);">
          <!-- Header Banner -->
          <tr>
            <td style="padding: 28px 32px; background: linear-gradient(135deg, #18181B 0%%, #121214 100%%); border-bottom: 1px solid #26262A; text-align: center;">
              <div style="display: inline-block; padding: 8px 16px; background-color: rgba(123, 150, 245, 0.12); border: 1px solid rgba(123, 150, 245, 0.3); border-radius: 20px; font-family: monospace; font-size: 11px; font-weight: bold; color: #7B96F5; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 12px;">
                SANOC JABAR MONITORING
              </div>
              <h1 style="margin: 0; font-size: 20px; font-weight: 800; color: #FFFFFF; letter-spacing: -0.5px;">
                %s
              </h1>
            </td>
          </tr>
          
          <!-- Content Body -->
          <tr>
            <td style="padding: 32px;">
              <p style="margin: 0 0 16px 0; font-size: 14px; line-height: 1.6; color: #A1A1AA;">
                Halo,
              </p>
              <p style="margin: 0 0 24px 0; font-size: 14px; line-height: 1.6; color: #D4D4D8;">
                %s
              </p>
              
              <!-- 6-Digit OTP Box -->
              <div style="background-color: #18181B; border: 1px solid #3ECF8E; border-radius: 12px; padding: 20px; text-align: center; margin: 24px 0;">
                <div style="font-size: 11px; font-family: monospace; color: #3ECF8E; font-weight: bold; text-transform: uppercase; letter-spacing: 1.5px; margin-bottom: 8px;">
                  KODE VERIFIKASI OTP
                </div>
                <div style="font-size: 32px; font-family: monospace; font-weight: 800; color: #FFFFFF; letter-spacing: 8px; text-shadow: 0 0 10px rgba(62, 207, 142, 0.3);">
                  %s
                </div>
                <div style="font-size: 11px; color: #71717A; margin-top: 8px;">
                  Masa berlaku kode: <strong>15 Menit</strong>
                </div>
              </div>
              
              <p style="margin: 24px 0 0 0; font-size: 12px; line-height: 1.6; color: #71717A; border-top: 1px solid #26262A; padding-top: 16px;">
                ⚠️ <strong>Penting:</strong> Jangan berikan kode ini kepada siapa pun. Tim SANOC tidak akan pernah meminta kode verifikasi Anda. Jika Anda tidak merasa melakukan tindakan ini, abaikan email ini.
              </p>
            </td>
          </tr>
          
          <!-- Footer -->
          <tr>
            <td style="padding: 20px 32px; background-color: #0E0E10; border-top: 1px solid #26262A; text-align: center; font-size: 11px; color: #52525B;">
              Diskominfo Pemerintah Daerah Provinsi Jawa Barat &bull; Sistem Monitoring SANOC &copy; 2026
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, subject, headerTitle, descText, otpCode)

	// Extract plain sender email for SMTP MAIL FROM
	fromEmail := from
	if strings.Contains(from, "<") && strings.Contains(from, ">") {
		start := strings.Index(from, "<")
		end := strings.Index(from, ">")
		if start < end {
			fromEmail = from[start+1 : end]
		}
	}

	// Check if SMTP is configured
	if m.cfg.SMTPHost == "" {
		log.Printf("[MAILER_NOTICE] SMTP_HOST not configured in .env. OTP for %s: %s", recipientEmail, otpCode)
		return nil
	}

	smtpUser := strings.TrimSpace(m.cfg.SMTPUser)
	smtpPass := strings.TrimSpace(m.cfg.SMTPPassword)
	if strings.Contains(m.cfg.SMTPHost, "gmail.com") {
		smtpPass = strings.ReplaceAll(smtpPass, " ", "")
	}

	envelopeFrom := fromEmail
	if smtpUser != "" {
		envelopeFrom = smtpUser
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = recipientEmail
	headers["Reply-To"] = "SANOC No-Reply <sanoc.noreply@gmail.com>"
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	log.Printf("[MAILER_DISPATCH] Dispatching OTP [%s] email -> To: %s (From: %s, Envelope: %s)", otpCode, recipientEmail, from, envelopeFrom)

	smtpPort := m.cfg.SMTPPort
	if smtpPort == "" {
		smtpPort = "587"
	}
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, smtpPort)

	// Case 1: Authenticated SMTP (e.g. Gmail / Mailgun / TLS)
	if smtpUser != "" && smtpPass != "" {
		auth := smtp.PlainAuth("", smtpUser, smtpPass, m.cfg.SMTPHost)
		portInt, _ := strconv.Atoi(smtpPort)

		// Helper function for SSL 465
		sendViaSSL := func() error {
			tlsConfig := &tls.Config{
				ServerName: m.cfg.SMTPHost,
			}
			conn, err := tls.Dial("tcp", m.cfg.SMTPHost+":465", tlsConfig)
			if err != nil {
				log.Printf("[MAILER_ERROR] TLS Dial failed to %s:465: %v", m.cfg.SMTPHost, err)
				return err
			}
			defer conn.Close()

			client, err := smtp.NewClient(conn, m.cfg.SMTPHost)
			if err != nil {
				log.Printf("[MAILER_ERROR] SMTP NewClient failed: %v", err)
				return err
			}
			defer client.Quit()

			if err := client.Auth(auth); err != nil {
				log.Printf("[MAILER_ERROR] SMTP Auth failed: %v", err)
				return err
			}
			if err := client.Mail(envelopeFrom); err != nil {
				return err
			}
			if err := client.Rcpt(recipientEmail); err != nil {
				return err
			}
			w, err := client.Data()
			if err != nil {
				return err
			}
			if _, err = w.Write([]byte(message)); err != nil {
				return err
			}
			_ = w.Close()
			log.Printf("[MAILER_SUCCESS] Real email sent via SSL (465) to %s", recipientEmail)
			return nil
		}

		if portInt == 465 {
			return sendViaSSL()
		}

		// Try Standard STARTTLS connection (Port 587 / 25)
		err := smtp.SendMail(addr, auth, envelopeFrom, []string{recipientEmail}, []byte(message))
		if err != nil {
			log.Printf("[MAILER_WARNING] SendMail via %s failed (%v), attempting SSL fallback on port 465...", addr, err)
			sslErr := sendViaSSL()
			if sslErr == nil {
				return nil
			}
			log.Printf("[MAILER_ERROR] Both STARTTLS and SSL sending failed to %s: %v", recipientEmail, err)
			return err
		}

		log.Printf("[MAILER_SUCCESS] Real email sent successfully to %s", recipientEmail)
		return nil
	}

	// Case 2: Local SMTP / Mailpit (e.g. Laragon Mailpit 127.0.0.1:1025 without auth)
	c, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("[MAILER_ERROR] Failed to connect to Mailpit/SMTP server at %s: %v", addr, err)
		return err
	}
	defer c.Quit()

	if err := c.Mail(fromEmail); err != nil {
		log.Printf("[MAILER_ERROR] Mail command failed: %v", err)
		return err
	}
	if err := c.Rcpt(recipientEmail); err != nil {
		log.Printf("[MAILER_ERROR] Rcpt command failed: %v", err)
		return err
	}
	wc, err := c.Data()
	if err != nil {
		log.Printf("[MAILER_ERROR] Data command failed: %v", err)
		return err
	}
	if _, err := wc.Write([]byte(message)); err != nil {
		log.Printf("[MAILER_ERROR] Write email body failed: %v", err)
		return err
	}
	if err := wc.Close(); err != nil {
		log.Printf("[MAILER_ERROR] Close email writer failed: %v", err)
		return err
	}

	log.Printf("[MAILER_SUCCESS] Real email delivered to Mailpit (%s) for %s [OTP: %s]", addr, recipientEmail, otpCode)
	return nil
}
