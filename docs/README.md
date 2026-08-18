# SANOC Infrastructure — System Documentation
*Dokumentasi Sistem & Infrastruktur SANOC (Diskominfo Jawa Barat)*

Welcome to the official technical and operational documentation set for **SANOC (Sanditel Network Operations Center)**, an enterprise-grade IT infrastructure monitoring and incident alert system built for government networks.

Selamat datang di dokumentasi resmi teknis dan operasional untuk **SANOC (Sanditel Network Operations Center)**, platform monitoring infrastruktur jaringan dan audit kepatuhan SLA berbasis enterprise.

---

## 📚 Documentation Index / Indeks Dokumentasi

| # | Document Title / Judul Dokumen | Description / Deskripsi |
|---|---|---|
| **01** | [System Architecture](01-architecture.md) | Overview of the 3-service architecture (Go Backend, Vue 3 SPA, Node.js WhatsApp Sidecar, PostgreSQL, Redis) and Poller Engine mechanics. / *Gambaran umum arsitektur 3-layanan.* |
| **02** | [Usage & Operations Guide](02-usage-guide.md) | Day-to-day operational guide covering RBAC roles, inventory, dashboard, live feed, incident handling, settings, and audit logs. / *Panduan operasional harian.* |
| **03** | [Non-Functional Requirements (NFR & NFSR)](03-nfr-nfsr.md) | Performance, reliability, usability, maintainability, scalability, and security requirements. / *Kebutuhan non-fungsional dan keamanan.* |
| **04** | [Use Case Diagram](04-use-case-diagram.md) | Mermaid Use Case diagram mapping `admin`, `pimpinan`, `anggota`, and `poller engine`. / *Diagram Use Case peran pengguna.* |
| **05** | [Polling & Alerting Flowchart](05-flowchart.md) | Mermaid flowchart detailing ICMP/SNMP execution, threshold processing, and alert dispatching. / *Alur kerja polling dan notifikasi.* |
| **06** | [Business Process Flow](06-business-flow.md) | End-to-end non-technical business flow from device registration through NOC incident response and executive SLA reporting. / *Alur proses bisnis operasional NOC.* |
| **07** | [Activity Diagram (Swimlane)](07-activity-diagram-swimlane.md) | Mermaid swimlane diagram illustrating tasks divided across **Poller Engine**, **Notification Pipeline**, and **NOC Staff**. / *Diagram aktivitas swimlane.* |
| **08** | [Sequence Diagrams](08-sequence-diagram.md) | Sequence diagrams for User Login, WebSockets, and Incident Alert Dispatching. / *Diagram sekuensial login & notifikasi.* |
| **09** | [Class Diagram](09-class-diagram.md) | Mermaid class diagram representing real Go domain structures and repository interfaces. / *Diagram kelas domain backend.* |
| **10** | [Entity Relationship Diagram (ERD)](10-erd.md) | Mermaid ERD matching the PostgreSQL database schema (`users`, `devices`, `incidents`, `whatsapp_targets`, etc.). / *Skema relasi database PostgreSQL.* |
| **11** | [Data & Memory Management](11-data-and-memory-management.md) | Breakdown of PostgreSQL persistence vs. Redis caching vs. Poller in-memory counters. / *Manajemen memori dan retensi data.* |
| **12** | [Local Installation Guide](12-installation.md) | Step-by-step setup guide for local development (Prerequisites, `.env` configs, DB migrations). / *Panduan instalasi lokal.* |
| **13** | [Docker Deployment Guide](13-docker-deployment.md) | Production deployment guide featuring Dockerfiles, `docker-compose.yml`, Nginx proxy, and SSL setup. / *Panduan pengerahan menggunakan Docker.* |
| **14** | [Git Branching Workflow](14-git-branching-workflow.md) | Practical Git guide for remote tracking, branch creation, and naming conventions. / *Alur kerja dan konvensi Git.* |
| **15** | [Security & MFA Setup](15-recaptcha-setup.md) | Guide for configuring TOTP 2FA Google Authenticator and bot security. / *Panduan keamanan dan konfigurasi TOTP MFA.* |
| **16** | [Production Deployment Guide](16-deployment.md) | Comprehensive step-by-step guide for bare-metal / VPS deployment with systemd services, Nginx, and SSL. / *Panduan pengerahan ke server produksi.* |
| **17** | [SMTP Email Configuration](17-smtp-email-configuration.md) | Guide for setting up production SMTP email gateway (Gmail, Gov Mail, SendGrid, Mailpit) and OTP verification. / *Panduan konfigurasi email SMTP & OTP.* |

---

## 🏛️ System Overview At A Glance / Gambaran Arsitektur

```
                  ┌──────────────────────────────────────────────────┐
                  │                 Vue 3 Frontend                   │
                  │     (Vite + Pinia + TypeScript + Tailwind)       │
                  └──────────────┬───────────────────┬───────────────┘
                                 │ REST API          │ WebSockets (/ws)
                                 ▼                   ▼
                  ┌──────────────────────────────────────────────────┐
                  │                    Go Backend                    │
                  │          (API Handlers + Poller Engine)          │
                  └──────────────┬─────────────────┬─────────────────┘
                                 │                 │
                      ┌──────────┴───────┐     ┌───┴──────────────┐
                      │PostgreSQL Database│     │   Redis Cache    │
                      │(Persisted State) │     │& Pub/Sub Queue   │
                      └──────────────────┘     └──────────────────┘
                                 │
                                 ▼ [Primary Attempt]
                  ┌───────────────────────────────┐
                  │      Node.js wa-sidecar       │
                  │  (Baileys WhatsApp Web API)   │
                  └──────────────┬────────────────┘
                                 │ [If WhatsApp Fails / Fallback]
                                 ▼
                  ┌───────────────────────────────┐
                  │       Telegram Bot API        │
                  │      (Fallback Channel)       │
                  └───────────────────────────────┘
```

---

## 🔒 Security & Privacy / Keamanan & Privasi
All sensitive keys, database passwords, and API secrets are strictly managed via environment variables (`.env`). No production secrets or personal access tokens are committed to repository files.
