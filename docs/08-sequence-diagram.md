# 08 — Sequence Diagrams / Diagram Sekuensial

This document contains Mermaid Sequence Diagrams covering two core system workflows in **SANOC v2.6.0**: **User Authentication, TOTP MFA & WebSocket Subscription**, and **Incident Detection & Notification Dispatch**.  
*Dokumen ini menyajikan Diagram Sekuensial untuk SANOC.*

---

## 1. Sequence 1: User Login, TOTP MFA & Realtime WebSocket Subscription

```mermaid
sequenceDiagram
    autonumber
    actor User as User (Browser)
    participant FE as Vue 3 Frontend
    participant BE as Go Backend API
    participant DB as PostgreSQL
    participant WS as WebSocket Hub (/ws)

    User->>FE: Input Credentials (email/username, password)
    FE->>BE: POST /api/v1/auth/login {email, password}
    BE->>DB: SELECT * FROM users WHERE email=$1 OR username=$1
    DB-->>BE: Return User Record & Password Hash
    BE->>BE: Verify Password via bcrypt
    
    alt Credentials Valid & MFA Enabled
        BE-->>FE: HTTP 200 OK {mfa_required: true, temp_token: "..."}
        FE-->>User: Display 6-Digit TOTP Passcode Prompt
        User->>FE: Enter 6-Digit TOTP Code
        FE->>BE: POST /api/v1/auth/verify-mfa {code: "123456"}
        BE->>BE: Validate TOTP Code (±90s Clock Skew Tolerance)
    end

    alt Authentication Success
        BE->>DB: INSERT INTO user_logs (action='login', ip, user_agent)
        BE-->>FE: HTTP 200 OK {token: "JWT...", user: {id, username, name, role}}
        FE->>FE: Store JWT & User State in Pinia
        FE->>User: Redirect to Dashboard (/dashboard)
        
        par Initial Data Fetch
            FE->>BE: GET /api/v1/dashboard/summary (Bearer JWT)
            BE-->>FE: HTTP 200 OK (Summary Cards Data)
        and WebSocket Realtime Handshake
            FE->>WS: Connect ws://localhost:8080/ws?token=JWT
            WS-->>FE: 101 Switching Protocols (Connection Established)
        end
    end
```

---

## 2. Sequence 2: Incident Detection, Aggregation & Notification Dispatch

```mermaid
sequenceDiagram
    autonumber
    participant PE as Poller Engine
    participant DB as PostgreSQL
    participant WS as WebSocket Hub (/ws)
    participant RD as Redis Queue
    participant WA as wa-sidecar (Node.js)
    participant TG as Telegram Bot API

    PE->>PE: ICMP/SNMP Probe Fails (Consecutive >= Threshold)
    PE->>DB: INSERT INTO incidents (status='ACTIVE')
    PE->>WS: Broadcast WS Event {type: "INCIDENT_CREATED"}
    PE->>RD: LPUSH notif:buffer (Incident Payload)
    
    loop Batch Flusher Window
        RD->>RD: Pop Batch Messages
    end
    
    RD->>WA: HTTP POST /send-message (Bearer SIDECAR_SECRET)
    
    alt WhatsApp Delivery Success
        WA-->>RD: HTTP 200 OK {status: "delivered"}
        RD->>DB: INSERT INTO notification_logs (channel='WhatsApp', status='Sent')
    else WhatsApp Delivery Fails / Timeout
        WA-->>RD: HTTP 500 / Timeout Connection
        RD->>TG: HTTPS POST api.telegram.org/bot<TOKEN>/sendMessage
        TG-->>RD: HTTP 200 OK {ok: true}
        RD->>DB: INSERT INTO notification_logs (channel='Telegram', status='Sent')
    end
```
