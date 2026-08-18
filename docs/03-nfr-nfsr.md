# 03 — Non-Functional Requirements (NFR & NFSR) / Kebutuhan Non-Fungsional & Keamanan

This document specifies the concrete Non-Functional Requirements (NFR) and Non-Functional Security Requirements (NFSR) enforced by the **SANOC v2.6.0** codebase.  
*Dokumen ini menetapkan Kebutuhan Non-Fungsional (NFR) dan Keamanan (NFSR) pada aplikasi SANOC.*

---

## 1. Non-Functional Requirements (NFR) / Kebutuhan Non-Fungsional

| Category | Requirement ID | Specification & Live Implementation Detail |
|---|---|---|
| **Performance** | `NFR-PERF-01` | **Polling Cycle Efficiency**: Completes a full ICMP/SNMP probe cycle for 1,000+ devices within < 15 seconds using a goroutine worker pool with configurable batch concurrency. |
| **Performance** | `NFR-PERF-02` | **Probe Latency & Timeout**: ICMP ping timeout set to 2.0s per packet; SNMP probes time out in 1.5s to prevent worker pool starvation. |
| **Performance** | `NFR-PERF-03` | **Realtime Push Overhead**: WebSocket message broadcast latency from poller state transition to client UI rendering is < 50 milliseconds. |
| **Reliability** | `NFR-RELI-01` | **False Positive Prevention**: Enforces consecutive failure thresholds and a debounce window before transitioning status to `DOWN`. |
| **Reliability** | `NFR-RELI-02` | **Notification Channel Redundancy**: Implements automatic fallback from WhatsApp to Telegram if `wa-sidecar` fails or times out. |
| **Usability** | `NFR-USAB-01` | **Role-Tailored UI**: Displays actionable views based on role (`admin`, `pimpinan`, `anggota`), hiding administrative action buttons from read-only users. |
| **Usability** | `NFR-USAB-02` | **High-Density NOC Visibility**: Dark-mode dashboard optimized for 24/7 SANOC wall monitors, providing clear color-coded status badges (`UP` = Green, `DOWN` = Red). |
| **Reporting** | `NFR-REPO-01` | **Tab-Specific Exports**: Excel (`.xls`), CSV (`.csv`), and PDF (`.pdf`) exports strictly contain dataset matching active report tab (**Downtime by Device**, **Recurring Issues**, **Active Incidents**). |
| **Reporting** | `NFR-REPO-02` | **Multi-Page A4 PDF Layout**: A4 PDF print layouts eliminate empty Page 1 gaps (`break-before: avoid`), repeat table column headers across pages, and format print titles dynamically. |

---

## 2. Non-Functional Security Requirements (NFSR) / Kebutuhan Keamanan

| Category | Requirement ID | Specification & Live Implementation Detail |
|---|---|---|
| **Authentication** | `NFSR-AUTH-01` | **Bcrypt Password Hashing**: User credentials stored in PostgreSQL `users` table using `bcrypt` password hashing with cost factor 10. |
| **Authentication** | `NFSR-AUTH-02` | **JWT Token Security**: API access secured by signed JSON Web Tokens (JWT) with standard expiration and algorithm validation (`HS256`). |
| **Authentication** | `NFSR-AUTH-03` | **TOTP 2FA (MFA)**: RFC 6238 Time-Based One-Time Password support with Google Authenticator and $\pm 90s$ clock-skew tolerance window. |
| **Authentication** | `NFSR-AUTH-04` | **Real Email OTP Verification**: New user registration requires active real email verification via a 6-digit OTP code (15-min expiry) delivered by Gmail SMTP with dummy domain blacklisting. |
| **Authentication** | `NFSR-AUTH-05` | **Dual-Method Password Reset**: Profile password management supports verification via current password (min 12 chars) or verified email OTP. |
| **Authorization** | `NFSR-AUTH-06` | **Strict Backend RBAC Middleware**: Endpoints protected by `RequireRole("admin")` or `RequireRole("admin", "anggota")` at router level. |
| **Upload Security** | `NFSR-UPLD-01` | **MIME Magic-Byte Checking**: File uploads validate magic-bytes (JPG, PNG, WEBP), enforce 10MB limits, sanitize filenames (`filepath.Base`), and send `X-Content-Type-Options: nosniff` headers. |
| **Secrets Handling** | `NFSR-SECR-01` | **Environment Isolation**: All secrets, JWT tokens, DB passwords, and API keys stored exclusively in `.env` files and isolated via `.gitignore`. |
| **Audit Logging** | `NFSR-AUDT-01` | **User Action Audit Trail**: All authentication events (`login`, `logout`) and administrative mutations (`add_device`, `edit_user`) log to `user_logs`. |
