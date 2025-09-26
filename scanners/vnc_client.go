package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os/exec"
	"runtime"
)

func ConnectVNC(target models.Target, port int) error {
	fmt.Printf("👁️ Подключаемся к VNC %s:%d...\n", target.IP, port)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// TigerVNC или RealVNC
		cmd = exec.Command("vncviewer",
			fmt.Sprintf("%s:%d", target.IP, port),
			"-password", target.Password)

	case "darwin", "linux":
		// Используем vinagre, remmina или vncviewer
		vncUrl := fmt.Sprintf("vnc://%s@%s:%d", target.Username, target.IP, port)
		if target.Password != "" {
			vncUrl = fmt.Sprintf("vnc://%s:%s@%s:%d", target.Username, target.Password, target.IP, port)
		}

		cmd = exec.Command("vinagre", vncUrl)

	default:
		return fmt.Errorf("VNC не поддерживается на %s", runtime.GOOS)
	}

	fmt.Printf("🚀 Запускаем VNC клиент...\n")
	if err := cmd.Start(); err != nil {
		// Попробуем альтернативный клиент
		fmt.Printf("⚠️ Первый клиент не найден, пробуем альтернативный...\n")
		return startAlternativeVNC(target, port)
	}

	fmt.Println("✅ VNC клиент запущен. Закройте окно для завершения.")
	return cmd.Wait()
}

func startAlternativeVNC(target models.Target, port int) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "linux" {
		cmd = exec.Command("remmina", "-c", fmt.Sprintf("vnc://%s:%d", target.IP, port))
	} else {
		return fmt.Errorf("Установите VNC клиент (TigerVNC, RealVNC, Vinagre)")
	}

	return cmd.Start()
}
