package scanners

import (
	"bufio"
	"desktop2proxy/models"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func ConnectTelnet(target models.Target, port int) error {
	fmt.Printf("📟 Подключаемся к Telnet %s:%d...\n", target.IP, port)

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(target.IP, strconv.Itoa(port)), 10*time.Second)
	if err != nil {
		return fmt.Errorf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Telnet подключение установлено! Для выхода: Ctrl+C")

	// Чтение входящих данных
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Чтение пользовательского ввода
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "exit" || text == "quit" {
			break
		}
		_, err := conn.Write([]byte(text + "\r\n"))
		if err != nil {
			fmt.Printf("❌ Ошибка отправки: %v\n", err)
			break
		}
	}

	return nil
}
