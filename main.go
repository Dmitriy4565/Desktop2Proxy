package main

import (
	"bufio"
	"desktop2proxy/models"
	"desktop2proxy/scanners"
	"fmt"
	"os"
	"strings"
)

func main() {
	targetIP := getInput("Введите IP адрес для сканирования: ")
	username := getInput("Введите имя пользователя: ")
	password := getInput("Введите пароль: ")

	target := models.Target{
		IP:       strings.TrimSpace(targetIP),
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	if target.IP == "" {
		fmt.Println("❌ IP адрес не может быть пустым!")
		return
	}

	fmt.Printf("\n🔍 Сканируем хост %s...\n", target.IP)
	if target.Username != "" {
		fmt.Printf("👤 Используются credentials: %s/%s\n", target.Username, "***")
	}
	fmt.Println()

	manager := scanners.NewScannerManager()
	allScanners := manager.GetAllScanners()

	fmt.Println("🔄 Начинаем сканирование...")

	results := scanners.ProbeAllProtocols(target, allScanners)

	if len(results) > 0 {
		fmt.Printf("\n🎯 Успешные подключения: %d\n\n", len(results))
		for i, result := range results {
			fmt.Printf("%d. Протокол: %s\n", i+1, result.Protocol)
			fmt.Printf("   Порт: %d\n", result.Port)
			fmt.Printf("   Статус: Успешная аутентификация\n")
			if result.Banner != "" {
				fmt.Printf("   Информация: %s\n", result.Banner)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("\n❌ Не удалось подключиться ни по одному протоколу")
	}

	fmt.Print("Нажмите Enter для выхода...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func getInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
