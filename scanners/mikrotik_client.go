package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func ConnectMikroTik(target models.Target, port int) error {
	fmt.Printf("🔄 Подключаемся к MikroTik %s:%d...\n", target.IP, port)

	// Метод 1: SSH (если доступен)
	if port == 22 || isSSHPortOpen(target.IP, port) {
		fmt.Println("💡 Используем SSH подключение...")
		return connectSSHForMikroTik(target, port)
	}

	// Метод 2: Веб-интерфейс
	fmt.Println("💡 Открываем веб-интерфейс MikroTik...")
	scheme := "http"
	if port == 443 {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, port)
	return openBrowser(url)
}

// SSH подключение для MikroTik
func connectSSHForMikroTik(target models.Target, port int) error {
	fmt.Printf("🔐 SSH подключение к %s:%d...\n", target.IP, port)

	cmd := exec.Command("ssh",
		fmt.Sprintf("%s@%s", target.Username, target.IP),
		"-p", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=no")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка SSH: %v", err)
	}
	return nil
}

// Вспомогательная функция для проверки SSH
func isSSHPortOpen(ip string, port int) bool {
	return port >= 22 && port <= 2222
}

// Вспомогательная функция для открытия браузера
func openBrowser(url string) error {
	fmt.Printf("🌐 Открываем: %s\n", url)

	var cmd *exec.Cmd
	if CommandExists("xdg-open") {
		cmd = exec.Command("xdg-open", url)
	} else if CommandExists("firefox") {
		cmd = exec.Command("firefox", url)
	} else if CommandExists("chromium") {
		cmd = exec.Command("chromium", url)
	} else {
		return fmt.Errorf("❌ Браузер не найден. Откройте вручную: %s", url)
	}

	return cmd.Start()
}
