package main

import (
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/engine"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/bukhuk/arb-scanner/internal/provider"
	"github.com/bukhuk/arb-scanner/internal/ui"
	"time"
)

func main() {
	ticks := make(chan model.Tick)

	arbEngine := engine.NewEngine(0.0001)

	BinanceProvider := &provider.BinanceProvider{Symbol: "btcusdt"}
	BinanceProvider.Start(ticks)

	ByBitProvider := &provider.ByBitProvider{Symbol: "btcusdt"}
	ByBitProvider.Start(ticks)

	OKXProvider := &provider.OKXProvider{Symbol: "BTC-USDT"}
	OKXProvider.Start(ticks)

	monitor := ui.NewMonitor()
	fmt.Print("\033[2J")

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			monitor.Render(arbEngine.GetPrices(), arbEngine.GetOptimal())
		}
	}()

	for t := range ticks {
		arbEngine.ProcessTick(t)
	}
}
