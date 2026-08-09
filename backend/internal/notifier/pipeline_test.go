package notifier

import (
	"testing"
)

func TestFormatForWhatsApp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML bold and code conversion",
			input:    "🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>Router Utama</b>\nIP: <code>192.168.1.1</code>",
			expected: "🔴 *INCIDENT — Device DOWN*\nDevice: *Router Utama*\nIP: `192.168.1.1`",
		},
		{
			name:     "Double asterisk markdown conversion",
			input:    "🔴 **INCIDENT — Device DOWN**\nDevice: **Switch Core**",
			expected: "🔴 *INCIDENT — Device DOWN*\nDevice: *Switch Core*",
		},
		{
			name:     "Italic HTML tag conversion",
			input:    "Status: <i>critical</i>",
			expected: "Status: _critical_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatForWhatsApp(tt.input)
			if result != tt.expected {
				t.Errorf("FormatForWhatsApp() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFormatForTelegram(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Preserve HTML bold and code",
			input:    "🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>Router Utama</b>\nIP: <code>192.168.1.1</code>",
			expected: "🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>Router Utama</b>\nIP: <code>192.168.1.1</code>",
		},
		{
			name:     "Convert WhatsApp single asterisk markdown to Telegram HTML",
			input:    "🔴 *INCIDENT — Device DOWN*\nDevice: *Switch Core*\nIP: `10.0.0.1`",
			expected: "🔴 <b>INCIDENT — Device DOWN</b>\nDevice: <b>Switch Core</b>\nIP: <code>10.0.0.1</code>",
		},
		{
			name:     "Convert double asterisk markdown to Telegram HTML",
			input:    "🔴 **INCIDENT — Device DOWN**",
			expected: "🔴 <b>INCIDENT — Device DOWN</b>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatForTelegram(tt.input)
			if result != tt.expected {
				t.Errorf("FormatForTelegram() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
