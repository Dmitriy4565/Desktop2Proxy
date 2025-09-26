package models

type Target struct {
	IP       string
	Username string
	Password string
}

type ProbeResult struct {
	Protocol   string
	Port       int
	Success    bool
	Banner     string
	Error      string
	DeviceInfo *DeviceInfo // Новая строка - информация об устройстве
}

type DeviceInfo struct {
	OS         string // Windows, Linux, Router, Switch, etc.
	DeviceType string // Server, Workstation, NetworkDevice, IoT
	Vendor     string // Cisco, Microsoft, Dell, etc.
	Model      string // Specific model if detectable
}
