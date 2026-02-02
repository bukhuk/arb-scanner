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

type BinanceProvider struct {
	Symbol string
}

func (p *BinanceProvider) GetName() string {
	return "Binance"
}

func (p *BinanceProvider) Start(output chan<- model.Tick) {
	go func() {
		delay := time.Second
		for {
			log.Printf("[%s] Connection to %s...", p.GetName(), p.Symbol)
			err := p.connectAndListen(output)
			if err != nil {
				log.Printf("[%s] Connection lost: %v. Retrying in %v...", p.GetName(), err, delay)
				time.Sleep(delay)
				delay <<= 1
				if delay > time.Minute {
					delay = time.Minute
				}
				continue
			}
			delay = time.Second
		}
	}()
}

func (p *BinanceProvider) connectAndListen(output chan<- model.Tick) error {
	symbol := strings.ToLower(p.Symbol)
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@bookTicker", symbol)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Binance read error: %v", err)
			return err
		}

		var data struct {
			Symbol  string `json:"s"`
			BestBid string `json:"b"`
			BidQty  string `json:"B"`
			BestAsk string `json:"a"`
			AskQty  string `json:"A"`
		}

		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		bid, _ := strconv.ParseFloat(data.BestBid, 64)
		ask, _ := strconv.ParseFloat(data.BestAsk, 64)

		output <- model.Tick{
			Exchange:  p.GetName(),
			Symbol:    data.Symbol,
			BestBid:   int64(bid * 100_000_000),
			BestAsk:   int64(ask * 100_000_000),
			Timestamp: time.Now(),
		}
	}
}
