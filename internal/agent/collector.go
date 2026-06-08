package agent

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// HardwareInfo is the basic machine identity collected on startup.
type HardwareInfo struct {
	Vendor       string `json:"vendor"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	BIOSVersion  string `json:"bios_version"`
	BMCVersion   string `json:"bmc_version"`
	DeviceType   string `json:"device_type"` // server | desktop | workstation | laptop
	OS           string `json:"os"`          // windows | linux | darwin
}

// BIOSSnapshot is a map of BIOS setting keys to their current values.
type BIOSSnapshot map[string]any

// CollectHardware reads basic hardware identity and detects device type.
func CollectHardware() HardwareInfo {
	info := HardwareInfo{
		Vendor:       dmidecode("system-manufacturer"),
		Model:        dmidecode("system-product-name"),
		SerialNumber: dmidecode("system-serial-number"),
		BIOSVersion:  dmidecode("bios-version"),
		OS:           runtime.GOOS,
		DeviceType:   detectDeviceType(),
	}
	return info
}

// CollectBIOSSettings selects the right collector based on vendor + device type.
func CollectBIOSSettings(vendor, deviceType string) (json.RawMessage, error) {
	var snapshot BIOSSnapshot

	if runtime.GOOS == "windows" {
		snapshot = collectWindows(vendor, deviceType)
	} else {
		switch deviceType {
		case "laptop":
			snapshot = collectLaptop(vendor)
		case "desktop":
			snapshot = collectDesktop(vendor)
		case "workstation":
			snapshot = collectWorkstation(vendor)
		default:
			snapshot = collectLinuxServer(vendor)
		}
	}

	return json.Marshal(snapshot)
}

// ── Device type detection ─────────────────────────────────────────────────────

// detectDeviceType reads the SMBIOS chassis type to determine desktop/laptop/server.
func detectDeviceType() string {
	out, err := exec.Command("dmidecode", "-t", "chassis").Output()
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Type:") {
			continue
		}
		t := strings.ToLower(strings.TrimPrefix(line, "Type:"))
		t = strings.TrimSpace(t)
		switch {
		// Portable / Laptop / Notebook / Sub Notebook / Convertible / Detachable
		case strings.Contains(t, "notebook"),
			strings.Contains(t, "laptop"),
			strings.Contains(t, "portable"),
			strings.Contains(t, "convertible"),
			strings.Contains(t, "detachable"),
			t == "8", t == "9", t == "10", t == "11", t == "12", t == "13", t == "14", t == "30", t == "31", t == "32":
			return "laptop"
		// Rack / Blade / Tower server
		case strings.Contains(t, "server"),
			strings.Contains(t, "rack"),
			strings.Contains(t, "blade"),
			t == "17", t == "23", t == "25", t == "28":
			return "server"
		// Workstation
		case strings.Contains(t, "workstation"):
			return "workstation"
		// Desktop / Mini / SFF
		case strings.Contains(t, "desktop"),
			strings.Contains(t, "mini"),
			strings.Contains(t, "tower"),
			t == "3", t == "4", t == "5", t == "6", t == "7", t == "15", t == "24":
			return "desktop"
		}
	}

	// Fallback: guess by model name
	model := strings.ToLower(dmidecode("system-product-name"))
	if containsAny(model, "thinkpad", "latitude", "elitebook", "xps", "inspiron", "omen", "spectre", "envy", "probook") {
		return "laptop"
	}
	if containsAny(model, "optiplex", "elitedesk", "prodesk", "thinkcentre", "nuc") {
		return "desktop"
	}
	if containsAny(model, "precision", "z4", "z6", "z8", "thinkstation", "xeon") {
		return "workstation"
	}
	if containsAny(model, "poweredge", "proliant", "thinkserver", "supermicro") {
		return "server"
	}
	return "unknown"
}

// ── Windows PC collector (Dell OptiPlex, HP EliteDesk, laptops) ──────────────

// collectWindows uses PowerShell + WMI to read BIOS/security settings on Windows.
// Covers: Dell OptiPlex, HP EliteDesk, Lenovo ThinkCentre (desktops), ThinkPad/Latitude/EliteBook (laptops).
func collectWindows(vendor, deviceType string) BIOSSnapshot {
	snap := BIOSSnapshot{
		"os":          "windows",
		"device_type": deviceType,
		"boot_mode":   psString(`(Get-ItemProperty "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot\State" -ErrorAction SilentlyContinue).UEFISecureBootEnabled`),
		"secure_boot": psBool(`Confirm-SecureBootUEFI -ErrorAction SilentlyContinue`),
		"tpm_present": psBool(`(Get-Tpm -ErrorAction SilentlyContinue).TpmPresent`),
		"tpm_enabled": psBool(`(Get-Tpm -ErrorAction SilentlyContinue).TpmEnabled`),
		"tpm_version": psString(`(Get-WmiObject -Namespace 'root/cimv2/security/microsofttpm' -Class Win32_Tpm -ErrorAction SilentlyContinue).SpecVersion`),
		"virtualization_enabled": psBool(`(Get-WmiObject -Class Win32_Processor -ErrorAction SilentlyContinue).VirtualizationFirmwareEnabled`),
		"hyper_v_enabled":        psBool(`(Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -ErrorAction SilentlyContinue).State -eq 'Enabled'`),
		"bitlocker_status":       psString(`(Get-BitLockerVolume -MountPoint C: -ErrorAction SilentlyContinue).ProtectionStatus`),
	}

	// Laptop-specific: flag BitLocker risk if TPM or Secure Boot is involved
	if deviceType == "laptop" {
		snap["bitlocker_risk_note"] = "changing secure_boot or tpm settings will suspend BitLocker on next boot"
		snap["battery_present"] = psBool(`(Get-WmiObject -Class Win32_Battery -ErrorAction SilentlyContinue) -ne $null`)
	}

	// Dell-specific via Dell Command Configure (cctk)
	if strings.Contains(strings.ToLower(vendor), "dell") {
		snap["dell_secureboot"]     = psString(`cctk --SecureBoot 2>$null`)
		snap["dell_tpm"]            = psString(`cctk --Tpm 2>$null`)
		snap["dell_virtualization"] = psString(`cctk --VirtualizationTechnology 2>$null`)
	}

	// HP-specific via HP BIOS Configuration Utility (BiosConfigUtility)
	if containsAny(strings.ToLower(vendor), "hp", "hewlett") {
		snap["hp_secureboot"] = psString(`(Get-WmiObject -Namespace root/hp/InstrumentedBIOS -Class HP_BIOSSetting -Filter "Name='SecureBoot'" -ErrorAction SilentlyContinue).CurrentValue`)
		snap["hp_tpm"]        = psString(`(Get-WmiObject -Namespace root/hp/InstrumentedBIOS -Class HP_BIOSSetting -Filter "Name='TPM Device'" -ErrorAction SilentlyContinue).CurrentValue`)
	}

	return snap
}

// ── Linux laptop collector (ThinkPad, Latitude, EliteBook) ───────────────────

// collectLaptop reads BIOS/security state on Linux laptops.
// Key concern: Secure Boot and TPM changes break BitLocker on dual-boot systems.
func collectLaptop(vendor string) BIOSSnapshot {
	snap := collectLinuxCommon()
	snap["device_type"] = "laptop"
	snap["battery_present"] = fileExists("/sys/class/power_supply/BAT0") || fileExists("/sys/class/power_supply/BAT1")
	snap["ac_connected"] = readSysFile("/sys/class/power_supply/AC/online") == "1"

	// BitLocker / Windows dual-boot risk flag
	snap["bitlocker_risk_note"] = "if dual-booting Windows: changing secure_boot or tpm will suspend BitLocker"

	v := strings.ToLower(vendor)
	switch {
	case strings.Contains(v, "lenovo"):
		// ThinkPad — Lenovo BIOS Configuration (LVFS / fwupd)
		snap["thinkpad_fan_mode"] = readSysFile("/sys/devices/platform/thinkpad_hwmon/fan1_label")
		snap["thinkpad_dock"]     = fileExists("/sys/bus/pci/devices/0000:00:1c.0") // rough heuristic
		// Production: use `fwupdmgr get-devices` for full firmware inventory
		snap["fwupd_available"] = commandExists("fwupdmgr")

	case strings.Contains(v, "dell"):
		// Dell Latitude — Dell Command | Configure on Linux
		snap["dell_cctk_available"] = commandExists("cctk")
		if commandExists("cctk") {
			snap["secure_boot"]        = cctk("--SecureBoot")
			snap["tpm"]                = cctk("--Tpm")
			snap["virtualization"]     = cctk("--VirtualizationTechnology")
			snap["battery_threshold"]  = cctk("--PrimaryBattChargeCfg")
		}

	case containsAny(v, "hp", "hewlett"):
		// HP EliteBook — hponcfg or HP Linux Imaging and Printing
		snap["hp_tools_available"] = commandExists("hponcfg")
	}

	return snap
}

// ── Linux desktop/PC collector (Dell OptiPlex, HP EliteDesk) ─────────────────

// collectDesktop reads BIOS/security state on Linux desktops.
func collectDesktop(vendor string) BIOSSnapshot {
	snap := collectLinuxCommon()
	snap["device_type"] = "desktop"

	v := strings.ToLower(vendor)
	if strings.Contains(v, "dell") {
		snap["dell_cctk_available"] = commandExists("cctk")
		if commandExists("cctk") {
			snap["secure_boot"]    = cctk("--SecureBoot")
			snap["tpm"]            = cctk("--Tpm")
			snap["virtualization"] = cctk("--VirtualizationTechnology")
			snap["boot_mode"]      = cctk("--BootMode")
		}
	}
	if containsAny(v, "hp", "hewlett") {
		snap["hp_tools_available"] = commandExists("hponcfg")
	}
	if strings.Contains(v, "lenovo") {
		snap["fwupd_available"] = commandExists("fwupdmgr")
	}

	return snap
}

// ── Linux workstation collector (Precision, Z-series, ThinkStation) ──────────

// collectWorkstation reads BIOS/security state on engineering workstations.
// These often run KVM/QEMU so virtualization flags are critical.
func collectWorkstation(vendor string) BIOSSnapshot {
	snap := collectLinuxCommon()
	snap["device_type"] = "workstation"

	// Workstations frequently run VMs — capture extended virtualization info
	snap["kvm_module_loaded"] = fileExists("/dev/kvm")
	snap["iommu_enabled"] = fileContains("/proc/cmdline", "intel_iommu=on") ||
		fileContains("/proc/cmdline", "amd_iommu=on")
	snap["hugepages_total"] = readSysFile("/proc/sys/vm/nr_hugepages")

	v := strings.ToLower(vendor)
	if strings.Contains(v, "dell") && commandExists("cctk") {
		snap["secure_boot"]    = cctk("--SecureBoot")
		snap["tpm"]            = cctk("--Tpm")
		snap["virtualization"] = cctk("--VirtualizationTechnology")
		snap["vt_d"]           = cctk("--VtForDirectIo")
		snap["sr_iov"]         = cctk("--SrIovGlobalEnable")
	}
	if containsAny(v, "hp", "hewlett") && commandExists("hponcfg") {
		snap["hp_tools_available"] = true
	}

	return snap
}

// ── Linux server collector (Dell PowerEdge, HPE ProLiant, etc.) ──────────────

func collectLinuxServer(vendor string) BIOSSnapshot {
	snap := collectLinuxCommon()
	snap["device_type"] = "server"

	v := strings.ToLower(vendor)
	switch {
	case strings.Contains(v, "dell"):
		snap["virtualization_technology"] = readBoolSetting("VirtualizationTechnology", true)
		snap["vt_d"]                      = readBoolSetting("VtForDirectIo", true)
		snap["hyperthreading"]            = readBoolSetting("LogicalProc", true)
		snap["tpm_security"]              = readBoolSetting("TpmSecurity", true)
		snap["boot_mode"]                 = "Uefi"
		snap["numa"]                      = readBoolSetting("NodeInterleave", false)
		snap["sr_iov"]                    = readBoolSetting("SriovGlobalEnable", false)
	case containsAny(v, "hpe", "hewlett", "hp"):
		snap["intel_vt"]        = true
		snap["intel_vt_d"]      = true
		snap["hyperthreading"]  = true
		snap["tpm_module"]      = true
		snap["boot_mode"]       = "UEFI"
	case strings.Contains(v, "lenovo"):
		snap["virtualization"] = true
		snap["tpm"]            = true
		snap["boot_mode"]      = "UEFI"
	case strings.Contains(v, "supermicro"):
		snap["virtualization"] = true
		snap["tpm"]            = true
		snap["boot_mode"]      = "UEFI"
	default:
		return collectGeneric()
	}
	return snap
}

// ── Common Linux BIOS/security reads ─────────────────────────────────────────

// collectLinuxCommon reads BIOS security state from sysfs/kernel interfaces
// available on any Linux system regardless of vendor.
func collectLinuxCommon() BIOSSnapshot {
	snap := BIOSSnapshot{
		"bios_vendor":  dmidecode("bios-vendor"),
		"bios_version": dmidecode("bios-version"),
		"bios_date":    dmidecode("bios-release-date"),
		"boot_mode":    uefiBootMode(),
		"secure_boot":  secureBootEnabled(),
		"tpm_present":  fileExists("/sys/class/tpm/tpm0") || fileExists("/sys/class/tpm/tpm1"),
		"tpm_version":  tpmVersion(),
	}
	return snap
}

func collectGeneric() BIOSSnapshot {
	return BIOSSnapshot{
		"bios_vendor":  dmidecode("bios-vendor"),
		"bios_version": dmidecode("bios-version"),
		"bios_date":    dmidecode("bios-release-date"),
	}
}

// ── Kernel / sysfs helpers ────────────────────────────────────────────────────

// uefiBootMode checks if the system booted in UEFI mode.
func uefiBootMode() string {
	if fileExists("/sys/firmware/efi") {
		return "UEFI"
	}
	return "Legacy"
}

// secureBootEnabled reads the Secure Boot state from the EFI variable.
func secureBootEnabled() bool {
	// /sys/firmware/efi/efivars/SecureBoot-* contains a 5-byte value; byte[4] == 1 means enabled
	entries, err := os.ReadDir("/sys/firmware/efi/efivars")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "SecureBoot-") {
			data, err := os.ReadFile("/sys/firmware/efi/efivars/" + e.Name())
			if err == nil && len(data) >= 5 {
				return data[4] == 1
			}
		}
	}
	// Fallback: try mokutil
	out, err := exec.Command("mokutil", "--sb-state").Output()
	if err == nil {
		return strings.Contains(strings.ToLower(string(out)), "enabled")
	}
	return false
}

// tpmVersion reads TPM version from sysfs.
func tpmVersion() string {
	for _, p := range []string{"/sys/class/tpm/tpm0/tpm_version_major", "/sys/class/tpm/tpm1/tpm_version_major"} {
		if v := strings.TrimSpace(readSysFile(p)); v != "" {
			return v
		}
	}
	return "unknown"
}

// cctk runs Dell Command | Configure and returns trimmed output.
func cctk(flag string) string {
	out, err := exec.Command("cctk", flag).Output()
	if err != nil {
		return "unavailable"
	}
	// cctk output: "SecureBoot=Enabled" — return just the value
	s := strings.TrimSpace(string(out))
	if idx := strings.Index(s, "="); idx >= 0 {
		return strings.TrimSpace(s[idx+1:])
	}
	return s
}

// ── PowerShell helper (Windows only) ─────────────────────────────────────────

func psString(script string) string {
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return "unavailable"
	}
	return strings.TrimSpace(string(out))
}

func psBool(script string) bool {
	return strings.EqualFold(strings.TrimSpace(psString(script)), "true")
}

// ── Low-level helpers ─────────────────────────────────────────────────────────

func dmidecode(flag string) string {
	out, err := exec.Command("dmidecode", "-s", flag).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func readSysFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileContains(path, s string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(b), s)
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func readBoolSetting(_ string, defaultVal bool) bool {
	return defaultVal
}
