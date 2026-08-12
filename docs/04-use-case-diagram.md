# 04 — Use Case Diagram / Diagram Use Case

This document presents the Use Case Diagram for **SANOC v2.6.0**, illustrating system interaction boundaries and RBAC permissions across all user roles and internal daemons.  
*Dokumen ini menyajikan Diagram Use Case untuk SANOC.*

---

## 1. Mermaid Use Case Diagram

```mermaid
flowchart LR
    %% Actors Definition
    subgraph Actors ["System Actors"]
        Admin["👤 Admin<br/>(Full Access)"]
        Anggota["👤 Anggota<br/>(NOC Operator)"]
        Pimpinan["👤 Pimpinan<br/>(Executive)"]
        PollerEngine["⚙️ Poller Engine<br/>(Automated Daemon)"]
    end

    %% Use Cases Boundary
    subgraph SANOCSystem ["SANOC System Boundary"]
        
        %% Authentication & General
        UC_Login(["UC-01: Authenticate / Login / TOTP 2FA"])
        UC_Profile(["UC-02: Manage Profile & Upload Photo"])
        UC_Dashboard(["UC-03: Monitor Dashboard & Live Feed"])
        
        %% Inventory Management
        UC_ViewInventory(["UC-04: View Device Inventory & Metrics"])
        UC_ManageDevice(["UC-05: Add / Edit / Delete Device"])
        UC_AutoDetect(["UC-06: Subnet Auto-Detect & Import"])
        
        %% Incident Lifecycle
        UC_ViewIncidents(["UC-07: View Incidents & Outage Timeline"])
        UC_ManualResolve(["UC-08: Manually Resolve Incident"])
        UC_AutoResolve(["UC-09: Auto-Detect & Resolve Outage"])
        
        %% Reports & Analytics
        UC_ViewReports(["UC-10: View Reports & Export PDF/Excel/CSV"])
        
        %% System & Integrations Config
        UC_ConfigSettings(["UC-11: Configure Polling & Threshold Settings"])
        UC_ConfigIntegrations(["UC-12: Pair WhatsApp QR & Setup Telegram Bot"])
        
        %% User & Audit Management
        UC_ManageUsers(["UC-13: Manage User Accounts & Username"])
        UC_ViewLogs(["UC-14: View User Activity Audit Logs"])
    end

    %% User Connections
    Pimpinan --> UC_Login
    Pimpinan --> UC_Profile
    Pimpinan --> UC_Dashboard
    Pimpinan --> UC_ViewInventory
    Pimpinan --> UC_ViewIncidents
    Pimpinan --> UC_ViewReports

    Anggota --> UC_Login
    Anggota --> UC_Profile
    Anggota --> UC_Dashboard
    Anggota --> UC_ViewInventory
    Anggota --> UC_ManageDevice
    Anggota --> UC_AutoDetect
    Anggota --> UC_ViewIncidents
    Anggota --> UC_ManualResolve
    Anggota --> UC_ViewReports

    Admin --> UC_Login
    Admin --> UC_Profile
    Admin --> UC_Dashboard
    Admin --> UC_ViewInventory
    Admin --> UC_ManageDevice
    Admin --> UC_AutoDetect
    Admin --> UC_ViewIncidents
    Admin --> UC_ManualResolve
    Admin --> UC_ViewReports
    Admin --> UC_ConfigSettings
    Admin --> UC_ConfigIntegrations
    Admin --> UC_ManageUsers
    Admin --> UC_ViewLogs

    PollerEngine --> UC_AutoResolve
```
