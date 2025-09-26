package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
	"io"
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
	// Создаем endpoint
	endpoint := &winrm.Endpoint{
		Host:     target.IP,
		Port:     port,
		HTTPS:    s.UseHTTPS,
		Insecure: true,
		Timeout:  10 * time.Second,
	}

	// Создаем клиент
	client, err := winrm.NewClient(endpoint, target.Username, target.Password)
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
	exitCode, err := client.RunWithContext(ctx, cmd, io.Discard, io.Discard)

	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("WinRM connection failed: %v", err),
		}
	}

	// Exit code 0 обычно означает успех
	if exitCode == 0 {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   "WinRM accessible - command executed successfully",
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  false,
		Error:    fmt.Sprintf("Command exited with code: %d", exitCode),
	}
}
