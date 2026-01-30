package engine

import (
	"github.com/bukhuk/arb-scanner/internal/model"
	"log"
	"sync"
)

type Engine struct {
	mu     sync.RWMutex
	prices map[string]model.Tick
	fee    float64
}

func NewEngine(fee float64) *Engine {
	return &Engine{
		prices: make(map[string]model.Tick),
		fee:    fee,
	}
}

func (e *Engine) ProcessTick(tick model.Tick) {
	e.mu.Lock()
	e.prices[tick.Exchange] = tick
	e.mu.Unlock()

	e.checkArbitrage(tick)
}

func (e *Engine) checkArbitrage(newTick model.Tick) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for exchange, lastTick := range e.prices {
		if exchange == newTick.Exchange {
			continue
		}

		if profit := e.calcSpread(newTick, lastTick); profit > 0 {
			log.Printf("[PROFIT] Buy at %s (%f), Sell at %s (%f) | Net Profit: %f%%",
				newTick.Exchange, float64(newTick.BestAsk)/100_000_000,
				lastTick.Exchange, float64(lastTick.BestBid)/100_000_000,
				profit*100)
		}

		if profit := e.calcSpread(lastTick, newTick); profit > 0 {
			log.Printf("[PROFIT] Buy at %s (%f), Sell at %s (%f) | Net Profit: %f%%",
				lastTick.Exchange, float64(lastTick.BestAsk)/100_000_000,
				newTick.Exchange, float64(newTick.BestBid)/100_000_000,
				profit*100)
		}
	}
}

func (e *Engine) calcSpread(buyer model.Tick, seller model.Tick) float64 {
	return (float64(seller.BestBid)*(1-e.fee))/(float64(buyer.BestAsk)*(1+e.fee)) - 1
}
