package ui

import (
	"fmt"
	"github.com/bukhuk/arb-scanner/internal/engine"
	"github.com/bukhuk/arb-scanner/internal/model"
	"sort"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
	ColorYellow = "\033[33m"
	ColorGray   = "\033[90m"
)

type Monitor struct {
	StartTime time.Time
}

func NewMonitor() *Monitor {
	return &Monitor{StartTime: time.Now()}
}

func (m *Monitor) Render(prices map[string]model.Tick, optimal engine.Spread) {
	fmt.Print("\033[H")

	fmt.Printf("%s========= ARB-SCANNER ACTIVE | Uptime: %v %s=========\n",
		ColorCyan, time.Since(m.StartTime).Round(time.Second), ColorReset)

	fmt.Println(ColorGray + "--------------------------------------------------------------------------------" + ColorReset)
	fmt.Printf("%-15s | %-18s | %-18s | %-10s\n", "Exchange", "Bid (Sell)", "Ask (Buy)", "Status")
	fmt.Println(ColorGray + "--------------------------------------------------------------------------------" + ColorReset)

	names := make([]string, 0, len(prices))

	for name, _ := range prices {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		tick := prices[name]
		status := ColorGreen + "LIVE" + ColorReset
		if time.Since(tick.Timestamp) > 5*time.Second {
			status = ColorRed + "STALE" + ColorReset
		}

		fmt.Printf("%-15s | %-18.2f | %-18.2f | %-10s\033[K\n",
			name,
			float64(tick.BestBid)/1e8,
			float64(tick.BestAsk)/1e8,
			status,
		)
	}

	fmt.Println(ColorGray + "--------------------------------------------------------------------------------" + ColorReset)

	if optimal.Profit > 0 && time.Since(optimal.Timestamp) < 2*time.Second {
		fmt.Printf("%s[!!!] SIGNAL: Buy %s -> Sell %s%s\033[K\n",
			ColorYellow, optimal.Buyer, optimal.Seller, ColorReset)

		fmt.Printf("%sPROFIT: %.4f%% %s(Prices: %.2f / %.2f) %sAge: %dms%s\033[K\n",
			ColorGreen, optimal.Profit*100,
			ColorGray, optimal.BuyPrice, optimal.SellPrice,
			ColorCyan, time.Since(optimal.Timestamp).Milliseconds(), ColorReset)
	} else {
		fmt.Printf("%sSearching for profitable opportunities...%s\033[K\n", ColorGray, ColorReset)
		fmt.Println("\033[K")
	}
	fmt.Println(ColorGray + "--------------------------------------------------------------------------------" + ColorReset)
}
