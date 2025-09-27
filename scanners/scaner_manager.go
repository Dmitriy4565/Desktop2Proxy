package scanners

import (
	"context"
	"desktop2proxy/models"
	"time"
)

// ScannerManager управляет всеми сканерами
type ScannerManager struct {
	scanners []Scanner
}

// NewScannerManager создает менеджер со всеми доступными сканерами
func NewScannerManager() *ScannerManager {
	return &ScannerManager{
		scanners: []Scanner{
			// Веб-протоколы
			&HTTPScanner{Protocol: "HTTP"},
			&HTTPScanner{Protocol: "HTTPS"},

			// Удаленное управление
			&SSHScanner{},
			&WinRMScanner{UseHTTPS: false},
			&WinRMScanner{UseHTTPS: true},
			&RDPScanner{},
			&TelnetScanner{},
			&VNCScanner{},
			&MikroTikScanner{}, // Добавляем MikroTik сканер

			// Сетевые протоколы
			&SNMPScanner{},

			// TCP сканеры для популярных портов
			&TCPScanner{ProtocolName: "FTP", Port: 21},
			&TCPScanner{ProtocolName: "SMTP", Port: 25},
			&TCPScanner{ProtocolName: "DNS", Port: 53}, // Оставляем только TCP DNS
			&TCPScanner{ProtocolName: "HTTP-Alt", Port: 8080},
			&TCPScanner{ProtocolName: "HTTPS-Alt", Port: 8443},

			// Базы данных
			&TCPScanner{ProtocolName: "MySQL", Port: 3306},
			&TCPScanner{ProtocolName: "PostgreSQL", Port: 5432},
			&TCPScanner{ProtocolName: "Redis", Port: 6379},
			&TCPScanner{ProtocolName: "MongoDB", Port: 27017},
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
// ProbeAllProtocols проверяет ВСЕ протоколы
func ProbeAllProtocols(target models.Target, scanners []Scanner) []models.ProbeResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Увеличили до 60 секунд
	defer cancel()

	resultChan := make(chan models.ProbeResult, len(scanners))
	var successfulConnections []models.ProbeResult

	// Запускаем все проверки параллельно с увеличенными таймаутами
	for _, scanner := range scanners {
		go func(s Scanner) {
			// Для каждого сканера свой таймаут
			scannerCtx, cancelScanner := context.WithTimeout(ctx, 15*time.Second) // Увеличили до 15 сек
			defer cancelScanner()

			result := s.CheckProtocol(scannerCtx, target, s.GetDefaultPort())
			resultChan <- result
		}(scanner)
	}

	for i := 0; i < len(scanners); i++ {
		result := <-resultChan
		if result.Success {
			successfulConnections = append(successfulConnections, result)
		}
	}

	return successfulConnections
}
