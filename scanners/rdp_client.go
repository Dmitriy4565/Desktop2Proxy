package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func LaunchRDPConnection(target models.Target, port int) error {
	fmt.Printf("🖥️ Подключаемся к RDP %s:%d...\n", target.IP, port)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Создаем RDP файл с credentials
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

		// Сохраняем во временный файл
		tmpFile := "auto_connect.rdp"
		if err := os.WriteFile(tmpFile, []byte(rdpContent), 0644); err != nil {
			return fmt.Errorf("❌ Ошибка создания RDP файла: %v", err)
		}
		defer os.Remove(tmpFile)

		cmd = exec.Command("mstsc", tmpFile)
		fmt.Printf("🔑 Используем логин: %s, пароль: ***\n", target.Username)

	case "darwin", "linux":
		return fmt.Errorf("❌ RDP клиент для %s не настроен", runtime.GOOS)
	}

	fmt.Println("🚀 Запускаем Remote Desktop...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка запуска RDP: %v", err)
	}

	return nil
}
