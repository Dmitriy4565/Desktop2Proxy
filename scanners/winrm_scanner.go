package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
	"time"

	"github.com/masterzen/winrm"
)

type WinRMScanner struct {
	UseHTTPS bool
}

func (s *WinRMScanner) GetName() string {
	if s.UseHTTPS {
		return "WinRM-HTTPS"
	}
	return "WinRM-HTTP"
}

func (s *WinRMScanner) GetDefaultPort() int {
	if s.UseHTTPS {
		return 5986
	}
	return 5985
}

func (s *WinRMScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	// Создаем endpoint правильным способом
	endpoint := winrm.NewEndpoint(
		target.IP,
		port,
		s.UseHTTPS, // HTTPS
		true,       // Insecure
		nil,        // CA cert
		nil,        // Client cert
		nil,        // Client key
		5*time.Second,
	)

	// Создаем клиент правильным способом
	client, err := winrm.NewClientWithParameters(
		endpoint,
		target.Username,
		target.Password,
		winrm.DefaultParameters,
	)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("WinRM client creation failed: %v", err),
		}
	}

	// Пробуем выполнить команду
	cmd := "echo WinRM-test-success"
	_, err = client.Run(cmd, nil, nil)

	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("WinRM connection failed: %v", err),
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   "WinRM accessible - command executed successfully",
	}
}
