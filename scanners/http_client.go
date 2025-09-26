package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os/exec"
	"runtime"
)

func OpenBrowser(target models.Target, result models.ProbeResult) error {
	scheme := "http"
	if result.Protocol == "HTTPS" {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, result.Port)

	fmt.Printf("🌐 Открываем: %s\n", url)
	if target.Username != "" {
		fmt.Printf("🔑 Логин: %s, Пароль: %s\n", target.Username, target.Password)
	}

	var cmd *exec.Cmd
	var browserName string

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
		browserName = "браузер по умолчанию"

	case "darwin":
		cmd = exec.Command("open", url)
		browserName = "Safari/браузер по умолчанию"

	case "linux":
		// LINUX: Пробуем разные способы открытия браузера
		if commandExists("xdg-open") {
			cmd = exec.Command("xdg-open", url)
			browserName = "браузер по умолчанию (xdg-open)"
		} else if commandExists("firefox") {
			cmd = exec.Command("firefox", url)
			browserName = "Firefox"
		} else if commandExists("chromium-browser") {
			cmd = exec.Command("chromium-browser", url)
			browserName = "Chromium"
		} else if commandExists("google-chrome") {
			cmd = exec.Command("google-chrome", url)
			browserName = "Google Chrome"
		} else if commandExists("opera") {
			cmd = exec.Command("opera", url)
			browserName = "Opera"
		} else {
			return fmt.Errorf("❌ Не найден браузер. Установите: sudo apt install firefox")
		}

	default:
		return fmt.Errorf("❌ Неподдерживаемая ОС: %s", runtime.GOOS)
	}

	fmt.Printf("🚀 Запускаем %s...\n", browserName)

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Ошибка открытия браузера: %v\n", err)
		fmt.Printf("🔗 Откройте вручную: %s\n", url)

		// Дополнительная информация для Linux
		if runtime.GOOS == "linux" {
			fmt.Println("💡 На Linux убедитесь что:")
			fmt.Println("   - Браузер установлен (sudo apt install firefox)")
			fmt.Println("   - X11 сервер запущен (для графического режима)")
			fmt.Println("   - Переменная DISPLAY установлена")
		}

		return err
	}

	fmt.Println("✅ Браузер успешно запущен")
	return nil
}

// Автономная версия для открытия любой URL
func OpenURL(url string) error {
	fmt.Printf("🌐 Открываем URL: %s\n", url)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		if commandExists("xdg-open") {
			cmd = exec.Command("xdg-open", url)
		} else if commandExists("firefox") {
			cmd = exec.Command("firefox", url)
		} else {
			return fmt.Errorf("браузер не найден")
		}
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("неподдерживаемая ОС")
	}

	return cmd.Start()
}
