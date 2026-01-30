package main

import (
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/bukhuk/arb-scanner/internal/provider"
)

func main() {
	ticks := make(chan model.Tick)

	p1 := &provider.BinanceProvider{Symbol: "btcusdt"}
	p1.Start(ticks)

	p2 := &provider.BybitProvider{Symbol: "btcusdt"}
	p2.Start(ticks)

	for t := range ticks {
		fmt.Printf("Tick: %+v\n", t)
	}
}
