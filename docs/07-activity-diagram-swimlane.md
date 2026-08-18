# 07 — Activity Diagram (Swimlane) / Diagram Aktivitas Swimlane

This document presents a Swimlane Activity Diagram illustrating the distributed actions and interactions between the **Poller Engine**, the **Notification Pipeline**, and **SANOC Operators (`anggota`)**.  
*Dokumen ini menyajikan Diagram Aktivitas Swimlane untuk SANOC.*

---

## 1. Mermaid Swimlane Activity Diagram

```mermaid
flowchart TB
    subgraph PollerEngineLane ["⚙️ Poller Engine (Background Daemon)"]
        PE_Start["Start Polling Execution Cycle"] --> PE_Probe["Probe Target Devices (ICMP / SNMP)"]
        PE_Probe --> PE_CheckFail{"Health Check Failed?"}
        
        PE_CheckFail -- "No (UP)" --> PE_Reset["Reset Failure Counter = 0"] --> PE_End["Wait for Next Cycle"]
        PE_CheckFail -- "Yes (DOWN)" --> PE_Inc["Increment Failure Counter"]
        
        PE_Inc --> PE_ThreshCheck{"Counter >= Threshold?"}
        PE_ThreshCheck -- "No" --> PE_End
        PE_ThreshCheck -- "Yes" --> PE_CreateInc["Create ACTIVE Incident in DB"]
        
        PE_CreateInc --> PE_EmitWS["Emit Realtime WS Alert Badge to Client"]
        PE_EmitWS --> PE_PushBuffer["Push Notification Job to Redis Buffer"]
        
        PE_AutoDetect["Detect Device Recovery (UP)"] --> PE_ResolveInc["Auto-Resolve Active Incident in DB"]
        PE_ResolveInc --> PE_EmitWSResolve["Emit WS Incident Resolution Badge"]
    end

    subgraph NotificationPipelineLane ["🔔 Notification Pipeline (Sidecar & Telegram)"]
        NP_PopBuffer["Flush Aggregated Job Batch from Redis Buffer"] --> NP_RateCheck{"Rate Limit Exceeded?"}
        
        NP_RateCheck -- "Yes" --> NP_Throttle["Throttle / Delay Delivery"] --> NP_RateCheck
        NP_RateCheck -- "No" --> NP_SendWA["HTTP POST /send-message to wa-sidecar"]
        
        NP_SendWA --> NP_WACheck{"WhatsApp Dispatch Success?"}
        
        NP_WACheck -- "Yes" --> NP_LogWA["Write Status 'Sent' to notification_logs"]
        NP_WACheck -- "No / Disconnected" --> NP_SendTG["HTTP POST to Telegram Bot API"]
        
        NP_SendTG --> NP_TGCheck{"Telegram Dispatch Success?"}
        NP_TGCheck -- "Yes" --> NP_LogTG["Write Status 'Sent' (Telegram) to notification_logs"]
        NP_TGCheck -- "No" --> NP_LogFail["Write Status 'Failed' to notification_logs"]
        
        NP_LogWA --> NP_Done["Notification Pipeline Complete"]
        NP_LogTG --> NP_Done
        NP_LogFail --> NP_Done
    end

    subgraph SANOCStaffLane ["👤 SANOC Operator (Anggota / Operator)"]
        NOC_RecvAlert["Receive WhatsApp / Telegram Alert / WS Badge"] --> NOC_OpenView["Open Incident Detail View in Web UI"]
        NOC_OpenView --> NOC_Triage["Inspect Impact & Physical Location"]
        NOC_Triage --> NOC_FieldAction["Coordinate Field Repair"]
        NOC_FieldAction --> NOC_ManualResolve["Click Manual Resolve (Optional)"]
    end

    %% Inter-lane Cross Trigger Links
    PE_PushBuffer -- "Trigger Alert Delivery" --> NP_PopBuffer
    NP_Done -- "Deliver Message Notification" --> NOC_RecvAlert
```
