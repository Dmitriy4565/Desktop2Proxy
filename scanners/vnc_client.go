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

	switch runtime.GOOS {
	case "windows":
		// Windows версия - TigerVNC или встроенный
		if CommandExists("vncviewer") {
			cmd = exec.Command("vncviewer",
				fmt.Sprintf("%s:%d", target.IP, port),
				"-password", target.Password)
		} else {
			return fmt.Errorf("❌ Установите TigerVNC или используйте RDP")
		}

	case "linux":
		// LINUX ВЕРСИЯ - используем только TigerVNC
		if CommandExists("vncviewer") {
			fmt.Println("💡 Используем TigerVNC viewer...")

			if target.Password != "" {
				// Создаем временный файл с паролем для безопасности
				passFile := "/tmp/vncpasswd_tiger"
				if err := os.WriteFile(passFile, []byte(target.Password), 0600); err != nil {
					return fmt.Errorf("❌ Ошибка создания файла пароля: %v", err)
				}
				defer os.Remove(passFile)

				// TigerVNC с паролем
				cmd = exec.Command("vncviewer",
					fmt.Sprintf("%s:%d", target.IP, port),
					"-passwd", passFile,
					"-quality", "9", // Качество изображения
					"-compresslevel", "6", // Сжатие
					"-encodings", "tight") // Кодировка
			} else {
				// TigerVNC без пароля
				cmd = exec.Command("vncviewer",
					fmt.Sprintf("%s:%d", target.IP, port),
					"-quality", "9",
					"-compresslevel", "6",
					"-encodings", "tight")
			}

		} else {
			return fmt.Errorf("❌ TigerVNC не найден. Установите: sudo pacman -S tigervnc")
		}

	case "darwin":
		// macOS версия - Screen Sharing или TigerVNC
		if CommandExists("open") {
			// Пробуем встроенный Screen Sharing
			vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
			cmd = exec.Command("open", vncUrl)
		} else if CommandExists("vncviewer") {
			// TigerVNC для macOS
			cmd = exec.Command("vncviewer",
				fmt.Sprintf("%s:%d", target.IP, port),
				"-password", target.Password)
		} else {
			return fmt.Errorf("❌ Используйте Screen Sharing или установите TigerVNC")
		}

	default:
		return fmt.Errorf("❌ Неподдерживаемая ОС: %s", runtime.GOOS)
	}

	fmt.Printf("🚀 Запускаем TigerVNC...\n")
	fmt.Printf("🔗 Адрес: %s:%d\n", target.IP, port)
	if target.Password != "" {
		fmt.Printf("🔑 Пароль: %s\n", "***")
	}

	// Подключаем стандартные потоки для интерактивного режима
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка запуска TigerVNC: %v\n💡 Проверьте подключение и пароль", err)
	}

	fmt.Println("✅ TigerVNC сессия завершена")
	return nil
}

// Простая альтернатива без пароля (для быстрого теста)
func ConnectVNCQuick(target models.Target, port int) error {
	fmt.Printf("👁️ Быстрое подключение к VNC %s:%d...\n", target.IP, port)

	if !CommandExists("vncviewer") {
		return fmt.Errorf("❌ TigerVNC не установлен")
	}

	cmd := exec.Command("vncviewer", fmt.Sprintf("%s:%d", target.IP, port))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("💡 Запускаем TigerVNC (без пароля)...")
	return cmd.Run()
}
