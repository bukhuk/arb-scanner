package model

import "time"

type Tick struct {
	Exchange  string
	Symbol    string
	BestBid   int64
	BestAsk   int64
	Timestamp time.Time
}
