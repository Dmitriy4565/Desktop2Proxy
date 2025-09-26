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
	showWelcome()

	// Запрашиваем данные один раз
	target := getTargetInfo()

	// Сканируем и сразу подключаемся
	runScanAndAutoConnect(target)
}

func showWelcome() {
	fmt.Println("🎯 =================================")
	fmt.Println("🎯    Desktop2Proxy Auto Connect")
	fmt.Println("🎯 =================================")
	fmt.Println()
}

func getTargetInfo() models.Target {
	fmt.Print("🎯 Введите IP адрес: ")
	ip := readInput()

	fmt.Print("👤 Введите логин (или Enter для пропуска): ")
	user := readInput()

	fmt.Print("🔑 Введите пароль (или Enter для пропуска): ")
	pass := readInput()

	return models.Target{
		IP:       strings.TrimSpace(ip),
		Username: strings.TrimSpace(user),
		Password: strings.TrimSpace(pass),
	}
}

func readInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func runScanAndAutoConnect(target models.Target) {
	if target.IP == "" {
		fmt.Println("❌ IP адрес не может быть пустым!")
		return
	}

	fmt.Printf("\n🔍 Сканируем хост %s...\n", target.IP)
	if target.Username != "" {
		fmt.Printf("👤 Используются credentials: %s/%s\n", target.Username, "***")
	}

	manager := scanners.NewScannerManager()
	allScanners := manager.GetAllScanners()

	fmt.Println("🔄 Начинаем сканирование...")

	results := scanners.ProbeAllProtocols(target, allScanners)

	if len(results) == 0 {
		fmt.Println("❌ Не найдено рабочих протоколов")
		fmt.Print("Нажмите Enter для выхода...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	// Автоматически выбираем лучший протокол для подключения
	bestProtocol := selectBestProtocol(results)
	fmt.Printf("🎯 Автоматически выбран протокол: %s\n", bestProtocol.Protocol)

	// Немедленно подключаемся
	autoConnectToProtocol(target, bestProtocol)
}

func selectBestProtocol(results []models.ProbeResult) models.ProbeResult {
	// Приоритет протоколов для автоматического подключения
	priority := map[string]int{
		"SSH":         100,
		"WinRM-HTTP":  90,
		"WinRM-HTTPS": 90,
		"Telnet":      80,
		"HTTP":        70,
		"HTTPS":       70,
		"RDP":         60,
		"VNC":         50,
	}

	var bestResult models.ProbeResult
	bestScore := -1

	for _, result := range results {
		score := priority[result.Protocol]
		if score > bestScore {
			bestScore = score
			bestResult = result
		}
	}

	return bestResult
}

func autoConnectToProtocol(target models.Target, result models.ProbeResult) {
	fmt.Printf("\n🚀 Автоподключение к %s://%s:%d...\n", result.Protocol, target.IP, result.Port)
	fmt.Println("⏳ Устанавливаем соединение...")

	switch result.Protocol {
	case "SSH":
		startSSHAutoConnect(target, result.Port)
	case "WinRM-HTTP", "WinRM-HTTPS":
		startWinRMAutoConnect(target, result.Port)
	case "Telnet":
		startTelnetAutoConnect(target, result.Port)
	case "HTTP", "HTTPS":
		openBrowserAuto(target, result)
	case "RDP":
		startRDPAutoConnect(target, result.Port)
	case "VNC":
		startVNCAutoConnect(target, result.Port)
	default:
		fmt.Printf("❌ Автоподключение для %s не реализовано\n", result.Protocol)
		showManualInstructions(target, result)
	}
}

// Заглушки для авто-подключения (реализуем постепенно)
func startSSHAutoConnect(target models.Target, port int) {
	fmt.Println("🔐 Устанавливаем SSH соединение...")
	fmt.Println("💡 SSH автоподключение в разработке")
	fmt.Printf("📝 Команда для ручного подключения: ssh %s@%s -p %d\n",
		target.Username, target.IP, port)
	waitForExit()
}

func startWinRMAutoConnect(target models.Target, port int) {
	fmt.Println("🪟 Подключаемся к Windows через WinRM...")
	fmt.Println("💡 WinRM автоподключение в разработке")
	waitForExit()
}

func startTelnetAutoConnect(target models.Target, port int) {
	fmt.Println("📟 Подключаемся через Telnet...")
	fmt.Println("💡 Telnet автоподключение в разработке")
	fmt.Printf("📝 Команда для ручного подключения: telnet %s %d\n", target.IP, port)
	waitForExit()
}

func openBrowserAuto(target models.Target, result models.ProbeResult) {
	scheme := "http"
	if result.Protocol == "HTTPS" {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, result.Port)
	fmt.Printf("🌐 Открываем браузер: %s\n", url)

	// Попытка автоматического открытия браузера
	if err := openBrowser(url); err != nil {
		fmt.Printf("❌ Не удалось открыть браузер автоматически\n")
		fmt.Printf("🔗 Откройте вручную: %s\n", url)
	}
	waitForExit()
}

func startRDPAutoConnect(target models.Target, port int) {
	fmt.Println("🖥️ Запускаем Remote Desktop...")
	fmt.Printf("🔑 Настройки RDP:\n")
	fmt.Printf("   Адрес: %s:%d\n", target.IP, port)
	fmt.Printf("   Логин: %s\n", target.Username)
	fmt.Printf("   Пароль: %s\n", target.Password)
	fmt.Println("💡 Запустите 'mstsc' и введите данные выше")
	waitForExit()
}

func startVNCAutoConnect(target models.Target, port int) {
	fmt.Println("👁️ Подключаемся через VNC...")
	fmt.Printf("🔑 Настройки VNC:\n")
	fmt.Printf("   Адрес: %s:%d\n", target.IP, port)
	fmt.Printf("   Пароль: %s\n", target.Password)
	waitForExit()
}

func openBrowser(url string) error {
	// Базовая реализация открытия браузера
	return fmt.Errorf("авто-открытие браузера не реализовано")
}

func showManualInstructions(target models.Target, result models.ProbeResult) {
	fmt.Printf("\n📋 Инструкции для ручного подключения:\n")
	fmt.Printf("Протокол: %s\n", result.Protocol)
	fmt.Printf("Адрес: %s:%d\n", target.IP, result.Port)
	if target.Username != "" {
		fmt.Printf("Логин: %s\n", target.Username)
	}
	if target.Password != "" {
		fmt.Printf("Пароль: %s\n", target.Password)
	}
	waitForExit()
}

func waitForExit() {
	fmt.Println("\n⏹️  Нажмите Enter для выхода...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
