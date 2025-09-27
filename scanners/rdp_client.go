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
			return fmt.Errorf("❌ RDP клиент не найден. Установите: sudo pacman -S freerdp или sudo pacman -S rdesktop")
		}

		fmt.Printf("💡 Используем %s...\n", rdpClient)

		// Аргументы в зависимости от клиента
		if strings.Contains(rdpClient, "freerdp") {
			// FreeRDP аргументы
			args := []string{
				"/v:" + target.IP + ":" + strconv.Itoa(port),
				"/u:" + target.Username,
				"/p:" + target.Password,
				"/cert-ignore",
				"+compression",
				"/gfx-h264",
				"/dynamic-resolution",
			}

			if strings.Contains(rdpClient, "3") {
				args = append(args, "/gfx:RFX")
			}

			cmd = exec.Command(rdpClient, args...)

		} else if strings.Contains(rdpClient, "rdesktop") {
			// rdesktop аргументы
			cmd = exec.Command(rdpClient,
				target.IP+":"+strconv.Itoa(port),
				"-u", target.Username,
				"-p", target.Password,
				"-g", "1024x768",
				"-a", "16",
				"-k", "en-us",
				"-z",      // Сжатие
				"-x", "l", // LAN качество (лучше чем 'm')
				"-P", // Кэширование битмапов
				"-D", // Без decorations
				"-N", // Синхронизация NumLock
				"-C") // Использовать private colormap
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
	fmt.Printf("🔑 Логин: %s, Пароль: %s\n", target.Username, "***")

	// Подключаем стандартные потоки
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка запуска RDP: %v\n💡 Проверьте подключение и учетные данные", err)
	}

	fmt.Println("✅ RDP сессия завершена")
	return nil
}

// Функция для поиска доступного RDP клиента
func findRDPClient() string {
	// Список возможных RDP клиентов (приоритет по порядку)
	rdpClients := []string{
		// Ставим rdesktop на ПЕРВОЕ место - он точно работает!
		"rdesktop",

		// Потом уже пробуем FreeRDP варианты
		"xfreerdp", "freerdp", "wlfreerdp",
		"xfreerdp3", "wlfreerdp3", "freerdp3",
		"/usr/bin/xfreerdp3", "/usr/bin/wlfreerdp3", "/usr/bin/freerdp3",
	}

	for _, client := range rdpClients {
		if CommandExists(client) {
			fmt.Printf("✅ Найден рабочий RDP клиент: %s\n", client)
			return client
		}
	}
	return ""
}
