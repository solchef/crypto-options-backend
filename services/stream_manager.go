package services

import (
	"log"
	"os"
	"strings"
	"time"
)

// StartMarketStreams sets up Binance trade + ticker streams for all symbols
func StartMarketStreams() {
	syms := os.Getenv("SYMBOLS")
	if syms == "" {
		syms = "btcusdt,ethusdt"
	}
	for _, s := range strings.Split(syms, ",") {
		symbol := strings.TrimSpace(s)
		if symbol == "" {
			continue
		}

		tradeCh := make(chan float64)
		statsCh := make(chan map[string]float64)

		go ListenPriceStream(symbol, tradeCh)
		go ListenTickerStream(symbol, statsCh)

		// Fan-out goroutines for broadcasting
		go func(sym string, ch <-chan float64) {
			for price := range ch {
				Broadcast(WSMessage{
					Type: "price_update",
					Data: map[string]interface{}{
						"symbol": sym,
						"price":  price,
					},
					Timestamp: time.Now().UnixMilli(),
				})
			}
		}(symbol, tradeCh)

		go func(sym string, ch <-chan map[string]float64) {
			for stats := range ch {
				Broadcast(WSMessage{
					Type: "ticker_update",
					Data: map[string]interface{}{
						"symbol":    sym,
						"change":    stats["change"],
						"changePct": stats["changePct"],
					},
					Timestamp: time.Now().UnixMilli(),
				})
			}
		}(symbol, statsCh)

		log.Println("📡 Started streams for", symbol)
	}
}
