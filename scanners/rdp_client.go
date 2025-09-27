package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Функция для поиска доступного RDP клиента (оптимизирована для Ubuntu)
func findRDPClient() string {
	// На Ubuntu приоритет: rdesktop для простых, remmina для двухфакторки
	rdpClients := []string{
		"rdesktop", // Основной для простых серверов
		"remmina",  // Для двухфакторки и CredSSP
		"xfreerdp", // Альтернатива
		"freerdp",  // Альтернатива
		"/usr/bin/rdesktop",
		"/usr/bin/remmina",
		"/usr/bin/xfreerdp",
	}

	for _, client := range rdpClients {
		if CommandExists(client) {
			fmt.Printf("✅ Найден RDP клиент: %s\n", client)
			return client
		}
	}
	return ""
}

func ConnectRDP(target models.Target, port int) error {
	fmt.Printf("🖥️ Подключаемся к RDP %s:%d...\n", target.IP, port)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows версия - встроенный mstsc
		rdpContent := fmt.Sprintf(`
screen mode id:i:2
desktopwidth:i:1024
desktopheight:i:768
full address:s:%s:%d
username:s:%s
password:s:%s
`, target.IP, port, target.Username, target.Password)

		tmpFile := "auto_connect.rdp"
		if err := os.WriteFile(tmpFile, []byte(rdpContent), 0644); err != nil {
			return fmt.Errorf("❌ Ошибка создания RDP файла: %v", err)
		}
		defer os.Remove(tmpFile)

		cmd = exec.Command("mstsc", tmpFile)

	case "linux":
		// LINUX ВЕРСИЯ - проверяем все возможные RDP клиенты
		rdpClient := findRDPClient()
		if rdpClient == "" {
			return fmt.Errorf("❌ RDP клиент не найден. Установите: sudo apt install rdesktop remmina remmina-plugin-rdp")
		}

		fmt.Printf("💡 Используем %s...\n", rdpClient)

		// Аргументы в зависимости от клиента
		if strings.Contains(rdpClient, "remmina") {
			// Remmina для двухфакторки и CredSSP
			fmt.Println("🔐 Запускаем Remmina для двухфакторной аутентификации...")
			fmt.Println("💡 Remmina откроет GUI - настройте подключение вручную")
			fmt.Printf("💡 Сервер: %s:%d\n", target.IP, port)
			fmt.Printf("💡 Логин: %s\n", target.Username)
			fmt.Printf("💡 Пароль: %s\n", "***")

			// Запускаем remmina в GUI режиме
			cmd = exec.Command("remmina")

		} else if strings.Contains(rdpClient, "freerdp") {
			// FreeRDP аргументы
			args := []string{
				"/v:" + target.IP + ":" + strconv.Itoa(port),
				"/u:" + target.Username,
				"/p:" + target.Password,
				"/cert-ignore",
				"+compression",
				"/gdi:sw",
			}
			cmd = exec.Command(rdpClient, args...)

		} else {
			// rdesktop для простых серверов (основной вариант)
			fmt.Println("🔐 Запускаем rdesktop...")
			fmt.Println("💡 Если сервер запросит:")
			fmt.Println("   1. Введите 'yes' для принятия сертификата")
			fmt.Println("   2. Введите пароль двухфакторки когда запросит")

			cmd = exec.Command(rdpClient,
				target.IP+":"+strconv.Itoa(port),
				"-u", target.Username,
				"-p", target.Password,
				"-g", "1024x768",
				"-a", "16",
				"-k", "en-us",
				"-z",      // Сжатие
				"-x", "l") // Качество LAN
		}

	case "darwin":
		return fmt.Errorf("❌ Используйте Microsoft Remote Desktop для macOS")

	default:
		return fmt.Errorf("❌ Неподдерживаемая ОС: %s", runtime.GOOS)
	}

	if cmd == nil {
		return fmt.Errorf("❌ Не удалось создать команду RDP")
	}

	fmt.Printf("🚀 Запускаем RDP клиент...\n")

	// Подключаем стандартные потоки
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка RDP: %v", err)
	}

	fmt.Println("✅ RDP сессия завершена")
	return nil
}

// Отдельная функция для серверов с двухфакторкой
func ConnectRDPWith2FA(target models.Target, port int) error {
	fmt.Println("🔐 Подключение с двухфакторной аутентификацией")

	if !CommandExists("remmina") {
		return fmt.Errorf("❌ Для двухфакторки требуется Remmina. Установите: sudo apt install remmina remmina-plugin-rdp")
	}

	fmt.Println("🖥️ Запускаем Remmina GUI...")
	fmt.Println("📝 Настройте подключение вручную:")
	fmt.Printf("   Сервер: %s:%d\n", target.IP, port)
	fmt.Printf("   Логин: %s\n", target.Username)
	fmt.Printf("   Пароль: %s\n", "***")
	fmt.Println("💡 Remmina поддерживает двухфакторку и современную аутентификацию")

	cmd := exec.Command("remmina")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
