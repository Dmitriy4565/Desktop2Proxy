package scanners

import (
	"context"
	"desktop2proxy/models" // Правильный импорт
	"fmt"
	"net"
	"strconv"
	"time"
)

type RDPScanner struct{}

func (s *RDPScanner) GetName() string {
	return "RDP"
}

func (s *RDPScanner) GetDefaultPort() int {
	return 3389
}

func (s *RDPScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult { // Исправлены типы
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return models.ProbeResult{ // Исправлено на models.ProbeResult
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer conn.Close()

	// RDP обычно отправляет banner при подключении
	banner := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(banner)
	if err != nil {
		// Для RDP иногда просто успешное подключение - уже показатель
		return models.ProbeResult{ // Исправлено на models.ProbeResult
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   "RDP port open (banner read failed)",
		}
	}

	bannerStr := string(banner[:n])
	return models.ProbeResult{ // Исправлено на models.ProbeResult
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   fmt.Sprintf("RDP banner: %s", bannerStr),
	}
}
