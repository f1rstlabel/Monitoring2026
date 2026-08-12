# 10 — Entity Relationship Diagram (ERD) / Diagram Relasi Entitas

This document presents the exact PostgreSQL database schema for **SANOC v2.6.0** (including Migration `016_security_mfa_schema.up.sql`).  
*Dokumen ini menyajikan Diagram Relasi Entitas (ERD) PostgreSQL untuk SANOC.*

---

## 1. Mermaid Entity Relationship Diagram

```mermaid
erDiagram
    users {
        VARCHAR_64 id PK
        VARCHAR_255 username UK
        VARCHAR_255 name
        VARCHAR_255 email UK
        VARCHAR_255 password
        VARCHAR_50 role
        VARCHAR_50 status
        TEXT avatar_url
        VARCHAR_100 last_active
        JSONB permissions
        BOOLEAN is_active
        BOOLEAN mfa_enabled
        VARCHAR_255 mfa_secret
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    user_logs {
        VARCHAR_64 id PK
        VARCHAR_64 user_id FK
        VARCHAR_50 action
        TEXT detail
        VARCHAR_50 ip_address
        TEXT user_agent
        VARCHAR_128 session_id
        TIMESTAMP occurred_at
    }

    devices {
        VARCHAR_64 id PK
        VARCHAR_255 name
        VARCHAR_50 type
        VARCHAR_50 ip
        VARCHAR_50 mac UK
        VARCHAR_50 status
        VARCHAR_50 addressing_mode
        VARCHAR_255 location
        VARCHAR_255 rack
        BOOLEAN snmp_enabled
        VARCHAR_255 snmp_community
        INT snmp_port
        INT custom_threshold
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    incidents {
        VARCHAR_64 id PK
        VARCHAR_64 device_id FK
        VARCHAR_255 device_name
        VARCHAR_50 device_type
        VARCHAR_50 device_ip
        VARCHAR_50 status
        TIMESTAMP start_time
        TIMESTAMP resolved_at
        VARCHAR_100 duration
    }

    notification_logs {
        VARCHAR_64 id PK
        VARCHAR_64 incident_id FK
        VARCHAR_50 channel
        VARCHAR_255 recipient
        VARCHAR_50 status
        TEXT error_message
        TIMESTAMP sent_at
    }

    users ||--o{ user_logs : "generates"
    devices ||--o{ incidents : "triggers"
    incidents ||--o{ notification_logs : "dispatches"
```
