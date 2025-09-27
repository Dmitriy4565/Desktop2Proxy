package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Функция проверки существования команды

// Функция для поиска RDP клиента для обычных серверов
func findSimpleRDPClient() string {
	// Для обычных серверов используем rdesktop
	rdpClients := []string{
		"rdesktop",
		"/usr/bin/rdesktop",
	}

	for _, client := range rdpClients {
		if CommandExists(client) {
			fmt.Printf("✅ Найден RDP клиент для обычных серверов: %s\n", client)
			return client
		}
	}
	return ""
}

// Функция для поиска RDP клиента для двухфакторных серверов
func find2FARDPClient() string {
	// Для двухфакторных серверов используем xfreerdp
	rdpClients := []string{
		"xfreerdp",
		"/usr/bin/xfreerdp",
	}

	for _, client := range rdpClients {
		if CommandExists(client) {
			fmt.Printf("✅ Найден RDP клиент для двухфакторки: %s\n", client)
			return client
		}
	}
	return ""
}

// Подключение для обычных серверов (логин/пароль)
func ConnectRDP(target models.Target, port int) error {
	fmt.Printf("🖥️ Подключаемся к RDP %s:%d...\n", target.IP, port)

	rdpClient := findSimpleRDPClient()
	if rdpClient == "" {
		return fmt.Errorf("❌ RDP клиент не найден. Установите: sudo pacman -S rdesktop")
	}

	fmt.Printf("💡 Используем %s для обычного подключения...\n", rdpClient)
	fmt.Println("🔐 Вводите данные по запросу сервера")

	cmd := exec.Command(rdpClient,
		target.IP+":"+strconv.Itoa(port),
		"-u", target.Username,
		"-p", target.Password,
		"-g", "1024x768",
		"-a", "16",
		"-k", "en-us",
		"-z") // Сжатие

	fmt.Printf("🚀 Запускаем...\n")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка RDP: %v", err)
	}

	fmt.Println("✅ RDP сессия завершена")
	return nil
}

// Подключение для серверов с двухфакторкой
func ConnectRDPWith2FA(target models.Target, port int) error {
	fmt.Printf("🖥️ Подключаемся к RDP с двухфакторкой %s:%d...\n", target.IP, port)

	rdpClient := find2FARDPClient()
	if rdpClient == "" {
		return fmt.Errorf("❌ RDP клиент для двухфакторки не найден. Установите: sudo pacman -S freerdp2")
	}

	fmt.Printf("💡 Используем %s для двухфакторной аутентификации...\n", rdpClient)
	fmt.Println("🔐 FreeRDP автоматически обработает CredSSP/NLA")

	// FreeRDP аргументы для CredSSP
	args := []string{
		"/v:" + target.IP + ":" + strconv.Itoa(port),
		"/u:" + target.Username,
		"/p:" + target.Password,
		"/cert-ignore", // Игнорировать сертификаты
		"/sec:nla",     // Network Level Authentication (CredSSP)
		"+compression", // Сжатие
		"/gdi:sw",      // Software rendering
		"/gfx",         // Graphics pipeline
		"/rfx",         // RemoteFX
		"/floatbar",    // Панель управления
	}

	cmd := exec.Command(rdpClient, args...)

	fmt.Printf("🚀 Запускаем...\n")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка RDP с двухфакторкой: %v", err)
	}

	fmt.Println("✅ RDP сессия с двухфакторкой завершена")
	return nil
}
