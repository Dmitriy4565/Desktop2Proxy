package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
	"net"
	"strconv"
	"time"
)

// TCPScanner общий сканер для TCP протоколов
type TCPScanner struct {
	ProtocolName string
	Port         int
}

func (s *TCPScanner) GetName() string {
	return s.ProtocolName
}

func (s *TCPScanner) GetDefaultPort() int {
	return s.Port
}

func (s *TCPScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("TCP подключение недоступно: %v", err),
		}
	}
	defer conn.Close()

	// Пытаемся прочитать баннер
	banner := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(banner)

	if err == nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   fmt.Sprintf("TCP сервис обнаружен: %s", string(banner[:n])),
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   "TCP порт открыт",
	}
}

// UDPScanner для UDP протоколов
type UDPScanner struct {
	ProtocolName string
	Port         int
}

func (s *UDPScanner) GetName() string {
	return s.ProtocolName
}

func (s *UDPScanner) GetDefaultPort() int {
	return s.Port
}

func (s *UDPScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))

	conn, err := net.DialTimeout("udp", address, 10*time.Second)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("UDP подключение недоступно: %v", err),
		}
	}
	defer conn.Close()

	// Для UDP отправляем тестовый пакет и ждем ответ
	testPacket := []byte("TEST")
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(testPacket)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    "Ошибка отправки UDP пакета",
		}
	}

	// Пытаемся получить ответ
	response := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(response)

	if err == nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   fmt.Sprintf("UDP сервис ответил: %s", string(response[:n])),
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   "UDP порт открыт (нет ответа)",
	}
}
