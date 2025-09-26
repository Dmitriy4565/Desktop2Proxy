package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os/exec"
	"runtime"
)

func OpenBrowser(target models.Target, result models.ProbeResult) {
	scheme := "http"
	if result.Protocol == "HTTPS" {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, result.Port)

	fmt.Printf("🌐 Открываем: %s\n", url)
	if target.Username != "" {
		fmt.Printf("🔑 Аутентификация: %s/***\n", target.Username)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Не удалось открыть браузер: %v\n", err)
		fmt.Printf("🔗 Откройте вручную: %s\n", url)
	}
}
