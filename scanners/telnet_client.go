package scanners

import (
	"bufio"
	"desktop2proxy/models"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func ConnectTelnet(target models.Target, port int) error {
	fmt.Printf("📟 Подключаемся к Telnet %s:%d...\n", target.IP, port)

	// Для Arch Linux проверяем оба возможных имени telnet
	if CommandExists("telnet") || CommandExists("telnet.netkit") {
		return startSystemTelnet(target, port)
	}

	// Запасной вариант: наша Go реализация
	return startGoTelnet(target, port)
}

// Остальной код без изменений...
func startSystemTelnet(target models.Target, port int) error {
	fmt.Println("💡 Используем системный telnet клиент...")

	// Определяем правильное имя команды для Arch Linux
	telnetCmd := "telnet"
	if !CommandExists("telnet") && CommandExists("telnet.netkit") {
		telnetCmd = "telnet.netkit"
	}

	cmd := exec.Command(telnetCmd, target.IP, strconv.Itoa(port))

	// Подключаем стандартные потоки
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("✅ Telnet сессия запущена. Для выхода: Ctrl+] затем 'quit'")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ Ошибка telnet: %v", err)
	}

	return nil
}

// Go реализация telnet (запасной вариант)
func startGoTelnet(target models.Target, port int) error {
	fmt.Println("💡 Используем встроенный telnet клиент...")

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(target.IP, strconv.Itoa(port)), 10*time.Second)
	if err != nil {
		return fmt.Errorf("❌ Ошибка подключения: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Telnet подключение установлено!")
	fmt.Println("💡 Для выхода используйте Ctrl+C или введите 'exit'")

	// Канал для graceful shutdown
	done := make(chan bool)

	// Чтение данных от сервера
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("❌ Ошибка чтения: %v\n", err)
				}
				done <- true
				return
			}
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
		}
	}()

	// Чтение пользовательского ввода и отправка
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()

			// Команды выхода
			if text == "exit" || text == "quit" || text == "logout" {
				conn.Write([]byte("exit\r\n"))
				done <- true
				return
			}

			// Отправляем команду с \r\n
			_, err := conn.Write([]byte(text + "\r\n"))
			if err != nil {
				fmt.Printf("❌ Ошибка отправки: %v\n", err)
				done <- true
				return
			}
		}
	}()

	// Ожидаем завершения
	<-done
	fmt.Println("👋 Telnet сессия завершена")
	return nil
}
