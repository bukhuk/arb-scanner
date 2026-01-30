package provider

import (
	"encoding/json"
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/model"
	"github.com/gorilla/websocket"
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

func (p *BybitProvider) Start(output chan<- model.Tick) error {
	url := "wss://stream.bybit.com/v5/public/spot"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		return fmt.Errorf("bybit dial error: %w", err)
	}

	go p.listen(conn, output)

	return nil
}

func (p *BybitProvider) listen(conn *websocket.Conn, output chan<- model.Tick) {
	defer conn.Close()

	sub := map[string]interface{}{
		"op":   "subscribe",
		"args": []string{"orderbook.1." + strings.ToUpper(p.Symbol)},
	}
	conn.WriteJSON(sub)

	var lastBid, lastAsk int64

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var resp struct {
			Type string `json:"type"` // snapshot или delta
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
