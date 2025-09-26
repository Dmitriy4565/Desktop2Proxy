package scanners

import "os/exec"

// CommandExists проверяет существует ли команда в системе
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
