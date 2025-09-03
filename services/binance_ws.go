// package services

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"strconv"
// 	"sync"

// 	"github.com/gorilla/websocket"
// )

// type PriceUpdate struct {
// 	Price float64
// }

// var priceMap = struct {
// 	sync.RWMutex
// 	data map[string]float64
// }{data: make(map[string]float64)}

// // ListenPriceStream subscribes to Binance WebSocket for a symbol
// func ListenPriceStream(symbol string) {
// 	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@trade", symbol)
// 	c, _, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		log.Println("WebSocket connection error:", err)
// 		return
// 	}
// 	defer c.Close()

// 	for {
// 		_, message, err := c.ReadMessage()
// 		if err != nil {
// 			log.Println("WebSocket read error:", err)
// 			return
// 		}

// 		//  log.Printf("Received message for %s: %s", symbol, message)

// 		var data map[string]interface{}
// 		if err := json.Unmarshal(message, &data); err != nil {
// 			continue
// 		}

// 		p, ok := data["p"].(string)
// 		if !ok {
// 			continue
// 		}

// 		price, err := strconv.ParseFloat(p, 64)
// 		if err != nil {
// 			continue
// 		}

// 		// Cache price
// 		priceMap.Lock()
// 		priceMap.data[symbol] = price
// 		priceMap.Unlock()

// 		// Broadcast price to all connected clients
// 		BroadcastPrice(symbol, price)
// 	}
// }

// // GetCurrentPrice returns the latest cached price
// func GetCurrentPrice(symbol string) float64 {
// 	priceMap.RLock()
// 	defer priceMap.RUnlock()
// 	return priceMap.data[symbol]
// }

package services

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type PriceUpdate struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change,omitempty"`
	ChangePercent float64 `json:"change_percent,omitempty"`
	High          float64 `json:"high,omitempty"`
	Low           float64 `json:"low,omitempty"`
	Volume        float64 `json:"volume,omitempty"`
	LastPrice     float64 `json:"last_price,omitempty"`
}

var priceMap = struct {
	sync.RWMutex
	data map[string]PriceUpdate
}{data: make(map[string]PriceUpdate)}

// ListenPriceStream subscribes to Binance trade price stream
func ListenPriceStream(symbol string, ch chan<- float64) {
	url := "wss://stream.binance.com:9443/ws/" + symbol + "@trade"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("trade dial:", err)
		return
	}
	defer c.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("trade read:", err)
			return
		}
		var data struct {
			Price string `json:"p"`
		}
		if err := json.Unmarshal(msg, &data); err == nil {
			if price, err := strconv.ParseFloat(data.Price, 64); err == nil {
				ch <- price
			}
		}
	}
}

// ListenTickerStream subscribes to Binance 24hr ticker stats
func ListenTickerStream(symbol string, ch chan<- map[string]float64) {
	url := "wss://stream.binance.com:9443/ws/" + symbol + "@ticker"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("stats dial:", err)
		return
	}
	defer c.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("stats read:", err)
			return
		}
		var data struct {
			PriceChange     string `json:"p"`
			PriceChangePerc string `json:"P"`
		}
		if err := json.Unmarshal(msg, &data); err == nil {
			change, _ := strconv.ParseFloat(data.PriceChange, 64)
			changePct, _ := strconv.ParseFloat(data.PriceChangePerc, 64)
			ch <- map[string]float64{
				"change":    change,
				"changePct": changePct,
			}
		}
	}
}

// GetCurrentPrice returns latest cached update
func GetCurrentPrice(symbol string) PriceUpdate {
	priceMap.RLock()
	defer priceMap.RUnlock()
	return priceMap.data[symbol]
}
