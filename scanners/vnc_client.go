package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func ConnectVNC(target models.Target, port int) error {
	fmt.Printf("👁️ Подключаемся к VNC %s:%d...\n", target.IP, port)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows версия - TigerVNC
		if CommandExists("vncviewer") {
			// Простой вызов без лишних опций
			cmd = exec.Command("vncviewer",
				fmt.Sprintf("%s:%d", target.IP, port))
		} else {
			return fmt.Errorf("❌ Установите TigerVNC или используйте RDP")
		}

	case "linux":
		// LINUX ВЕРСИЯ - TigerVNC с правильными опциями
		if CommandExists("vncviewer") {
			fmt.Println("💡 Используем TigerVNC viewer...")

			// Базовые аргументы TigerVNC
			args := []string{
				target.IP + ":" + strconv.Itoa(port),
			}

			// Добавляем опции которые поддерживаются
			args = append(args,
				"-PreferredEncoding", "Tight", // Кодировка
				"-CompressLevel", "6", // Сжатие
				"-QualityLevel", "9", // Качество (правильное имя опции)
			)

			cmd = exec.Command("vncviewer", args...)

		} else if CommandExists("vinagre") {
			// Альтернатива - Vinagre
			fmt.Println("💡 Используем Vinagre...")
			cmd = exec.Command("vinagre",
				"vnc://"+target.IP+":"+strconv.Itoa(port))

		} else if CommandExists("remmina") {
			// Remmina для VNC
			fmt.Println("💡 Используем Remmina...")
			cmd = exec.Command("remmina",
				"-c", "vnc://"+target.IP+":"+strconv.Itoa(port))

		} else {
			return fmt.Errorf("❌ VNC клиент не найден. Установите: sudo pacman -S tigervnc")
		}

	case "darwin":
		// macOS версия
		if CommandExists("open") {
			// Встроенный Screen Sharing
			vncUrl := fmt.Sprintf("vnc://%s:%d", target.IP, port)
			cmd = exec.Command("open", vncUrl)
		} else if CommandExists("vncviewer") {
			// TigerVNC для macOS
			cmd = exec.Command("vncviewer",
				fmt.Sprintf("%s:%d", target.IP, port))
		} else {
			return fmt.Errorf("❌ Используйте Screen Sharing или установите TigerVNC")
		}

	default:
		return fmt.Errorf("❌ Неподдерживаемая ОС: %s", runtime.GOOS)
	}

	fmt.Printf("🚀 Запускаем VNC клиент...\n")
	fmt.Printf("🔗 Адрес: %s:%d\n", target.IP, port)
	if target.Password != "" {
		fmt.Printf("🔑 Пароль будет запрошен VNC клиентом\n")
	}

	// Подключаем стандартные потоки
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка VNC: %v", err)
	}

	fmt.Println("✅ VNC сессия завершена")
	return nil
}

// Упрощенная версия без сложных опций
func ConnectVNCQuick(target models.Target, port int) error {
	fmt.Printf("👁️ Быстрое подключение к VNC %s:%d...\n", target.IP, port)

	if !CommandExists("vncviewer") {
		return fmt.Errorf("❌ TigerVNC не установлен")
	}

	// Самый простой вызов - только адрес
	cmd := exec.Command("vncviewer", target.IP+":"+strconv.Itoa(port))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("💡 Запускаем TigerVNC...")
	return cmd.Run()
}
