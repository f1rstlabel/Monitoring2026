package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	repo := &SettingsRepository{db: db}
	repo.EnsureTable()
	return repo
}

func (r *SettingsRepository) EnsureTable() {
	if r.db == nil {
		return
	}
	query := `
	CREATE TABLE IF NOT EXISTS integration_settings (
		key VARCHAR(64) PRIMARY KEY,
		value JSONB NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	INSERT INTO integration_settings (key, value) VALUES
	('telegram', '{"bot_token": "", "chat_id": ""}'::jsonb),
	('whatsapp', '{"status": "disconnected", "linked_number": ""}'::jsonb),
	('system_settings', '{"rateLimitMaxMsgPerMin": 60, "polling": {"intervalSeconds": 15, "concurrencyBatchSize": 50}, "thresholds": [{"type": "Access Point", "consecutiveFailures": 3}, {"type": "Switch", "consecutiveFailures": 2}, {"type": "Router", "consecutiveFailures": 2}, {"type": "SmartPower", "consecutiveFailures": 4}, {"type": "CCTV", "consecutiveFailures": 5}, {"type": "NVR", "consecutiveFailures": 3}]}'::jsonb)
	ON CONFLICT (key) DO NOTHING;
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("[SettingsRepo] Table initialization check: %v", err)
	}
}

func (r *SettingsRepository) GetJSON(key string, target interface{}) error {
	if r.db == nil {
		return fmt.Errorf("no db connection")
	}
	query := `SELECT value FROM integration_settings WHERE key = $1`
	var raw []byte
	err := r.db.QueryRow(query, key).Scan(&raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func (r *SettingsRepository) SetJSON(key string, val interface{}) error {
	if r.db == nil {
		return nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	query := `
	INSERT INTO integration_settings (key, value, updated_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`
	_, err = r.db.Exec(query, key, data)
	return err
}
