package scanners

import (
	"context"
	"desktop2proxy/models"
	"net"
	"strconv"
	"strings"
	"time"
)

type TelnetScanner struct{}

func (s *TelnetScanner) GetName() string {
	return "Telnet"
}

func (s *TelnetScanner) GetDefaultPort() int {
	return 23
}

func (s *TelnetScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer conn.Close()

	banner := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(banner)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    err.Error(),
		}
	}

	bannerStr := strings.ToLower(string(banner[:n]))
	if strings.Contains(bannerStr, "login") || strings.Contains(bannerStr, "password") {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   "Telnet service detected",
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  false,
		Error:    "Does not look like a Telnet service",
	}
}
