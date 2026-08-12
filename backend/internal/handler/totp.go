package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"
)

// GenerateRandomBase32Secret generates a random 16-character base32 secret.
func GenerateRandomBase32Secret() (string, error) {
	randomBytes := make([]byte, 10)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	secret := base32.StdEncoding.EncodeToString(randomBytes)
	return strings.TrimRight(secret, "="), nil
}

// GenerateTOTPCode generates a 6-digit TOTP code for a given timestamp.
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	// Clean spaces and convert to uppercase
	secretUpper := strings.ToUpper(strings.TrimSpace(secret))
	secretUpper = strings.ReplaceAll(secretUpper, " ", "")
	secretUpper = strings.ReplaceAll(secretUpper, "-", "")

	if missingPadding := len(secretUpper) % 8; missingPadding != 0 {
		secretUpper += strings.Repeat("=", 8-missingPadding)
	}

	key, err := base32.StdEncoding.DecodeString(secretUpper)
	if err != nil {
		return "", err
	}

	counter := uint64(t.Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	h := hmac.New(sha1.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	truncatedHash := binary.BigEndian.Uint32(hash[offset : offset+4])
	truncatedHash &= 0x7FFFFFFF

	code := truncatedHash % uint32(math.Pow10(6))
	return fmt.Sprintf("%06d", code), nil
}

// VerifyTOTPCode checks if code matches TOTP for current time or adjacent windows (time skew tolerance up to ±90s).
func VerifyTOTPCode(secret, code string) bool {
	cleanSecret := strings.ReplaceAll(strings.TrimSpace(secret), " ", "")
	cleanCode := strings.ReplaceAll(strings.TrimSpace(code), " ", "")
	cleanCode = strings.ReplaceAll(cleanCode, "-", "")

	if len(cleanCode) != 6 {
		return false
	}

	now := time.Now()
	// Check 7 time windows (-90s to +90s) to account for system clock drift between server and phone
	windows := []time.Time{
		now.Add(-90 * time.Second),
		now.Add(-60 * time.Second),
		now.Add(-30 * time.Second),
		now,
		now.Add(30 * time.Second),
		now.Add(60 * time.Second),
		now.Add(90 * time.Second),
	}

	for _, w := range windows {
		expected, err := GenerateTOTPCode(cleanSecret, w)
		if err == nil && expected == cleanCode {
			return true
		}
	}
	return false
}

// GetTOTPURI generates the otpauth:// URI for QR code generators.
func GetTOTPURI(secret, accountName, issuer string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.PathEscape(issuer),
		url.PathEscape(accountName),
		secret,
		url.QueryEscape(issuer),
	)
}
