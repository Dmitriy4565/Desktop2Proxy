package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type VNCScanner struct{}

func (s *VNCScanner) GetName() string {
	return "VNC"
}

func (s *VNCScanner) GetDefaultPort() int {
	return 5900
}

func (s *VNCScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    "Сервис недоступен",
		}
	}
	defer conn.Close()

	// Читаем VNC баннер (VNC обычно отправляет версию при подключении)
	banner := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(banner)

	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   "Порт открыт (возможно VNC)",
			DeviceInfo: &models.DeviceInfo{
				OS:         "Unknown",
				DeviceType: "Remote Desktop",
				Vendor:     "VNC",
			},
		}
	}

	bannerStr := strings.TrimSpace(string(banner[:n]))
	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   fmt.Sprintf("VNC протокол: %s", bannerStr),
		DeviceInfo: &models.DeviceInfo{
			OS:         "Unknown",
			DeviceType: "VNC Server",
			Vendor:     "VNC",
		},
	}
}
