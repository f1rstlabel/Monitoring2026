# 05 — Polling & Alerting Flowchart / Flowchart Polling & Notifikasi

This document details the complete technical decision logic from initial device probing through threshold counter evaluation, status transition detection, notification aggregation, rate limiting, and alert delivery fallback in **SANOC v2.6.0**.  
*Dokumen ini menjelaskan alur keputusan teknis dari eksekusi probe ICMP/SNMP hingga pengiriman notifikasi.*

---

## 1. Mermaid Polling-to-Notification Flowchart

```mermaid
flowchart TD
    A["⏱️ Start Poller Cycle (Every N seconds)"] --> B["Fetch Active Devices & Settings from DB"]
    B --> C["Spawn Worker Pool (Concurrently)"]
    
    C --> D{"Addressing Mode?"}
    D -- "DHCP" --> E["Resolve MAC via ARP Table"]
    E --> F{"IP Changed?"}
    F -- "Yes" --> G["Update last_known_ip"] --> H["Send ICMP Echo Ping (3 Packets)"]
    F -- "No" --> H
    D -- "Static" --> H
    
    H --> I{"SNMP Enabled?"}
    I -- "Yes" --> J["Fetch SNMP CPU/Memory OIDs"] --> K["Record to device_metrics"] --> L{"Ping / Probe Success?"}
    I -- "No" --> L
    
    L -- "Success (UP)" --> M["Reset Consecutive Fail Counter = 0"]
    L -- "Failure (DOWN)" --> N["Increment Consecutive Fail Counter (+1)"]
    
    M --> O{"Previous Status == DOWN?"}
    N --> P{"Fail Counter >= Failure Threshold?"}
    
    P -- "No (Threshold Not Reached)" --> Q["Status Remains UP (No Transition)"] --> R["Publish Live Feed Log to WS"]
    P -- "Yes (Threshold Exceeded)" --> S{"Debounce Window Passed?"}
    
    S -- "No" --> Q
    S -- "Yes" --> T{"Previous Status == UP?"}
    
    T -- "No (Already DOWN)" --> U["Status Remains DOWN (No Transition)"] --> R
    T -- "Yes (Transition to DOWN)" --> V["Change Status to DOWN"]
    
    O -- "No (Already UP)" --> Q
    O -- "Yes (Transition to UP)" --> W["Change Status to UP"]
    
    V --> X["Create Active Incident in DB & Push WS Alert Badge"]
    W --> Y["Resolve Active Incident in DB & Push WS Alert Badge"]
    
    X --> Z["Push Incident Job to Redis Aggregation Queue (notif:buffer)"]
    Y --> Z
    
    Z --> AA["Aggregation Buffer Flushes Batch Window"]
    AA --> AB{"Rate Limit Check"}
    
    AB -- "Allowed" --> AC["Send HTTP POST to wa-sidecar"]
    AB -- "Exceeded" --> AD["Queue in Redis Buffer for Next Minute Window"]
    
    AC --> AE{"wa-sidecar Success?"}
    AE -- "Yes" --> AF["Log Delivery Success to notification_logs"]
    AE -- "No / Timeout" --> AG["Fallback: Send HTTPS Request to Telegram Bot API"]
    
    AG --> AH["Log Delivery (Primary Failed, Fallback Success/Fail)"]
```
