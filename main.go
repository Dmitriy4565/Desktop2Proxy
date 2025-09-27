package main

import (
	"bufio"
	"desktop2proxy/models"
	"desktop2proxy/scanners"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
		"MikroTik":    75,  // MikroTik (между VNC и RDP)
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
		// Используем функцию из пакета scanners вместо локальной
		if err := scanners.ConnectSSH(target, result.Port); err != nil {
			fmt.Printf("❌ Ошибка SSH: %v\n", err)
		}
	case "Telnet":
		// Аналогично для Telnet
		if err := scanners.ConnectTelnet(target, result.Port); err != nil {
			fmt.Printf("❌ Ошибка Telnet: %v\n", err)
		}
	case "HTTP", "HTTPS":
		openBrowserAuto(target, result)
	case "RDP":
		// Разделяем серверы: обычные и двухфакторные
		if target.IP == "198.18.200.225" {
			if err := scanners.ConnectRDPWith2FA(target, result.Port); err != nil {
				fmt.Printf("❌ Ошибка RDP: %v\n", err)
			}
		} else {
			if err := scanners.ConnectRDP(target, result.Port); err != nil {
				fmt.Printf("❌ Ошибка RDP: %v\n", err)
			}
		}
	case "VNC":
		if err := scanners.ConnectVNC(target, result.Port); err != nil {
			fmt.Printf("❌ Ошибка VNC: %v\n", err)
		}
	case "MikroTik":
		if err := scanners.ConnectMikroTik(target, result.Port); err != nil {
			fmt.Printf("❌ Ошибка MikroTik: %v\n", err)
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

// MIKROTIK ПОДКЛЮЧЕНИЕ
func startMikroTikAutoConnect(target models.Target, port int) {
	fmt.Printf("📡 Подключаемся к MikroTik %s:%d...\n", target.IP, port)

	// Пробуем разные методы подключения к MikroTik
	if port == 22 || isSSHPort(port) {
		fmt.Println("💡 Используем SSH подключение...")
		startSSHAutoConnect(target, port)
		return
	}

	if port == 80 || port == 443 || port == 8291 {
		fmt.Println("💡 Открываем веб-интерфейс MikroTik...")
		scheme := "http"
		if port == 443 {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, port)
		openBrowser(url)
		return
	}

	// API подключение
	fmt.Println("💡 API подключение к MikroTik...")
	fmt.Printf("👤 Логин: %s\n", target.Username)
	fmt.Printf("🔑 Пароль: %s\n", "***")
	fmt.Println("📝 Используйте WinBox или утилиты MikroTik для подключения")

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

// Вспомогательная функция для проверки SSH порта
func isSSHPort(port int) bool {
	return port >= 22 && port <= 2222
}

// Вспомогательная функция для открытия браузера по URL
func openBrowser(url string) {
	var cmd *exec.Cmd
	if commandExists("xdg-open") {
		cmd = exec.Command("xdg-open", url)
	} else if commandExists("firefox") {
		cmd = exec.Command("firefox", url)
	} else if commandExists("chromium") {
		cmd = exec.Command("chromium", url)
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

func scanMikroTikSpecific(target models.Target) []models.ProbeResult {
	fmt.Println("🎯 Целевое сканирование MikroTik портов...")

	mikrotikPorts := []int{8728, 8729, 8291, 22, 23, 80, 443, 2000, 20561}
	var results []models.ProbeResult

	for _, port := range mikrotikPorts {
		if isPortOpen(target.IP, port) {
			fmt.Printf("✅ Найден открытый порт MikroTik: %d\n", port)
			results = append(results, models.ProbeResult{
				Protocol: "MikroTik",
				Port:     port,
				Success:  true,
				Banner:   fmt.Sprintf("MikroTik порт %d открыт", port),
			})
		}
	}

	return results
}

// Функция проверки открытого порта
func isPortOpen(ip string, port int) bool {
	timeout := time.Second * 3
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
