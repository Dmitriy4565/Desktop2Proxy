package scanners

import (
	"context"
	"desktop2proxy/models"
	"fmt"
	"net/http"
	"time"
)

type HTTPScanner struct {
	Protocol string // "HTTP" или "HTTPS"
}

func (s *HTTPScanner) GetName() string {
	return s.Protocol
}

func (s *HTTPScanner) GetDefaultPort() int {
	if s.Protocol == "HTTPS" {
		return 443
	}
	return 80
}

func (s *HTTPScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	scheme := "http"
	if s.Protocol == "HTTPS" {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, target.IP, port)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("Request creation failed: %v", err),
		}
	}

	if target.Username != "" || target.Password != "" {
		req.SetBasicAuth(target.Username, target.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("HTTP request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		banner := fmt.Sprintf("Status: %s, Server: %s", resp.Status, resp.Header.Get("Server"))
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  true,
			Banner:   banner,
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  false,
		Error:    fmt.Sprintf("HTTP error: %s", resp.Status),
	}
}
