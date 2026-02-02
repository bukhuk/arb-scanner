package provider

import (
	"encoding/json"
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"strings"
	"time"
)

type BybitProvider struct {
	Symbol string
}

func (p *BybitProvider) GetName() string {
	return "Bybit"
}

func (p *BybitProvider) Start(output chan<- model.Tick) {
	go func() {
		delay := time.Second
		for {
			err := p.connectionAndListen(output)
			if err != nil {
				log.Printf("[%s] Connection lost: %v. Retrying in %v...", p.GetName(), err, delay)
				time.Sleep(delay)
				delay <<= 1
				if delay > time.Minute {
					delay = time.Minute
				}
			}
			delay = time.Second
		}
	}()
}

func (p *BybitProvider) connectionAndListen(output chan<- model.Tick) error {
	url := "wss://stream.bybit.com/v5/public/spot"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	defer conn.Close()

	if err != nil {
		return err
	}

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
