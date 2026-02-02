package main

import (
	"context"
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/engine"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/bukhuk/arb-scanner/internal/provider"
	"github.com/bukhuk/arb-scanner/internal/ui"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logFile, err := os.OpenFile("scanner.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	log.Println("Arb scanner started")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ticks := make(chan model.Tick)
	arbEngine := engine.NewEngine(0.0001)

	BinanceProvider := &provider.BinanceProvider{Symbol: "btcusdt"}
	BinanceProvider.Start(ctx, ticks)

	ByBitProvider := &provider.ByBitProvider{Symbol: "btcusdt"}
	ByBitProvider.Start(ctx, ticks)

	OKXProvider := &provider.OKXProvider{Symbol: "BTC-USDT"}
	OKXProvider.Start(ctx, ticks)

	monitor := ui.NewMonitor()
	fmt.Print("\033[2J")

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			monitor.Render(arbEngine.GetPrices(), arbEngine.GetOptimal())
		}
	}()

	go func() {
		for {
			select {
			case t := <-ticks:
				arbEngine.ProcessTick(t)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	log.Println("Arb scanner exited")
}
