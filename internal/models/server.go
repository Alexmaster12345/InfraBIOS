package models

import (
	"encoding/json"
	"time"
)

type ServerStatus string
type DeviceType string

const (
	StatusActive      ServerStatus = "active"
	StatusMaintenance ServerStatus = "maintenance"
	StatusOffline     ServerStatus = "offline"

	DeviceTypeServer      DeviceType = "server"
	DeviceTypeDesktop     DeviceType = "desktop"    // Windows/Linux PC (Dell OptiPlex, HP EliteDesk)
	DeviceTypeWorkstation DeviceType = "workstation" // Engineering workstation (Linux)
	DeviceTypeLaptop      DeviceType = "laptop"     // ThinkPad, Latitude, EliteBook
	DeviceTypeUnknown     DeviceType = "unknown"
)

type Server struct {
	ID           string       `json:"id"`
	Hostname     string       `json:"hostname"`
	IPAddress    string       `json:"ip_address"`
	Vendor       string       `json:"vendor"`
	Model        string       `json:"model"`
	SerialNumber string       `json:"serial_number"`
	BIOSVersion  string       `json:"bios_version"`
	BMCVersion   string       `json:"bmc_version"`
	DeviceType   DeviceType   `json:"device_type"`
	OS           string       `json:"os"`
	ProfileID    *string      `json:"profile_id,omitempty"`
	Status       ServerStatus `json:"status"`
	LastSeen     *time.Time   `json:"last_seen,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type CreateServerRequest struct {
	Hostname     string     `json:"hostname"`
	IPAddress    string     `json:"ip_address"`
	Vendor       string     `json:"vendor"`
	Model        string     `json:"model"`
	SerialNumber string     `json:"serial_number"`
	BIOSVersion  string     `json:"bios_version"`
	BMCVersion   string     `json:"bmc_version"`
	DeviceType   DeviceType `json:"device_type"`
	OS           string     `json:"os"`
}

type BIOSSettings struct {
	ID           string          `json:"id"`
	ServerID     string          `json:"server_id"`
	Settings     json.RawMessage `json:"settings"`
	AgentVersion string          `json:"agent_version"`
	CollectedAt  time.Time       `json:"collected_at"`
}

type SubmitInventoryRequest struct {
	Settings     json.RawMessage `json:"settings"`
	Firmware     json.RawMessage `json:"firmware,omitempty"`
	AgentVersion string          `json:"agent_version"`
}
