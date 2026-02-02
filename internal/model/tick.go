package model

import "time"

const PricePrecision = 100_000_000.0

type Tick struct {
	Exchange  string
	Symbol    string
	BestBid   int64
	BestAsk   int64
	Timestamp time.Time
	IsCurrent bool
}
