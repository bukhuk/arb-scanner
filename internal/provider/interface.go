package provider

import "github.com/bukhuk/arb-scanner/internal/model"

type Provider interface {
	Start(output chan<- model.Tick) error
	GetName() string
}
