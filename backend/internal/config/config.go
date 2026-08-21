package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost string
	RedisPort string

	ServerPort string

	JWTSecret          string
	CORSAllowedOrigin  string
	CookieSecure       bool
	CookieSameSite     string
	RateLimitLoginMax  int
	RecaptchaEnabled   bool
	RecaptchaSecret    string

	WhatsAppTargetNumber string
	WhatsAppGatewayURL   string // base URL of the Baileys wa-sidecar
	WhatsAppAPIToken     string // deprecated field kept for compat
	WASidecarToken       string // internal auth token shared with wa-sidecar

	TelegramBotToken string
	TelegramChatID   string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	GeminiAPIKey string
	GeminiModel  string
}

func LoadConfig() *Config {
	envPaths := []string{".env", "backend/.env", "../.env", "../../.env"}
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Overload(p)
			break
		}
	}

	cookieSecure := getEnv("COOKIE_SECURE", "false") == "true"
	recaptchaEnabled := getEnv("RECAPTCHA_ENABLED", "false") == "true"
	rateLimitMax := 5
	if maxStr := getEnv("RATE_LIMIT_LOGIN_MAX", "5"); maxStr != "" {
		fmt.Sscanf(maxStr, "%d", &rateLimitMax)
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "sanoc"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		ServerPort: getEnv("PORT", "8080"),

		JWTSecret:         getEnv("JWT_SECRET", "sanoc-dev-secret-key-production-change-this"),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),
		CookieSecure:      cookieSecure,
		CookieSameSite:    getEnv("COOKIE_SAMESITE", "Strict"),
		RateLimitLoginMax: rateLimitMax,
		RecaptchaEnabled:  recaptchaEnabled,
		RecaptchaSecret:   getEnv("RECAPTCHA_SECRET_KEY", ""),

		WhatsAppTargetNumber: getEnv("WHATSAPP_TARGET_NUMBER", "+6281290008888"),
		WhatsAppGatewayURL:   getEnv("WHATSAPP_GATEWAY_URL", "http://localhost:3001"),
		WhatsAppAPIToken:     getEnv("WHATSAPP_API_TOKEN", ""),
		WASidecarToken:       getEnvWithFallback("WA_SIDECAR_TOKEN", "INTERNAL_TOKEN", ""),

		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "SANOC Jabar Monitoring <noreply@jabarprov.go.id>"),

		GeminiAPIKey: getEnvWithFallback("GEMINI_API_KEY", "GOOGLE_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
	}
}

func (c *Config) PostgresDSN() string {
	if c.DBPassword != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
	}
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvWithFallback(key1, key2, defaultVal string) string {
	if val, ok := os.LookupEnv(key1); ok && val != "" {
		return val
	}
	if val, ok := os.LookupEnv(key2); ok && val != "" {
		return val
	}
	return defaultVal
}
