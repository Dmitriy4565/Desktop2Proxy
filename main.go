package main

import (
	"context"
	"desktop2proxy/models"
	"desktop2proxy/scanners"
	"fmt"
)

func main() {
	target := models.Target{
		IP:       "192.168.1.1",
		Username: "admin",
		Password: "password",
	}

	fmt.Printf("🔍 Сканируем хост %s...\n\n", target.IP)

	manager := scanners.NewScannerManager()
	allScanners := manager.GetAllScanners()

	fmt.Printf("Доступно протоколов для проверки: %d\n", len(allScanners))
	for _, scanner := range allScanners {
		fmt.Printf(" - %s (порт %d)\n", scanner.GetName(), scanner.GetDefaultPort())
	}
	fmt.Println()

	result := probeProtocols(target, allScanners)

	if result != nil {
		fmt.Printf("🎯 УСПЕХ! Найден рабочий протокол:\n")
		fmt.Printf("   Протокол: %s\n", result.Protocol)
		fmt.Printf("   Порт: %d\n", result.Port)
		if result.Banner != "" {
			fmt.Printf("   Информация: %s\n", result.Banner)
		}
	} else {
		fmt.Println("❌ Ни один протокол не подошел")
	}
}

func probeProtocols(target models.Target, scanners []scanners.Scanner) *models.ProbeResult {
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
