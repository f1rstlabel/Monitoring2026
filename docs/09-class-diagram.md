# 09 — Class Diagram / Diagram Kelas

This document presents the Mermaid Class Diagram for **SANOC v2.6.0**, detailing Go domain structs, core enums, setting structures, and repository interfaces.  
*Dokumen ini menyajikan Diagram Kelas untuk SANOC.*

---

## 1. Mermaid Class Diagram

```mermaid
classDiagram
    %% Core Enums
    class DeviceType {
        <<enumeration>>
        AccessPoint
        Switch
        Router
        SmartPower
        CCTV
        NVR
    }

    class DeviceStatus {
        <<enumeration>>
        StatusUP ("UP")
        StatusDOWN ("DOWN")
    }

    class Role {
        <<enumeration>>
        RoleAdmin ("admin")
        RolePimpinan ("pimpinan")
        RoleAnggota ("anggota")
    }

    class AddressingMode {
        <<enumeration>>
        AddressingStatic ("Static")
        AddressingDHCP ("DHCP")
    }

    %% Domain Entities
    class Device {
        +string ID
        +string Name
        +DeviceType Type
        +string IP
        +string MAC
        +DeviceStatus Status
        +AddressingMode AddressingMode
        +string Location
        +string Rack
        +bool SNMPEnabled
        +string Community
        +int SNMPPort
        +int CustomThreshold
    }

    class User {
        +string ID
        +string Username
        +string Name
        +string Email
        +Role Role
        +string Status
        +string AvatarURL
        +string LastActive
        +bool MFAEnabled
        +string MFASecret
    }

    class Incident {
        +string ID
        +string DeviceID
        +string DeviceName
        +string DeviceType
        +string DeviceIP
        +string Status
        +string StartTime
        +string ResolvedAt
        +string Duration
    }

    class UserRepository {
        <<interface>>
        +GetAll() ([]User, error)
        +GetByID(id) (*User, error)
        +GetByEmail(email) (*User, error)
        +Create(user, passwordHash) error
        +UpdateUser(user) error
        +SetMFASetup(id, enabled, secret) error
    }

    User --> Role
    Device --> DeviceType
    Device --> DeviceStatus
    Device --> AddressingMode
```
