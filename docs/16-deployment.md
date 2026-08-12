# 16 — Production Deployment Guide / Panduan Deployment Produksi

Guide for deploying **SANOC (Smart Agent & Network Operations Center)** in production environments.  
*Panduan pengerahan SANOC ke lingkungan produksi (VPS / Bare-Metal / Cloud).*

---

## 1. Production Architecture Overview / Gambaran Arsitektur Produksi

```text
                               ┌────────────────────────────────┐
                               │     Nginx Web Server / TLS     │
                               │   (Reverse Proxy & Static SPA) │
                               └───────────────┬────────────────┘
                                               │
                       ┌───────────────────────┴───────────────────────┐
                       │                                               │
                       ▼ (/api/v1 & /ws)                               ▼ (Static Bundle)
         ┌───────────────────────────┐                   ┌───────────────────────────┐
         │       Go Backend API      │                   │     Vue 3 Frontend SPA    │
         │   (systemd: sanoc-api)    │                   │      (Nginx /dist dir)    │
         └─────────────┬─────────────┘                   └───────────────────────────┘
                       │
         ┌─────────────┴─────────────┐
         ▼                           ▼
┌───────────────────┐       ┌───────────────────┐
│PostgreSQL Database│       │    Redis Cache    │
│(port 5432)        │       │ (port 6379)       │
└───────────────────┘       └───────────────────┘
         │
         ▼
┌───────────────────────────────────┐
│      Node.js wa-sidecar Service   │
│     (systemd: sanoc-wa-sidecar)   │
└───────────────────────────────────┘
```

---

## 2. Environment Configuration (No Hardcoded Credentials)

Create `.env` in the backend directory from `.env.example`:

```bash
cd backend
cp .env.example .env
```

Ensure environment variables use production values:
```env
SANOC_PORT=8080
SANOC_DB_HOST=127.0.0.1
SANOC_DB_PORT=5432
SANOC_DB_USER=sanoc_admin
SANOC_DB_PASSWORD=your_secure_db_password_here
SANOC_DB_NAME=sanoc
SANOC_JWT_SECRET=your_jwt_signing_secret_here
SANOC_CORS_ALLOWED_ORIGINS=https://sanoc.jabarprov.go.id
```

---

## 3. Nginx Reverse Proxy & SSL Configuration

```nginx
server {
    listen 80;
    server_name sanoc.jabarprov.go.id;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name sanoc.jabarprov.go.id;

    ssl_certificate /etc/letsencrypt/live/sanoc.jabarprov.go.id/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sanoc.jabarprov.go.id/privkey.pem;

    root /var/www/sanoc/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }

    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
        add_header X-Content-Type-Options "nosniff";
    }
}
```

---

## 4. Security Check & Deployment Checklist
- [x] Environment files (`.env`) placed out of public web root and listed in `.gitignore`.
- [x] `X-Content-Type-Options: nosniff` header enabled for `/uploads/` route.
- [x] Database credentials and secrets loaded via environment variables.
- [x] HTTPS SSL enabled with Let's Encrypt / Certbot.
