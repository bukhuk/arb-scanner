package main

import (
	"github.com/bukhuk/arb-scanner/internal/engine"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/bukhuk/arb-scanner/internal/provider"
)

func main() {
	ticks := make(chan model.Tick)

	arbEngine := engine.NewEngine(0.0001)

	p1 := &provider.BinanceProvider{Symbol: "btcusdt"}
	p1.Start(ticks)

	p2 := &provider.BybitProvider{Symbol: "btcusdt"}
	p2.Start(ticks)

	for t := range ticks {
		arbEngine.ProcessTick(t)
	}
}
