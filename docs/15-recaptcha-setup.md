# 15 — Security, Anti-Bot & TOTP MFA Configuration
*Konfigurasi Keamanan, Anti-Bot & TOTP MFA*

This guide outlines security practices, Two-Factor Authentication (TOTP MFA), and optional bot protection for **SANOC**.  
*Panduan ini memberikan instruksi konfigurasi keamanan, Autentikasi 2FA (TOTP MFA), dan proteksi bot untuk SANOC.*

---

## 1. Two-Factor Authentication (TOTP MFA) Configuration

SANOC implements RFC 6238 Time-Based One-Time Password (TOTP) algorithm compatible with Google Authenticator, Authy, and 1Password.

### 1.1 Toleransi Clock Skew & Window Standard
To handle system clock drift between mobile devices and the server, SANOC evaluates a $\pm 90$-second window (7 consecutive 30-second time steps: $-90s, -60s, -30s, 0s, +30s, +60s, +90s$).

### 1.2 User Activation Workflow
1. Navigate to **User Profile** (`/profile`).
2. Click **Setup 2FA**.
3. Scan the generated QR code or enter the secret string manually in your Authenticator app.
4. Enter the current 6-digit code to confirm activation.

---

## 2. Google reCAPTCHA v3 Anti-Bot Setup (Optional)

1. Open [Google reCAPTCHA Admin Console](https://www.google.com/recaptcha/admin/).
2. Register a new site:
   - **Type**: reCAPTCHA v3 (Score-based).
   - **Domains**: Add your production domain (e.g. `sanoc.jabarprov.go.id`).
3. Copy the **Site Key** and **Secret Key**.

### Environment Configuration
**Backend (`.env`)**:
```env
RECAPTCHA_ENABLED=true
RECAPTCHA_SECRET_KEY=your_production_secret_key_here
```

**Frontend (`.env`)**:
```env
VITE_RECAPTCHA_SITE_KEY=your_production_site_key_here
```

---

## 3. Data Protection & Sensitive Variable Isolation
- **No Hardcoded Secrets**: All JWT secrets, DB passwords, and API tokens MUST be loaded from environment variables (`.env`).
- **Git Safety**: `.env`, `*.key`, `*.pem`, `*.sql.bak`, and `uploads/` are strictly isolated in `.gitignore`.
