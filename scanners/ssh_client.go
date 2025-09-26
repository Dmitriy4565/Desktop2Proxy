package scanners

import (
	"desktop2proxy/models"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func LaunchSSHSession(target models.Target, port int) error {
	config := &ssh.ClientConfig{
		User: target.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(target.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	fmt.Printf("🔐 Подключаемся к %s@%s:%d...\n", target.Username, target.IP, port)

	client, err := ssh.Dial("tcp", net.JoinHostPort(target.IP, strconv.Itoa(port)), config)
	if err != nil {
		return fmt.Errorf("❌ SSH ошибка: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("❌ Ошибка сессии: %v", err)
	}
	defer session.Close()

	// Настраиваем интерактивный терминал
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, oldState)

	// Получаем размер терминала
	w, h, err := term.GetSize(fd)
	if err != nil {
		return err
	}

	// Настраиваем PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		return err
	}

	// Перенаправляем ввод/вывод
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	fmt.Println("✅ SSH сессия установлена! Команды выполняются на удаленном устройстве.")
	fmt.Println("💡 Для выхода введите 'exit' или нажмите Ctrl+D")

	// Запускаем shell
	if err := session.Shell(); err != nil {
		return err
	}

	// Ждем завершения
	if err := session.Wait(); err != nil {
		if err != io.EOF {
			return err
		}
	}

	return nil
}
