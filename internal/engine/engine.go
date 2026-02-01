package engine

import (
	"github.com/bukhuk/arb-scanner/internal/model"
	"sync"
	"time"
)

type Spread struct {
	Timestamp time.Time
	Profit    float64
	mu        sync.RWMutex
	Buyer     string
	BuyPrice  float64
	Seller    string
	SellPrice float64
}

type Engine struct {
	mu      sync.RWMutex
	prices  map[string]model.Tick
	fee     float64
	optimal Spread
}

func NewEngine(fee float64) *Engine {
	return &Engine{
		prices: make(map[string]model.Tick),
		fee:    fee,
	}
}

func (e *Engine) ProcessTick(tick model.Tick) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.prices[tick.Exchange] = tick

	e.checkArbitrage(tick)
}

func (e *Engine) GetPrices() map[string]model.Tick {
	e.mu.RLock()
	defer e.mu.RUnlock()
	snapshot := make(map[string]model.Tick, len(e.prices))
	for k, v := range e.prices {
		snapshot[k] = v
	}
	return snapshot
}

func (e *Engine) checkArbitrage(newTick model.Tick) {
	for exchange, lastTick := range e.prices {
		if exchange == newTick.Exchange {
			continue
		}

		if profit := e.calcSpread(newTick, lastTick); profit > 0 {
			e.updateOptimal(newTick, lastTick, profit)
		}

		if profit := e.calcSpread(lastTick, newTick); profit > 0 {
			e.updateOptimal(lastTick, newTick, profit)
		}
	}
}

func (e *Engine) calcSpread(buyer model.Tick, seller model.Tick) float64 {
	return (float64(seller.BestBid)*(1-e.fee))/(float64(buyer.BestAsk)*(1+e.fee)) - 1
}

func (e *Engine) updateOptimal(buyer model.Tick, seller model.Tick, profit float64) {
	var t time.Time
	if buyer.Timestamp.Before(seller.Timestamp) {
		t = seller.Timestamp
	} else {
		t = buyer.Timestamp
	}

	if t.Sub(e.optimal.Timestamp).Milliseconds() > 500 {
		e.optimal.Profit = 0
	}

	if profit > e.optimal.Profit || t.Sub(e.optimal.Timestamp).Milliseconds() > 100 {
		e.optimal.Timestamp = t
		e.optimal.Profit = profit
		e.optimal.Buyer = buyer.Exchange
		e.optimal.BuyPrice = float64(buyer.BestAsk) / 100_000_000
		e.optimal.Seller = seller.Exchange
		e.optimal.SellPrice = float64(seller.BestBid) / 100_000_000
	}
}

func (e *Engine) GetOptimal() Spread {
	e.mu.RLock()
	defer e.mu.RUnlock()
	c := Spread{
		Timestamp: e.optimal.Timestamp,
		Profit:    e.optimal.Profit,
		Buyer:     e.optimal.Buyer,
		BuyPrice:  e.optimal.BuyPrice,
		Seller:    e.optimal.Seller,
		SellPrice: e.optimal.SellPrice,
	}
	return c
}
