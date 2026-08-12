# 01 — System Architecture / Arsitektur Sistem

This document describes the concrete system architecture and core mechanics of **SANOC (Smart Agent & Network Operations Center)**.  
*Dokumen ini menjelaskan arsitektur sistem dan mekanika utama dari SANOC.*

---

## 1. High-Level Architecture Overview / Gambaran Umum Arsitektur

SANOC is engineered as a modern, decoupled **three-service architecture** designed for high throughput network monitoring, fault-tolerant alerting, and real-time operator visibility.

*SANOC dirancang dengan arsitektur 3-layanan modern yang terpisah untuk pemantauan jaringan throughput tinggi, sistem peringatan andal, dan visibilitas operator real-time.*

```mermaid
flowchart TB
    subgraph ClientLayer ["Client Layer"]
        UI["Vue 3 SPA (Vite, Pinia, TS, Tailwind)<br/>Port 5173"]
    end

    subgraph CoreLayer ["Core Service Layer"]
        GoBackend["Go Backend API & Poller Engine<br/>(Goroutines & Net-Ping)<br/>Port 8080"]
    end

    subgraph DataLayer ["Data & State Layer"]
        PG[(PostgreSQL 14+<br/>Persistent Storage)]
        Redis[(Redis 6+<br/>Cache, Pub/Sub, Buffer)]
    end

    subgraph IntegrationLayer ["Integration & Sidecar Services"]
        WASidecar["Node.js wa-sidecar<br/>(Baileys WhatsApp Web API)<br/>Port 3001"]
        TelegramAPI["Telegram Bot API<br/>(api.telegram.org)"]
    end

    subgraph TargetNetwork ["Monitored Infrastructure"]
        Devices["Network Devices<br/>(Routers, Switches, APs, CCTV, NVR, SmartPower)"]
    end

    %% Communication Links
    UI -- "REST API (HTTP/JWT)" --> GoBackend
    UI -- "WebSockets (/ws)" --> GoBackend
    
    GoBackend -- "SQL Queries / pgx" --> PG
    GoBackend -- "Status Cache / Aggregation Buffer / Rate Limits" --> Redis
    
    GoBackend -- "1. Primary Alert (Internal HTTP)" --> WASidecar
    WASidecar -. "If Failed / Timeout" .-> GoBackend
    GoBackend -- "2. Fallback Alert (HTTPS API)" --> TelegramAPI
    
    GoBackend -- "ICMP Ping & SNMP v2c Probes" --> Devices
    GoBackend -- "ARP Table Lookup (MAC-to-IP)" --> Devices
```

---

## 2. Core Service Components / Komponen Utama Service

### 2.1 Go Backend (`/backend`)
- **API Engine**: Built in Go (Golang 1.22+) using standard Gin Web Framework. Serves REST endpoints for authentication, device management, incident tracking, reports, settings, and user administration.
- **Security & RBAC**: Protected routes use JWT middleware enforcing RBAC (`admin`, `pimpinan`, `anggota`) and TOTP MFA (`pquerna/otp`).
- **WebSocket Gateway (`/ws`)**: Pushes real-time polling logs (Live Feed) and state transition events (Notification Bell alerts) directly to connected browser clients.
- **Poller Engine**: Background daemon running inside the Go binary executing periodic health probes.

### 2.2 Vue 3 Frontend (`/frontend`)
- **Single Page Application (SPA)**: Built with Vue 3 (Composition API), Vite, TypeScript, and Pinia state management.
- **Design System**: Uses dark-mode styling for high-density NOC dashboards.
- **Realtime Integration**: Subscribes to backend WebSockets (`/ws`) to dynamically update status badges, charts, and notification feeds without manual page reloads.

### 2.3 Node.js WhatsApp Sidecar (`/backend/wa-sidecar`)
- **Baileys Web API**: Independent Node.js service wrapping `@whiskeysockets/baileys` to connect to WhatsApp Web via QR Code pairing.
- **Token Security**: Communicates with the Go backend over internal HTTP (`POST /send-message`) authenticated via a shared Bearer token.
- **Session Persistence**: Stores session credentials on disk so WhatsApp connection survives service restarts without requiring re-pairing.

### 2.4 Data & State Infrastructure
- **PostgreSQL**: Stores relational domain entities (Users, User Logs, Devices, Device Status Logs, Metrics, Incidents, Notification Logs, WhatsApp Targets, Integration Settings).
- **Redis**: Acts as an in-memory cache for live device status (`device:status:<id>`), rate-limiting token bucket (`ratelimit:wa`), and notification batching buffer (`notif:buffer`).
