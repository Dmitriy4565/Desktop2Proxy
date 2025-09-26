package scanners

import (
	"desktop2proxy/models"
	"strings"
)

// AnalyzeDeviceInfo анализирует баннеры и определяет тип устройства
func AnalyzeDeviceInfo(results []models.ProbeResult) *models.DeviceInfo {
	info := &models.DeviceInfo{}

	for _, result := range results {
		if result.Success {
			detectFromBanner(info, result.Banner, result.Protocol, result.Port)
		}
	}

	// Если не удалось определить, ставим Unknown
	if info.OS == "" {
		info.OS = "Unknown"
	}
	if info.DeviceType == "" {
		info.DeviceType = "Unknown"
	}

	return info
}

func detectFromBanner(info *models.DeviceInfo, banner string, protocol string, port int) {
	banner = strings.ToLower(banner)

	// Определение по HTTP/HTTPS
	if protocol == "HTTP" || protocol == "HTTPS" {
		if strings.Contains(banner, "iis") || strings.Contains(banner, "microsoft") {
			info.OS = "Windows"
			info.Vendor = "Microsoft"
			info.DeviceType = "Web Server"
		}
		if strings.Contains(banner, "apache") || strings.Contains(banner, "nginx") {
			info.OS = "Linux/Unix"
			info.DeviceType = "Web Server"
		}
		if strings.Contains(banner, "routeros") || strings.Contains(banner, "mikrotik") {
			info.OS = "RouterOS"
			info.Vendor = "MikroTik"
			info.DeviceType = "Router"
		}
	}

	// Определение по SSH
	if protocol == "SSH" {
		if strings.Contains(banner, "openssh") {
			info.OS = "Linux/Unix"
			info.DeviceType = "Server"
		}
		if strings.Contains(banner, "cisco") {
			info.OS = "Cisco IOS"
			info.Vendor = "Cisco"
			info.DeviceType = "Network Device"
		}
	}

	// Определение по WinRM
	if protocol == "WinRM-HTTP" || protocol == "WinRM-HTTPS" {
		info.OS = "Windows"
		info.Vendor = "Microsoft"
		info.DeviceType = "Windows Server/Workstation"
	}

	// Определение по RDP
	if protocol == "RDP" {
		info.OS = "Windows"
		info.Vendor = "Microsoft"
		info.DeviceType = "Windows Machine"
	}

	// Определение по портам
	switch port {
	case 22:
		if info.DeviceType == "" {
			info.DeviceType = "SSH Server"
		}
	case 80, 443:
		if info.DeviceType == "" {
			info.DeviceType = "Web Server"
		}
	case 3389:
		info.OS = "Windows"
		info.DeviceType = "Remote Desktop Host"
	case 5985, 5986:
		info.OS = "Windows"
		info.DeviceType = "Windows Management Host"
	case 23:
		info.DeviceType = "Telnet Server (often Network Device)"
	case 161:
		info.DeviceType = "SNMP Device (Network/Server)"
	}
}
