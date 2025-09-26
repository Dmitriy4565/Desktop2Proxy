package scanners

import (
	"context"
	"desktop2proxy/models"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHScanner struct{}

func (s *SSHScanner) GetName() string {
	return "SSH"
}

func (s *SSHScanner) GetDefaultPort() int {
	return 22
}

func (s *SSHScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	config := &ssh.ClientConfig{
		User: target.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(target.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	address := net.JoinHostPort(target.IP, strconv.Itoa(port))

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer session.Close()

	err = session.Run("echo 'connection test'")
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    err.Error(),
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   "SSH connection successful",
	}
}
