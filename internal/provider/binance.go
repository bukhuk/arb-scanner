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

func (p *BinanceProvider) Start(output chan<- model.Tick) error {
	symbol := strings.ToLower(p.Symbol)
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@bookTicker", symbol)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		return fmt.Errorf("binance dial error: %w", err)
	}

	go p.listen(conn, output)

	return nil
}

func (p *BinanceProvider) listen(conn *websocket.Conn, output chan<- model.Tick) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Binance read error: %v", err)
			return
		}

		var data struct {
			Symbol  string `json:"s"`
			BestBid string `json:"b"`
			BestAsk string `json:"a"`
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
