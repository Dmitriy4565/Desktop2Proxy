package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
	"net"
	"strconv"
	"time"
)

// BaseScanner содержит общую логику для всех сканеров
type BaseScanner struct {
	Name        string
	DefaultPort int
}

// CheckTCPConnection общая функция проверки TCP соединения
func (s *BaseScanner) CheckTCPConnection(target models.Target, port int, timeout time.Duration) (net.Conn, error) {
	address := net.JoinHostPort(target.IP, strconv.Itoa(port))
	return net.DialTimeout("tcp", address, timeout)
}

// ReadBanner читает баннер из соединения
func (s *BaseScanner) ReadBanner(conn net.Conn, timeout time.Duration) (string, error) {
	banner := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.Read(banner)
	if err != nil {
		return "", err
	}
	return string(banner[:n]), nil
}

// CommonProbeResult создает стандартный результат проверки
func CommonProbeResult(name string, port int, success bool, errMsg, banner string) models.ProbeResult {
	if success {
		return models.ProbeResult{
			Protocol: name,
			Port:     port,
			Success:  true,
			Banner:   banner,
		}
	}
	return models.ProbeResult{
		Protocol: name,
		Port:     port,
		Success:  false,
		Error:    errMsg,
	}
}

// WithTimeout выполняет функцию с таймаутом
func WithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}
