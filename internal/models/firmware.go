package models

import "time"

type FirmwareStatus string

const (
	FirmwareApproved FirmwareStatus = "approved"
	FirmwareOutdated FirmwareStatus = "outdated"
	FirmwareUnknown  FirmwareStatus = "unknown"
)

type FirmwareEntry struct {
	ID              string         `json:"id"`
	ServerID        string         `json:"server_id"`
	Component       string         `json:"component"` // bios, bmc, nic, storage
	CurrentVersion  string         `json:"current_version"`
	ApprovedVersion string         `json:"approved_version"`
	Status          FirmwareStatus `json:"status"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type UpsertFirmwareRequest struct {
	Component       string `json:"component"`
	CurrentVersion  string `json:"current_version"`
	ApprovedVersion string `json:"approved_version"`
}
