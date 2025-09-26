package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
)

// ScannerManager управляет всеми сканерами
type ScannerManager struct {
	scanners []Scanner
}

// NewScannerManager создает менеджер со всеми доступными сканерами
func NewScannerManager() *ScannerManager {
	return &ScannerManager{
		scanners: []Scanner{
			&SSHScanner{},
			&TelnetScanner{},
			&HTTPScanner{Protocol: "HTTP"},
			&HTTPScanner{Protocol: "HTTPS"},
			&SNMPScanner{},
			&WinRMScanner{UseHTTPS: false},
			&WinRMScanner{UseHTTPS: true},
			&RDPScanner{},
		},
	}
}

// GetAllScanners возвращает все зарегистрированные сканеры
func (sm *ScannerManager) GetAllScanners() []Scanner {
	return sm.scanners
}

// GetScannerByName возвращает сканер по имени протокола
func (sm *ScannerManager) GetScannerByName(name string) Scanner {
	for _, scanner := range sm.scanners {
		if scanner.GetName() == name {
			return scanner
		}
	}
	return nil
}

// ProbeProtocols проверяет все протоколы параллельно
func ProbeProtocols(target models.Target, scanners []Scanner) *models.ProbeResult {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan models.ProbeResult, len(scanners))

	for _, scanner := range scanners {
		go func(s Scanner) {
			result := s.CheckProtocol(ctx, target, s.GetDefaultPort())
			resultChan <- result
		}(scanner)
	}

	for range scanners {
		result := <-resultChan
		if result.Success {
			cancel()
			return &result
		} else {
			fmt.Printf("❌ %s:%d - %s\n", result.Protocol, result.Port, result.Error)
		}
	}

	return nil
}
