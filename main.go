package main

import (
    "context"
    "fmt"
    "log"
    "protocol-scanner/scanner"
)

func main() {
    // Пример использования
    target := Target{
        IP:       "192.168.1.1",
        Username: "admin", 
        Password: "password",
    }

    fmt.Printf("Сканируем хост %s...\n", target.IP)
    
    // Создаем менеджер сканеров
    manager := scanner.NewScannerManager()
    scanners := manager.GetAllScanners()
    
    fmt.Printf("Доступно сканеров: %d\n", len(scanners))
    for _, s := range scanners {
        fmt.Printf(" - %s (порт %d)\n", s.GetName(), s.GetDefaultPort())
    }
    
    result := probeProtocols(target, scanners)
    
    if result != nil {
        fmt.Printf("🎯 УСПЕХ! Протокол: %s, Порт: %d\n", result.Protocol, result.Port)
        if result.Banner != "" {
            fmt.Printf("   Доп. информация: %s\n", result.Banner)
        }
    } else {
        fmt.Println("❌ Ни один протокол не подошел")
    }
}

func probeProtocols(target Target, scanners []scanner.Scanner) *ProbeResult {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    resultChan := make(chan ProbeResult, len(scanners))

    // Запускаем все сканеры параллельно
    for _, scanner := range scanners {
        go func(s scanner.Scanner) {
            port := s.GetDefaultPort()
            result := s.CheckProtocol(ctx, target, port)
            resultChan <- result
        }(scanner)
    }

    // Ждем результаты
    for range scanners {
        result := <-resultChan
        if result.Success {
            cancel() // Останавливаем остальные проверки
            return &result
        } else {
            fmt.Printf("❌ %s:%d - %s\n", result.Protocol, result.Port, result.Error)
        }
    }

    return nil
}