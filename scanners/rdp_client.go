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

// Функция проверки существования команды
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Функция для поиска доступного RDP клиента (оптимизирована для Ubuntu)
func findRDPClient() string {
	// На Ubuntu приоритет: remmina -> freerdp -> rdesktop
	rdpClients := []string{
		"remmina",  // Лучший - поддерживает CredSSP
		"xfreerdp", // FreeRDP (стабильная версия)
		"freerdp",  // Альтернативное имя
		"rdesktop", // Базовый (для простых серверов)
		"/usr/bin/remmina",
		"/usr/bin/xfreerdp",
		"/usr/bin/rdesktop",
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
			return fmt.Errorf("❌ RDP клиент не найден. Установите: sudo apt install remmina remmina-plugin-rdp")
		}

		fmt.Printf("💡 Используем %s...\n", rdpClient)

		// Аргументы в зависимости от клиента
		if strings.Contains(rdpClient, "remmina") {
			// Remmina для Ubuntu - современный клиент
			cmd = exec.Command("remmina",
				"-c", fmt.Sprintf("rdp://%s:%d", target.IP, port),
				"-u", target.Username,
				"-p", target.Password)

		} else if strings.Contains(rdpClient, "freerdp") {
			// FreeRDP аргументы для Ubuntu
			args := []string{
				"/v:" + target.IP + ":" + strconv.Itoa(port),
				"/u:" + target.Username,
				"/p:" + target.Password,
				"/cert-ignore",
				"+compression",
				"/gdi:sw",
			}
			cmd = exec.Command(rdpClient, args...)

		} else if strings.Contains(rdpClient, "rdesktop") {
			fmt.Println("🔐 RDesktop - вводите данные интерактивно")
			fmt.Println("💡 Если запросит: 1. 'yes' для сертификата 2. Пароль двухфакторки")

			cmd = exec.Command(rdpClient,
				target.IP+":"+strconv.Itoa(port),
				"-u", target.Username,
				"-p", target.Password,
				"-g", "1024x768",
				"-a", "16",
				"-k", "en-us")
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
