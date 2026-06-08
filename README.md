# InfraBIOS

**BIOS Configuration Management System — Endpoints & Servers**

A centralized platform for discovering, auditing, configuring, and enforcing BIOS/UEFI settings across enterprise PCs, laptops, workstations, and servers.

The platform provides a single source of truth for firmware configuration, security settings, compliance policies, and hardware lifecycle management across your entire fleet — from a developer's laptop to a datacenter rack.

> Think of it as **Git + Ansible + Fleet Management**, but specifically for BIOS and firmware configuration.

---

## Why Endpoints + Servers

A server-only BIOS manager is useful. A platform that covers **every piece of hardware in the organization** is transformative:

| Device Class | Examples | Risk Without Management |
|---|---|---|
| Windows PCs | Dell OptiPlex, HP EliteDesk | Secure Boot disabled, TPM bypassed |
| Linux PCs | Engineering workstations | Virtualization off, inconsistent boot mode |
| Laptops | ThinkPad, Latitude, EliteBook | BitLocker breaks on BIOS change |
| Servers | PowerEdge, ProLiant, ThinkSystem | Performance drift, compliance failure |

One policy engine. One audit trail. One API. All hardware.

---

## The Problem

Configuration drift is silent and compounding:

| Device | Secure Boot | TPM | Virtualization |
|---|---|---|---|
| pc-001 | ✅ | ✅ | ✅ |
| pc-002 | ❌ | ✅ | ✅ |
| lab-001 | ✅ | ❌ | ✅ |
| lab-002 | ❌ | ❌ | ❌ |

These inconsistencies cause performance degradation, security gaps, deployment failures, and compliance violations.

InfraBIOS provides a **single source of truth** — define a policy once, enforce it everywhere.

---

## Supported Devices

### Endpoints
- Windows PCs (Dell OptiPlex, HP EliteDesk, Lenovo ThinkCentre)
- Linux PCs and engineering workstations
- Laptops (ThinkPad, Latitude, EliteBook, MacBook via EFI)

### Infrastructure
- Physical servers (Dell PowerEdge, HPE ProLiant, Lenovo ThinkSystem, Supermicro)
- Lab servers and bare-metal CI nodes
- Datacenter servers at any scale

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│              Operators / IT / Security / CI                  │
│         REST API  ·  curl  ·  Terraform  ·  Ansible          │
└───────────────────────────┬──────────────────────────────────┘
                            │ Bearer token (HTTPS)
┌───────────────────────────▼──────────────────────────────────┐
│                    InfraBIOS Server                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐  │
│  │ /devices │ │/policies │ │  /drift  │ │  /compliance   │  │
│  │/firmware │ │/changes  │ │/snapshots│ │    /jobs       │  │
│  └────┬─────┘ └──────────┘ └──────────┘ └────────────────┘  │
│       │                                                      │
│  ┌────▼──────────────────────────────────────────────────┐   │
│  │  Compliance Engine  ·  Drift Detector  ·  Job Queue   │   │
│  └────────────────────────────┬──────────────────────────┘   │
│                               │                              │
│  ┌────────────────────────────▼──────────────────────────┐   │
│  │              PostgreSQL (JSONB settings)               │   │
│  └───────────────────────────────────────────────────────┘   │
└───────────────────────────────┬──────────────────────────────┘
                                │ Agent token (HTTPS)
     ┌──────────────────────────┼───────────────────────────┐
     │                          │                           │
┌────▼───────────┐   ┌──────────▼──────┐   ┌───────────────▼──┐
│  infrabios-    │   │  infrabios-     │   │  infrabios-      │
│  agent         │   │  agent          │   │  agent           │
│  Dell OptiPlex │   │  HPE ProLiant   │   │  Lenovo ThinkPad │
│  (Windows PC)  │   │  (Server)       │   │  (Laptop)        │
└────────────────┘   └─────────────────┘   └──────────────────┘
```

---

## Core Features

| Feature | Description |
|---|---|
| **Hardware Discovery** | Auto-collect hostname, manufacturer, model, BIOS version via `dmidecode` + vendor CLIs |
| **BIOS Inventory** | Store every setting as JSONB — full history per device |
| **Policy Management** | Create security baselines (Corporate Security, CIS, NIST) and assign per device class |
| **Compliance Engine** | Compare expected vs actual — scored reports with per-key violations |
| **Drift Detection** | Per-collection diff — emits events the moment any setting changes |
| **Configuration Deployment** | Push BIOS settings to thousands of devices via job queue |
| **Firmware Lifecycle** | Track BIOS, TPM, BMC, NIC firmware versions and approval status |
| **Security Monitoring** | Continuously audit Secure Boot, TPM, BIOS passwords, virtualization flags |
| **Change Management** | Full approval workflow — request → review → apply → audit trail |
| **Snapshots** | Point-in-time backup of full settings + firmware before any change |
| **Audit Logging** | Every change records who, what, when, and why |
| **Bulk Fleet Operations** | Apply policies to 10, 100, or 10,000 devices in one job |

---

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 14+

### Setup

```bash
git clone https://github.com/Alexmaster12345/InfraBIOS.git
cd InfraBIOS

cp .env.example .env
# Edit .env — set DATABASE_URL and tokens

