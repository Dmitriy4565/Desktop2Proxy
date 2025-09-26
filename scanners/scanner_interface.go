package scanners

import (
	"context"
	"desktop2proxy/models"
)

type Scanner interface {
	GetName() string
	GetDefaultPort() int
	CheckProtocol(ctx context.Context, target models.Target, port int) models.ProbeResult
}
