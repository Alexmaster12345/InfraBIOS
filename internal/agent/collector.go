package agent

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// HardwareInfo is the basic machine identity collected on startup.
type HardwareInfo struct {
	Vendor       string `json:"vendor"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	BIOSVersion  string `json:"bios_version"`
	BMCVersion   string `json:"bmc_version"`
}

// BIOSSnapshot is a map of BIOS setting keys to their current values.
type BIOSSnapshot map[string]any

// CollectHardware reads basic hardware identity via dmidecode.
// Falls back to safe defaults if dmidecode is unavailable.
func CollectHardware() HardwareInfo {
	info := HardwareInfo{
		Vendor:      dmidecode("system-manufacturer"),
		Model:       dmidecode("system-product-name"),
		SerialNumber: dmidecode("system-serial-number"),
		BIOSVersion: dmidecode("bios-version"),
	}
	return info
}

// CollectBIOSSettings reads BIOS settings using vendor-specific tooling.
// Currently returns a simulated snapshot; extend per-vendor as needed.
func CollectBIOSSettings(vendor string) (json.RawMessage, error) {
	var snapshot BIOSSnapshot

	switch strings.ToLower(vendor) {
	case "dell inc.", "dell":
		snapshot = collectDell()
	case "hpe", "hewlett packard enterprise", "hp":
		snapshot = collectHPE()
	default:
		snapshot = collectGeneric()
	}

	return json.Marshal(snapshot)
}

// dmidecode runs dmidecode and returns trimmed output.
func dmidecode(flag string) string {
	out, err := exec.Command("dmidecode", "-s", flag).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func collectDell() BIOSSnapshot {
	// Production: exec racadm get BIOS.* or use iDRAC REST API
	return BIOSSnapshot{
		"virtualization_technology": readBoolSetting("VirtualizationTechnology", true),
		"vt_d":                      readBoolSetting("VtForDirectIo", true),
		"secure_boot":               readBoolSetting("SecureBoot", true),
		"hyperthreading":            readBoolSetting("LogicalProc", true),
		"tpm_security":              readBoolSetting("TpmSecurity", true),
		"boot_mode":                 "Uefi",
		"numa":                      readBoolSetting("NodeInterleave", false),
		"sr_iov":                    readBoolSetting("SriovGlobalEnable", false),
	}
}

func collectHPE() BIOSSnapshot {
	// Production: use iLO REST API or hponcfg
	return BIOSSnapshot{
		"intel_vt":    true,
		"intel_vt_d":  true,
		"secure_boot": true,
		"hyperthreading": true,
		"tpm_module":  true,
		"boot_mode":   "UEFI",
	}
}

func collectGeneric() BIOSSnapshot {
	// Generic: dmidecode can read some values
	return BIOSSnapshot{
		"bios_vendor":  dmidecode("bios-vendor"),
		"bios_version": dmidecode("bios-version"),
		"bios_date":    dmidecode("bios-release-date"),
	}
}

// readBoolSetting is a stub — in production this would exec a vendor CLI
// and parse the output. Here it returns the defaultVal.
func readBoolSetting(_ string, defaultVal bool) bool {
	return defaultVal
}
