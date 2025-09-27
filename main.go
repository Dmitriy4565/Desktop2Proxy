package main

import (
	"bufio"
	"desktop2proxy/models"
	"desktop2proxy/scanners"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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
	fmt.Println("🎯    Desktop2Proxy Linux Auto Connect")
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
	// Приоритет протоколов для Linux
	priority := map[string]int{
		"SSH":         100, // Лучший - нативная консоль
		"Telnet":      90,  // Консольный доступ
		"VNC":         80,  // Графический Linux
		"RDP":         70,  // Графический Windows
		"HTTP":        60,  // Веб-интерфейсы
		"HTTPS":       60,
		"WinRM-HTTP":  50, // Windows management
		"WinRM-HTTPS": 50,
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
	case "Telnet":
		startTelnetAutoConnect(target, result.Port)
	case "HTTP", "HTTPS":
		openBrowserAuto(target, result)
	case "RDP":
		// Разделяем серверы: обычные и двухфакторные
		if target.IP == "198.18.200.225" {
			// Сервер с двухфакторкой
			if err := scanners.ConnectRDPWith2FA(target, result.Port); err != nil {
				fmt.Printf("❌ Ошибка RDP: %v\n", err)
			}
		} else {
			// Обычные серверы
			if err := scanners.ConnectRDP(target, result.Port); err != nil {
				fmt.Printf("❌ Ошибка RDP: %v\n", err)
			}
		}
	case "VNC":
		if err := scanners.ConnectVNC(target, result.Port); err != nil {
			fmt.Printf("❌ Ошибка VNC: %v\n", err)
		}
	case "WinRM-HTTP", "WinRM-HTTPS":
		startWinRMAutoConnect(target, result.Port)
	default:
		fmt.Printf("❌ Автоподключение для %s не реализовано\n", result.Protocol)
		showManualInstructions(target, result)
	}
}

// РЕАЛЬНОЕ SSH ПОДКЛЮЧЕНИЕ
func startSSHAutoConnect(target models.Target, port int) {
	fmt.Printf("🔐 Подключаемся к SSH %s@%s:%d...\n", target.Username, target.IP, port)

	// Проверяем установлен ли SSH
	if !commandExists("ssh") {
		fmt.Println("❌ SSH клиент не установлен. Установите: sudo pacman -S openssh")
		waitForExit()
		return
	}

	// Строим команду SSH
	sshArgs := []string{
		fmt.Sprintf("%s@%s", target.Username, target.IP),
		"-p", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}

	fmt.Println("✅ Запускаем SSH сессию...")
	fmt.Println("💡 Для выхода используйте Ctrl+D или введите 'exit'")

	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Ошибка SSH: %v\n", err)
	}
}

// РЕАЛЬНОЕ TELNET ПОДКЛЮЧЕНИЕ
func startTelnetAutoConnect(target models.Target, port int) {
	fmt.Printf("📟 Подключаемся к Telnet %s:%d...\n", target.IP, port)

	// Для Arch Linux проверяем оба возможных имени
	if !commandExists("telnet") && !commandExists("telnet.netkit") {
		fmt.Println("❌ Telnet клиент не установлен. Установите: sudo pacman -S inetutils")
		waitForExit()
		return
	}

	// Определяем правильное имя команды
	telnetCmd := "telnet"
	if !commandExists("telnet") && commandExists("telnet.netkit") {
		telnetCmd = "telnet.netkit"
	}

	fmt.Println("✅ Запускаем Telnet сессию...")
	fmt.Println("💡 Для выхода используйте Ctrl+] затем введите 'quit'")

	cmd := exec.Command(telnetCmd, target.IP, strconv.Itoa(port))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Ошибка Telnet: %v\n", err)
	}
}

// РЕАЛЬНОЕ ОТКРЫТИЕ БРАУЗЕРА
func openBrowserAuto(target models.Target, result models.ProbeResult) {
	scheme := "http"
	if result.Protocol == "HTTPS" {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, result.Port)

	fmt.Printf("🌐 Открываем браузер: %s\n", url)

	var cmd *exec.Cmd

	// Пробуем разные способы открытия браузера
	if commandExists("xdg-open") {
		cmd = exec.Command("xdg-open", url)
	} else if commandExists("firefox") {
		cmd = exec.Command("firefox", url)
	} else if commandExists("chromium") {
		cmd = exec.Command("chromium", url)
	} else if commandExists("google-chrome") {
		cmd = exec.Command("google-chrome", url)
	} else {
		fmt.Printf("❌ Не найден браузер. Откройте вручную: %s\n", url)
		waitForExit()
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Ошибка открытия браузера: %v\n", err)
		fmt.Printf("🔗 Откройте вручную: %s\n", url)
	} else {
		fmt.Println("✅ Браузер запущен")
	}

	waitForExit()
}

// WinRM ПОДКЛЮЧЕНИЕ
func startWinRMAutoConnect(target models.Target, port int) {
	fmt.Printf("🪟 Подключаемся к WinRM %s:%d...\n", target.IP, port)
	fmt.Println("💡 WinRM подключение требует дополнительных настроек")
	fmt.Printf("📝 Используйте: winrs -r:https://%s:%d -u:%s -p:%s\n",
		target.IP, port, target.Username, target.Password)
	waitForExit()
}

// УТИЛИТА: Проверка существования команды
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
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