go mod tidy
createdb infrabios
make migrate
make run
```

### Run Agent (on each managed server)

```bash
INFRABIOS_SERVER_URL=http://<server>:8080 \
INFRABIOS_AGENT_TOKEN=agent-changeme \
INFRABIOS_SERVER_ID=<uuid> \
./bin/infrabios-agent
```

---

## API Reference

All endpoints (except `/health`) require: `Authorization: Bearer <token>`

### Servers
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/servers` | Register a server |
| `GET` | `/api/v1/servers` | List all servers |
| `GET` | `/api/v1/servers/{id}` | Get server detail |
| `PUT` | `/api/v1/servers/{id}/profile` | Assign a BIOS profile |
| `GET` | `/api/v1/servers/{id}/settings` | Latest collected BIOS settings |
| `DELETE` | `/api/v1/servers/{id}` | Remove server |

### Profiles
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/profiles` | Create profile |
| `GET` | `/api/v1/profiles` | List profiles |
| `PUT` | `/api/v1/profiles/{id}` | Update profile |
| `DELETE` | `/api/v1/profiles/{id}` | Delete profile |

### Compliance
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/servers/{id}/compliance/scan` | Run compliance check |
| `GET` | `/api/v1/servers/{id}/compliance` | Latest report |
| `GET` | `/api/v1/servers/{id}/compliance/history` | Full history |

### Drift Events
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/drift?server_id=&status=` | List drift events |
| `POST` | `/api/v1/drift/{id}/resolve` | Mark resolved |
| `POST` | `/api/v1/drift/{id}/ignore` | Suppress |

### Firmware
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/servers/{id}/firmware` | List firmware inventory |
| `POST` | `/api/v1/servers/{id}/firmware` | Add/update component |
| `POST` | `/api/v1/servers/{id}/firmware/{component}/approve` | Approve version |

### Changes (Approval Workflow)
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/changes` | Request a change |
| `GET` | `/api/v1/changes?status=pending` | List pending approvals |
| `POST` | `/api/v1/changes/{id}/review` | Approve or reject |

### Snapshots
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/servers/{id}/snapshots` | Take snapshot |
| `GET` | `/api/v1/servers/{id}/snapshots` | List snapshots |
| `DELETE` | `/api/v1/snapshots/{id}` | Delete snapshot |

### Jobs
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/jobs` | Create bulk job |
| `GET` | `/api/v1/jobs?status=running` | List jobs |
| `GET` | `/api/v1/jobs/{id}` | Get job status |
| `POST` | `/api/v1/jobs/{id}/cancel` | Cancel job |

---

## Database Schema

```
bios_profiles       — reusable setting templates
servers             — physical server inventory
bios_settings       — BIOS snapshots per server (JSONB, historical)
firmware_inventory  — per-component version + approval status
drift_events        — detected configuration changes
change_requests     — approval workflow entries
snapshots           — point-in-time full backups
compliance_reports  — scored compliance results
jobs                — bulk fleet operation tracking
```

---

## Supported Vendors

| Vendor | Collection | Apply |
|---|---|---|
| Dell | `dmidecode` + `racadm` stub | `racadm set BIOS.*` |
| HPE | `dmidecode` + iLO REST stub | iLO PATCH |
| Lenovo | `dmidecode` + OneCLI stub | OneCLI set |
| Supermicro | `dmidecode` + IPMI stub | IPMI raw |

Vendor stubs are in [internal/agent/collector.go](internal/agent/collector.go).

---

## Roadmap

### Phase 1 — Foundation (current)
- [x] Hardware discovery via `dmidecode`
- [x] BIOS inventory collection and storage
- [x] REST API with Bearer token auth
- [x] Compliance reporting
- [x] Drift detection

### Phase 2 — Policy Engine
- [ ] Policy templates (CIS Benchmark, NIST, corporate baseline)
- [ ] Per-device-class policy assignment (PC vs server vs laptop)
- [ ] Scheduled compliance scans
- [ ] Email / webhook alerts on drift

### Phase 3 — Remote Configuration
- [ ] Remote BIOS setting changes via vendor CLI (Dell `racadm`, HPE iLO, Lenovo OneCLI)
- [ ] Firmware update orchestration
- [ ] Bulk operations with rollback on failure
- [ ] Pre/post-change snapshot automation

### Phase 4 — Endpoint Management
- [ ] Windows PC agent (WMI + PowerShell BIOS bridge)
- [ ] Laptop fleet support (battery policy, lid/power settings)
- [ ] Hardware lifecycle tracking (age, warranty, EOL)
- [ ] Security automation (auto-enforce Secure Boot, TPM enrollment)

### Phase 5 — Intelligence
- [ ] AI recommendation engine (workload-aware BIOS tuning)
- [ ] Predictive compliance (flag devices likely to drift)
- [ ] Anomaly detection (unexpected BIOS changes)
- [ ] Autonomous remediation with approval bypass for critical security settings

---

## Project Name Alternatives

**Enterprise:**
FirmwareOps · FleetConfig · FirmwareFleet · Hardware Control Plane · Firmware Governance Platform · Platform Firmware Manager

**Open-source:**
`firmwarectl` · `fleetctl` · `hwconfig` · `biosctl` · `firmwared` · `fwfleet`

---

## License

MIT
