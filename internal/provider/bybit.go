package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"strings"
	"time"
)

type ByBitProvider struct {
	Symbol string
}

func (p *ByBitProvider) GetName() string {
	return "Bybit"
}

func (p *ByBitProvider) Start(ctx context.Context, output chan<- model.Tick) {
	go func() {
		delay := time.Second
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			log.Printf("[%s] Connection to %s...", p.GetName(), p.Symbol)
			err := p.connectionAndListen(ctx, output)
			if err != nil {
				log.Printf("[%s] Connection lost: %v. Retrying in %v...", p.GetName(), err, delay)
				select {
				case <-time.After(delay):
					delay <<= 1
					if delay > time.Minute {
						delay = time.Minute
					}
				case <-ctx.Done():
					return
				}
				continue
			}
			delay = time.Second
		}
	}()
}

func (p *ByBitProvider) connectionAndListen(ctx context.Context, output chan<- model.Tick) error {
	url := "wss://stream.bybit.com/v5/public/spot"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	defer conn.Close()

	if err != nil {
		return err
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
			return
		}
	}()

	subscribeMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": []string{"orderbook.1." + strings.ToUpper(p.Symbol)},
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("bybit subscribe error: %w", err)
	}

	var lastBid, lastAsk int64

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		var resp struct {
			Type string `json:"type"`
			Data struct {
				Bids [][]string `json:"b"`
				Asks [][]string `json:"a"`
			} `json:"data"`
		}

		if err := json.Unmarshal(message, &resp); err != nil {
			continue
		}

		if len(resp.Data.Bids) > 0 {
			val, _ := strconv.ParseFloat(resp.Data.Bids[0][0], 64)
			lastBid = int64(val * 100_000_000)
		}

		if len(resp.Data.Asks) > 0 {
			val, _ := strconv.ParseFloat(resp.Data.Asks[0][0], 64)
			lastAsk = int64(val * 100_000_000)
		}

		if lastBid > 0 && lastAsk > 0 {
			output <- model.Tick{
				Exchange:  p.GetName(),
				Symbol:    strings.ToUpper(p.Symbol),
				BestBid:   lastBid,
				BestAsk:   lastAsk,
				Timestamp: time.Now(),
			}
		}
	}
}
