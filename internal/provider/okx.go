package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"time"
)

type OKXProvider struct {
	Symbol string
}

func (p *OKXProvider) GetName() string {
	return "OKX"
}

func (p *OKXProvider) Start(ctx context.Context, output chan<- model.Tick) {
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

				output <- model.Tick{
					Exchange:  p.GetName(),
					IsCurrent: false,
				}

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

func (p *OKXProvider) connectionAndListen(ctx context.Context, output chan<- model.Tick) error {
	url := "wss://ws.okx.com:8443/ws/v5/public"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	defer conn.Close()

	if err != nil {
		return err
	}

	log.Printf("[%s] Connected to %s", p.GetName(), p.Symbol)

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
		"op": "subscribe",
		"args": []map[string]string{
			{
				"channel": "tickers",
				"instId":  p.Symbol,
			},
		},
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("okx subscribe error: %w", err)
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if string(message) == "{\"event\":\"subscribe\",\"arg\":{\"channel\":\"tickers\",\"instId\":\""+p.Symbol+"\"}}" {
			continue
		}

		var resp struct {
			Data []struct {
				InstId string `json:"instId"`
				BidPx  string `json:"bidPx"`
				AskPx  string `json:"askPx"`
			} `json:"data"`
		}

		if err := json.Unmarshal(message, &resp); err != nil {
			continue
		}

		if len(resp.Data) > 0 {
			d := resp.Data[0]
			bid, _ := strconv.ParseFloat(d.BidPx, 64)
			ask, _ := strconv.ParseFloat(d.AskPx, 64)

			output <- model.Tick{
				Exchange:  p.GetName(),
				Symbol:    d.InstId,
				BestBid:   int64(bid * model.PricePrecision),
				BestAsk:   int64(ask * model.PricePrecision),
				Timestamp: time.Now(),
				IsCurrent: true,
			}
		}
	}
}
