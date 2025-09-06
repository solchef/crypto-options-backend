// package services

// import (
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool { return true },
// }

// type WSMessage struct {
// 	Type      string      `json:"type"`
// 	Data      interface{} `json:"data"`
// 	Timestamp int64       `json:"timestamp"`
// }

// func ServeWS(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println("WebSocket upgrade error:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// Example: hardcode symbol for now
// 	// symbol := "btcusdt"

// 	// // Channels for trade + stats
// 	// tradeCh := make(chan float64)
// 	// statsCh := make(chan map[string]float64)

// 	// // Start Binance subscriptions
// 	// go ListenPriceStream(symbol, tradeCh)
// 	// go ListenTickerStream(symbol, statsCh)

// 	// Start Binance price streams (multiple symbols)
// 	syms := os.Getenv("SYMBOLS")
// 	if syms == "" {
// 		syms = "btcusdt,ethusdt"
// 	}
// 	for _, s := range strings.Split(syms, ",") {
// 		symbol := strings.TrimSpace(s)
// 		if symbol == "" {
// 			continue
// 		}
// 		tradeCh := make(chan float64)
// 		statsCh := make(chan map[string]float64)

// 		go ListenPriceStream(symbol, tradeCh)
// 		go ListenTickerStream(symbol, statsCh)

// 		// Keep last stats
// 		var lastChange, lastChangePct float64

// 		for {
// 			select {
// 			case price := <-tradeCh:
// 				msg := WSMessage{
// 					Type: "price_update",
// 					Data: map[string]interface{}{
// 						"symbol":       symbol,
// 						"price":        price,
// 						"change24h":    lastChange,
// 						"change24hPct": lastChangePct,
// 					},
// 					Timestamp: time.Now().UnixMilli(),
// 				}
// 				if err := conn.WriteJSON(msg); err != nil {
// 					log.Println("write error:", err)
// 					return
// 				}

// 			case stats := <-statsCh:
// 				lastChange = stats["change"]
// 				lastChangePct = stats["changePct"]

// 				// Optionally also push a stats-only update
// 				statsMsg := WSMessage{
// 					Type: "stats_update",
// 					Data: map[string]interface{}{
// 						"symbol":       symbol,
// 						"change24h":    lastChange,
// 						"change24hPct": lastChangePct,
// 					},
// 					Timestamp: time.Now().UnixMilli(),
// 				}
// 				if err := conn.WriteJSON(statsMsg); err != nil {
// 					log.Println("write error:", err)
// 					return
// 				}
// 			}
// 		}
// 	}
// }

package services

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins
}

var wsClients = make(map[*websocket.Conn]bool)
var wsMu sync.Mutex

// Broadcast sends a message to all connected WebSocket clients
func Broadcast(msg WSMessage) {
	wsMu.Lock()
	defer wsMu.Unlock()

	for conn := range wsClients {
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("broadcast error:", err)
			conn.Close()
			delete(wsClients, conn)
		}
	}
}

// ServeWS upgrades HTTP → WebSocket and registers the client
func ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	wsMu.Lock()
	wsClients[conn] = true
	wsMu.Unlock()

	log.Println("New WebSocket client connected")
}
