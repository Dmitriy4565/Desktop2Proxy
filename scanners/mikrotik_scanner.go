package scanners

import (
	"bufio"
	"context"
	"desktop2proxy/models"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type MikroTikScanner struct{}

func (s *MikroTikScanner) GetName() string {
	return "MikroTik"
}

func (s *MikroTikScanner) GetDefaultPort() int {
	return 8728 // API порт MikroTik
}

func (s *MikroTikScanner) GetCommonPorts() []int {
	return []int{
		8728,  // API порт (основной)
		8729,  // API SSL порт
		8291,  // WinBox
		22,    // SSH
		23,    // Telnet
		80,    // HTTP веб-интерфейс
		443,   // HTTPS веб-интерфейс
		21,    // FTP
		2000,  // Bandwidth test
		20561, // MAC Telnet
		5678,  // MikroTik Neighbor Discovery
		135,   // WinBox alternative
		139,   // WinBox alternative
		445,   // WinBox alternative
		5678,  // MikroTik API
		1701,  // PPTP
		1723,  // PPTP
		1812,  // RADIUS
		1813,  // RADIUS accounting
		500,   // IPSec
		4500,  // IPSec NAT-T
	}
}

func (s *MikroTikScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))

	// Проверяем подключение к порту
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer conn.Close()

	// Читаем баннер MikroTik
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	scanner := bufio.NewScanner(conn)

	if scanner.Scan() {
		banner := scanner.Text()

		// Проверяем характерные признаки MikroTik
		if strings.Contains(banner, "MikroTik") ||
			strings.Contains(strings.ToLower(banner), "routeros") ||
			strings.Contains(banner, "!done") {

			return models.ProbeResult{
				Protocol: s.GetName(),
				Port:     port,
				Success:  true,
				Banner:   fmt.Sprintf("MikroTik banner: %s", banner),
			}
		}
	}

	// Если порт открыт, но баннер не распознан, все равно считаем успехом
	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   "Port open (possible MikroTik)",
	}
}
