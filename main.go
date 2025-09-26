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
		startRDPAutoConnect(target, result.Port)
	case "VNC":
		startVNCAutoConnect(target, result.Port)
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
		fmt.Println("❌ SSH клиент не установлен. Установите: sudo apt install openssh-client")
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

	// Если пароль пустой, пробуем подключиться без него
	if target.Password == "" {
		sshArgs = append(sshArgs, "-o", "BatchMode=yes")
	}

	fmt.Println("✅ Запускаем SSH сессию...")
	fmt.Println("💡 Для выхода используйте Ctrl+D или введите 'exit'")

	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Ошибка SSH: %v\n", err)
		if target.Password != "" {
			fmt.Println("💡 Попробуйте ввести пароль вручную при запросе")
		}
	}
}

// РЕАЛЬНОЕ TELNET ПОДКЛЮЧЕНИЕ
func startTelnetAutoConnect(target models.Target, port int) {
	fmt.Printf("📟 Подключаемся к Telnet %s:%d...\n", target.IP, port)

	if !commandExists("telnet") {
		fmt.Println("❌ Telnet клиент не установлен. Установите: sudo apt install telnet")
		waitForExit()
		return
	}

	fmt.Println("✅ Запускаем Telnet сессию...")
	fmt.Println("💡 Для выхода используйте Ctrl+] затем введите 'quit'")

	cmd := exec.Command("telnet", target.IP, strconv.Itoa(port))
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
	} else if commandExists("chromium-browser") {
		cmd = exec.Command("chromium-browser", url)
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

// RDP ПОДКЛЮЧЕНИЕ Через Remmina
func startRDPAutoConnect(target models.Target, port int) {
	fmt.Printf("🖥️ Подключаемся к RDP %s:%d...\n", target.IP, port)

	if !commandExists("remmina") {
		fmt.Println("❌ Remmina не установлен. Установите: sudo apt install remmina")
		waitForExit()
		return
	}

	// Создаем временный профиль Remmina
	profileContent := fmt.Sprintf(`[remmina]
name=%s
protocol=RDP
server=%s
port=%d
username=%s
password=%s
`, target.IP, target.IP, port, target.Username, target.Password)

	profileFile := "/tmp/remmina_temp.remmina"
	if err := os.WriteFile(profileFile, []byte(profileContent), 0644); err != nil {
		fmt.Printf("❌ Ошибка создания профиля: %v\n", err)
		waitForExit()
		return
	}
	defer os.Remove(profileFile)

	fmt.Println("✅ Запускаем Remmina...")
	cmd := exec.Command("remmina", "-c", profileFile)
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Ошибка запуска Remmina: %v\n", err)
	} else {
		fmt.Println("✅ RDP подключение установлено")
	}

	waitForExit()
}

// VNC ПОДКЛЮЧЕНИЕ
func startVNCAutoConnect(target models.Target, port int) {
	fmt.Printf("👁️ Подключаемся к VNC %s:%d...\n", target.IP, port)

	// Пробуем разные VNC клиенты
	var cmd *exec.Cmd

	if commandExists("vinagre") {
		vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
		if target.Password != "" {
			vncUrl = fmt.Sprintf("vnc://%s@%s:%d", target.Password, target.IP, port)
		}
		cmd = exec.Command("vinagre", vncUrl)
	} else if commandExists("remmina") {
		vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
		cmd = exec.Command("remmina", "-c", vncUrl)
	} else {
		fmt.Println("❌ VNC клиент не найден. Установите: sudo apt install vinagre")
		waitForExit()
		return
	}

	fmt.Println("✅ Запускаем VNC клиент...")
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Ошибка запуска VNC: %v\n", err)
	} else {
		fmt.Println("✅ VNC подключение установлено")
	}

	waitForExit()
}

// WinRM ПОДКЛЮЧЕНИЕ (через wine или native go)
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
