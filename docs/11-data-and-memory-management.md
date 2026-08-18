# 11 — Data & Memory Management / Manajemen Data & Memori

This document explains how state, persistent data, and in-memory caches are managed across the services of **SANOC v2.6.0**, detailing service restart recovery behaviors.  
*Dokumen ini menjelaskan manajemen data, memori, dan retensi pada SANOC.*

---

## 1. Data Tier & Persistence Storage Breakdown

```mermaid
flowchart TD
    subgraph PersistentTier ["1. Persistent Storage (PostgreSQL 14+)"]
        PG_Users["users & user_logs"]
        PG_Devices["devices & device_status_logs"]
        PG_Telemetry["device_metrics (SNMP OID data)"]
        PG_Incidents["incidents & notification_logs"]
        PG_Integrations["whatsapp_targets & integration_settings"]
    end

    subgraph FileStorage ["2. Static Uploads & Disk Sessions"]
        F_Uploads["backend/uploads/ (User Avatars - Magic Byte Checked)"]
        F_Sessions["backend/wa-sidecar/auth_sessions/ (Baileys Session)"]
    end

    subgraph CacheTier ["3. In-Memory Cache & Pub/Sub (Redis 6+)"]
        R_Status["device:status:<id> (Live Status Cache)"]
        R_Buffer["notif:buffer (Notification Batch List)"]
        R_Rate["ratelimit:wa (Token Bucket Rate Limiter)"]
    end

    subgraph PollerMemory ["4. Poller Runtime Memory (Go Process)"]
        P_Counters["Fail Counters (map[string]int)"]
        P_Debounce["Debounce Timers & State Flags"]
        P_ARP["ARP Table Cache (MAC-to-IP Mappings)"]
    end

    subgraph WSMemory ["5. WebSocket Gateway Memory (Go Process)"]
        WS_Clients["Client Registry (map[*Client]bool)"]
    end
```

---

## 2. Storage Layer Responsibilities

### 2.1 PostgreSQL (Relational Persistence)
- **Primary Source of Truth**: Stores immutable domain data including user accounts, usernames, TOTP MFA secrets, role permissions, device inventory, incident timelines, and audit logs.
- **Durability**: Survives system reboots, hardware crashes, and container redeployments.

### 2.2 Redis (Ephemeral Cache & Queue Buffer)
- **Live Device Status Cache** (`device:status:<id>`): High-speed cache hit by API handlers to render dashboard grids.
- **Notification Aggregation Buffer** (`notif:buffer`): List storing queued alert jobs during multi-device outages.

### 2.3 Local Disk Storage (`backend/uploads/` & `auth_sessions/`)
- **User Avatar Storage**: Images uploaded via `/profile` stored under `backend/uploads/` with sanitized names (`filepath.Base`) and magic-byte security checks.
- **WhatsApp Web Session**: Session keys stored on disk to preserve QR Code pairing across container restarts.
