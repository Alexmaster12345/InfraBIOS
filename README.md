# InfraBIOS

**BIOS Configuration Management System**

A production-grade platform that centrally manages, audits, validates, and automates BIOS/UEFI settings across thousands of physical servers — replacing manual SSH firewall-style management with a programmable REST API backed by a full audit trail.

> Think of it as **Git + Ansible + Fleet Management**, but specifically for BIOS and firmware configuration.

---

## The Problem

Large datacenters suffer from configuration drift across mixed vendor fleets:

| Server | Secure Boot | VT-x | Hyperthreading |
|---|---|---|---|
| lab-001 | ✅ Enabled | ✅ Enabled | ✅ Enabled |
| lab-002 | ❌ Disabled | ✅ Enabled | ✅ Enabled |
| lab-003 | ✅ Enabled | ❌ Disabled | ✅ Enabled |

These inconsistencies cause performance degradation, security gaps, deployment failures, and compliance violations.

InfraBIOS provides a **single source of truth** — define a profile once, enforce it everywhere.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                       Operators / CI                         │
│              REST API  ·  curl  ·  Terraform  ·  Ansible     │
└───────────────────────────┬──────────────────────────────────┘
                            │ Bearer token (HTTPS)
┌───────────────────────────▼──────────────────────────────────┐
│                    InfraBIOS Server                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐  │
│  │ /servers │ │/profiles │ │  /drift  │ │  /compliance   │  │
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
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
┌────────▼───────┐   ┌──────────▼──────┐   ┌──────────▼──────┐
│  infrabios-    │   │  infrabios-     │   │  infrabios-     │
│  agent         │   │  agent          │   │  agent          │
│  Dell R760     │   │  HPE DL380      │   │  Supermicro X13 │
└────────────────┘   └─────────────────┘   └─────────────────┘
```

---

## Components

| Component | Description |
|---|---|
| **Hardware Discovery** | `dmidecode` + vendor CLI stubs (Dell `racadm`, HPE iLO, Lenovo OneCLI) |
| **BIOS Inventory DB** | JSONB-stored settings per server, full history |
| **Profile Management** | Reusable templates — Kubernetes, Security, Virtualization, AI |
| **Compliance Engine** | Compare actual vs profile → score + violation report |
| **Drift Detection** | Per-collection diff → emits drift events on any change |
| **Firmware Lifecycle** | Track BIOS/BMC/NIC versions and approval status |
| **Change Management** | Full approval workflow (pending → approved → applied) |
| **Snapshots** | Point-in-time backup of settings + firmware |
| **Job Queue** | Bulk operations across 10–10,000 servers |

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

## License

MIT
