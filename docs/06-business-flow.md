# 06 — Business Process Flow / Alur Proses Bisnis

This document describes the end-to-end business workflow of **SANOC v2.6.0** from an operational NOC perspective.  
*Dokumen ini menjelaskan alur proses bisnis operasional SANOC.*

---

## 1. Non-Technical Business Process Diagram

```mermaid
flowchart TD
    subgraph Step1 ["1. Onboarding & Baseline"]
        A["Network Administrator registers IT Assets<br/>(Routers, Switches, APs, CCTV, NVR)"] --> B["System establishes baseline monitoring parameters<br/>(MAC, IP, Thresholds, SNMP metrics)"]
    end

    subgraph Step2 ["2. Continuous Automated Monitoring"]
        B --> C["SANOC runs 24/7 background polling"]
        C --> D["Dashboard reflects live SLA Uptime & Latency"]
    end

    subgraph Step3 ["3. Outage Detection & Alerting"]
        C --> E{"Network Outage / Failure Detected?"}
        E -- "No" --> C
        E -- "Yes (Threshold Confirmed)" --> F["System automatically creates ACTIVE Incident"]
        F --> G["Broadcasting alert to Operators via WhatsApp & Telegram"]
    end

    subgraph Step4 ["4. Incident Triage & Physical Remediation"]
        G --> H["NOC Operator (Anggota) receives notification"]
        H --> I["Operator opens Incident Detail to inspect root cause"]
        I --> J["Field Technician dispatched for physical repair"]
    end

    subgraph Step5 ["5. Recovery & Incident Resolution"]
        J --> K{"Network Connectivity Restored?"}
        K -- "Manual Action" --> L["Operator clicks Manual Resolve"]
        K -- "Poller Detects UP" --> M["System automatically marks Incident as RESOLVED"]
        L --> N["Resolution timestamp logged"]
        M --> N
    end

    subgraph Step6 ["6. Analysis & Executive Review"]
        N --> O["System identifies recurring unstable assets (Flap Report)"]
        O --> P["Diskominfo Leadership (Pimpinan) exports Tab-Specific SLA Reports (PDF/Excel/CSV)"]
        P --> Q["Data-driven infrastructure planning & hardware maintenance"]
    end
```
