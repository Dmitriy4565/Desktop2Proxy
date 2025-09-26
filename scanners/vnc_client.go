package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func ConnectVNC(target models.Target, port int) error {
	fmt.Printf("👁️ Подключаемся к VNC %s:%d...\n", target.IP, port)

	var cmd *exec.Cmd
	var useRemmina bool

	switch runtime.GOOS {
	case "windows":
		// Windows версия
		if commandExists("vncviewer") {
			cmd = exec.Command("vncviewer",
				fmt.Sprintf("%s:%d", target.IP, port),
				"-password", target.Password)
		} else {
			return fmt.Errorf("❌ Установите VNC клиент (TigerVNC, RealVNC)")
		}

	case "linux":
		// LINUX ВЕРСИЯ - приоритет Remmina, потом Vinagre, потом vncviewer
		if commandExists("remmina") {
			// Remmina поддерживает профили с паролями
			profileContent := fmt.Sprintf(`[remmina]
name=AutoVNC_%s
protocol=VNC
server=%s
port=%d
password=%s
colordepth=32
quality=9
`, target.IP, target.IP, port, target.Password)

			profileFile := "/tmp/remmina_vnc.remmina"
			if err := os.WriteFile(profileFile, []byte(profileContent), 0644); err != nil {
				return fmt.Errorf("❌ Ошибка создания профиля VNC: %v", err)
			}
			defer os.Remove(profileFile)

			cmd = exec.Command("remmina", "-c", profileFile)
			useRemmina = true

		} else if commandExists("vinagre") {
			// Vinagre с поддержкой пароля в URL
			vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
			if target.Password != "" {
				vncUrl = fmt.Sprintf("vnc://:%s@%s:%d", target.Password, target.IP, port)
			}
			cmd = exec.Command("vinagre", vncUrl)

		} else if commandExists("vncviewer") {
			// TigerVNC/RealVNC viewer
			if target.Password != "" {
				// Создаем временный файл с паролем
				passFile := "/tmp/vncpasswd"
				if err := os.WriteFile(passFile, []byte(target.Password), 0600); err != nil {
					return fmt.Errorf("❌ Ошибка создания файла пароля: %v", err)
				}
				defer os.Remove(passFile)

				cmd = exec.Command("vncviewer",
					fmt.Sprintf("%s:%d", target.IP, port),
					"-passwd", passFile)
			} else {
				cmd = exec.Command("vncviewer", fmt.Sprintf("%s:%d", target.IP, port))
			}

		} else if commandExists("xtightvncviewer") {
			// Альтернативный VNC viewer
			cmd = exec.Command("xtightvncviewer", fmt.Sprintf("%s:%d", target.IP, port))

		} else {
			return fmt.Errorf("❌ VNC клиент не найден. Установите: sudo apt install remmina vinagre tigervnc-viewer")
		}

	case "darwin":
		// macOS версия
		if commandExists("open") {
			vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
			cmd = exec.Command("open", vncUrl)
		} else {
			return fmt.Errorf("❌ Используйте Screen Sharing или VNC клиент для macOS")
		}

	default:
		return fmt.Errorf("❌ Неподдерживаемая ОС: %s", runtime.GOOS)
	}

	fmt.Printf("🚀 Запускаем VNC клиент...\n")

	if useRemmina {
		fmt.Println("💡 Используется Remmina (поддержка профилей)")
	} else {
		fmt.Printf("💡 Используется: %s\n", cmd.Path)
	}

	if err := cmd.Start(); err != nil {
		// Пробуем альтернативный клиент
		fmt.Printf("⚠️ Ошибка запуска, пробуем альтернативный клиент...\n")
		return startAlternativeVNC(target, port)
	}

	fmt.Println("✅ VNC клиент запущен. Закройте окно для завершения.")

	if useRemmina {
		// Remmina не блокирует терминал, поэтому просто ждем завершения процесса
		return cmd.Wait()
	}

	// Для других клиентов ждем завершения
	return cmd.Wait()
}

func startAlternativeVNC(target models.Target, port int) error {
	fmt.Println("🔄 Пробуем альтернативные VNC клиенты...")

	// Простая попытка через remmina без профиля
	if commandExists("remmina") {
		vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
		cmd := exec.Command("remmina", "-c", vncUrl)
		if err := cmd.Start(); err == nil {
			fmt.Println("✅ Remmina запущен (без профиля)")
			return nil
		}
	}

	// Попробуем простой vncviewer
	if commandExists("vncviewer") {
		cmd := exec.Command("vncviewer", fmt.Sprintf("%s:%d", target.IP, port))
		if err := cmd.Start(); err == nil {
			fmt.Println("✅ VNCViewer запущен")
			return nil
		}
	}

	return fmt.Errorf("❌ Не удалось запустить ни один VNC клиент")
}

// Утилита для проверки существования команды
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
