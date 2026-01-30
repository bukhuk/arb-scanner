package main

import (
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/bukhuk/arb-scanner/internal/provider"
)

func main() {
	ticks := make(chan model.Tick)
	p := &provider.BinanceProvider{Symbol: "btcusdt"}
	p.Start(ticks)
	for t := range ticks {
		fmt.Printf("Tick: %+v\n", t)
	}
}
