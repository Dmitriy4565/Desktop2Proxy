package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func LaunchRDPConnection(target models.Target, port int) error {
	fmt.Printf("🖥️ Подключаемся к RDP %s:%d...\n", target.IP, port)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows версия (оставляем для совместимости)
		rdpContent := fmt.Sprintf(`
screen mode id:i:2
use multimon:i:0
desktopwidth:i:1024
desktopheight:i:768
session bpp:i:32
winposstr:s:0,1,0,0,800,600
full address:s:%s:%d
username:s:%s
password:s:%s
authentication level:i:2
`, target.IP, port, target.Username, target.Password)

		tmpFile := "auto_connect.rdp"
		if err := os.WriteFile(tmpFile, []byte(rdpContent), 0644); err != nil {
			return fmt.Errorf("❌ Ошибка создания RDP файла: %v", err)
		}
		defer os.Remove(tmpFile)

		cmd = exec.Command("mstsc", tmpFile)

	case "linux":
		// LINUX ВЕРСИЯ - используем Remmina или FreeRDP
		if commandExists("remmina") {
			// Создаем временный профиль Remmina
			profileContent := fmt.Sprintf(`[remmina]
name=AutoRDP_%s
protocol=RDP
server=%s
port=%d
username=%s
password=%s
colordepth=32
resolution=1024x768
`, target.IP, target.IP, port, target.Username, target.Password)

			profileFile := "/tmp/remmina_auto.remmina"
			if err := os.WriteFile(profileFile, []byte(profileContent), 0644); err != nil {
				return fmt.Errorf("❌ Ошибка создания профиля Remmina: %v", err)
			}
			defer os.Remove(profileFile)

			cmd = exec.Command("remmina", "-c", profileFile)

		} else if commandExists("xfreerdp") {
			// Используем FreeRDP
			cmd = exec.Command("xfreerdp",
				"/v:"+target.IP+":"+strconv.Itoa(port),
				"/u:"+target.Username,
				"/p:"+target.Password,
				"/gdi:sw",
				"/compression",
				"/rfx")

		} else if commandExists("rdesktop") {
			// Используем rdesktop (старая версия)
			cmd = exec.Command("rdesktop",
				target.IP+":"+strconv.Itoa(port),
				"-u", target.Username,
				"-p", target.Password,
				"-g", "1024x768")

		} else {
			return fmt.Errorf("❌ RDP клиент не найден. Установите: sudo apt install remmina")
		}

	case "darwin":
		// macOS версия
		return fmt.Errorf("❌ RDP клиент для macOS не настроен. Используйте Microsoft Remote Desktop")

	default:
		return fmt.Errorf("❌ Неподдерживаемая ОС: %s", runtime.GOOS)
	}

	fmt.Println("🚀 Запускаем RDP клиент...")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("❌ Ошибка запуска RDP: %v\n💡 Проверьте установлен ли RDP клиент", err)
	}

	fmt.Println("✅ RDP клиент запущен. Закройте окно для завершения.")
	return cmd.Wait()
}

// Утилита для проверки существования команды
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
