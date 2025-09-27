package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func ConnectSSH(target models.Target, port int) error {
	fmt.Printf("🔐 Подключаемся к SSH %s@%s:%d...\n", target.Username, target.IP, port)

	cmd := exec.Command("ssh",
		fmt.Sprintf("%s@%s", target.Username, target.IP),
		"-p", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=no")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
