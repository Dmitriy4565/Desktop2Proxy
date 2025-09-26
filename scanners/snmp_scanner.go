package scanners

import (
	"context"
	"desktop2proxy/models" // Исправлен импорт
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

type SNMPScanner struct{}

func (s *SNMPScanner) GetName() string {
	return "SNMP"
}

func (s *SNMPScanner) GetDefaultPort() int {
	return 161
}

func (s *SNMPScanner) CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult { // Исправлены типы
	snmp := &gosnmp.GoSNMP{
		Target:    target.IP,
		Port:      uint16(port),
		Community: target.Password,
		Version:   gosnmp.Version2c,
		Timeout:   5 * time.Second,
	}

	err := snmp.Connect()
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("Connect failed: %v", err),
		}
	}
	defer snmp.Conn.Close()

	oid := ".1.3.6.1.2.1.1.1.0"
	result, err := snmp.Get([]string{oid})
	if err != nil {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("SNMP Get failed: %v", err),
		}
	}

	if result.Error != gosnmp.NoError {
		return models.ProbeResult{
			Protocol: s.GetName(),
			Port:     port,
			Success:  false,
			Error:    fmt.Sprintf("SNMP error: %v", result.Error),
		}
	}

	return models.ProbeResult{
		Protocol: s.GetName(),
		Port:     port,
		Success:  true,
		Banner:   "SNMP accessible",
	}
}
