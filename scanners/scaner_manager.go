package scanners

import (
	"context"
	"desktop2proxy/models"
)

// ScannerManager управляет всеми сканерами
type ScannerManager struct {
	scanners []Scanner
}

// NewScannerManager создает менеджер со всеми доступными сканерами
func NewScannerManager() *ScannerManager {
	return &ScannerManager{
		scanners: []Scanner{
			// Реальные сканеры
			&HTTPScanner{Protocol: "HTTP"},
			&HTTPScanner{Protocol: "HTTPS"},
			&SSHScanner{},
			&WinRMScanner{UseHTTPS: false},
			&WinRMScanner{UseHTTPS: true},
			&RDPScanner{},
			&TelnetScanner{},
			&SNMPScanner{},
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

// ProbeAllProtocols проверяет ВСЕ протоколы с указанными логином/паролем
func ProbeAllProtocols(target models.Target, scanners []Scanner) []models.ProbeResult {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan models.ProbeResult, len(scanners))
	var successfulConnections []models.ProbeResult

	// Параллельно проверяем все протоколы с ОДНИМИ логином/паролем
	for _, scanner := range scanners {
		go func(s Scanner) {
			result := s.CheckProtocol(ctx, target, s.GetDefaultPort())
			resultChan <- result
		}(scanner)
	}

	// Собираем все успешные подключения
	for i := 0; i < len(scanners); i++ {
		result := <-resultChan
		if result.Success {
			successfulConnections = append(successfulConnections, result)
		}
	}

	return successfulConnections
}
