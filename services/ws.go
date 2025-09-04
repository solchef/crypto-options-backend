package services

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

func ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Example: hardcode symbol for now
	symbol := "btcusdt"

	// Channels for trade + stats
	tradeCh := make(chan float64)
	statsCh := make(chan map[string]float64)

	// Start Binance subscriptions
	go ListenPriceStream(symbol, tradeCh)
	// go ListenTickerStream(symbol, statsCh)

	// Keep last stats
	var lastChange, lastChangePct float64

	for {
		select {
		case price := <-tradeCh:
			msg := WSMessage{
				Type: "price_update",
				Data: map[string]interface{}{
					"symbol":       symbol,
					"price":        price,
					"change24h":    lastChange,
					"change24hPct": lastChangePct,
				},
				Timestamp: time.Now().UnixMilli(),
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("write error:", err)
				return
			}

		case stats := <-statsCh:
			lastChange = stats["change"]
			lastChangePct = stats["changePct"]

			// Optionally also push a stats-only update
			statsMsg := WSMessage{
				Type: "stats_update",
				Data: map[string]interface{}{
					"symbol":       symbol,
					"change24h":    lastChange,
					"change24hPct": lastChangePct,
				},
				Timestamp: time.Now().UnixMilli(),
			}
			if err := conn.WriteJSON(statsMsg); err != nil {
				log.Println("write error:", err)
				return
			}
		}
	}
}
