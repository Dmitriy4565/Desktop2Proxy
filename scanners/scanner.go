package scanners

import (
	"context"
	"desktop2proxy/models"
)

// Scanner интерфейс для всех сканеров протоколов
type Scanner interface {
	GetName() string
	GetDefaultPort() int
	CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult
}
