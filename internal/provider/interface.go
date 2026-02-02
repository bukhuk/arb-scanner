package provider

import (
	"context"
	"github.com/bukhuk/arb-scanner/internal/model"
)

type Provider interface {
	Start(ctx context.Context, output chan<- model.Tick) error
	GetName() string
}
