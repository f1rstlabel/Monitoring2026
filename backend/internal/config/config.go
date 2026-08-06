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

	WhatsAppTargetNumber string
	WhatsAppGatewayURL   string // base URL of the Baileys wa-sidecar
	WhatsAppAPIToken     string // deprecated field kept for compat
	WASidecarToken       string // internal auth token shared with wa-sidecar

	TelegramBotToken string
	TelegramChatID   string
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "govmonitor"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		ServerPort: getEnv("PORT", "8080"),

		JWTSecret:         getEnv("JWT_SECRET", "govmonitor-dev-secret-change-in-prod"),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),

		WhatsAppTargetNumber: getEnv("WHATSAPP_TARGET_NUMBER", "+6281290008888"),
		WhatsAppGatewayURL:   getEnv("WHATSAPP_GATEWAY_URL", "http://localhost:3001"),
		WhatsAppAPIToken:     getEnv("WHATSAPP_API_TOKEN", ""),
		WASidecarToken:       getEnvWithFallback("WA_SIDECAR_TOKEN", "INTERNAL_TOKEN", ""),

		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
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
