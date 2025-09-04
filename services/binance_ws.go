package services

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/solchef/crypto-options-backend/config"
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
// func ListenPriceStream(symbol string, ch chan<- float64) {
// 	url := "wss://stream.binance.com:9443/ws/" + symbol + "@trade"
// 	c, _, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		log.Println("trade dial:", err)
// 		return
// 	}
// 	defer c.Close()

// 	for {
// 		_, msg, err := c.ReadMessage()
// 		if err != nil {
// 			log.Println("trade read:", err)
// 			return
// 		}
// 		var data struct {
// 			Price string `json:"p"`
// 		}
// 		if err := json.Unmarshal(msg, &data); err == nil {
// 			if price, err := strconv.ParseFloat(data.Price, 64); err == nil {
// 				ch <- price
// 			}
// 		}
// 	}
// }

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

				// --- 1. Update in-memory cache ---
				priceMap.Lock()
				priceMap.data[symbol] = PriceUpdate{
					Symbol: symbol,
					Price:  price,
				}
				priceMap.Unlock()

				// --- 2. Write-through to Redis ---
				if err := config.Redis.Set(config.Ctx, "price:"+symbol, price, 0).Err(); err != nil {
					log.Println("redis set error:", err)
				}

				// --- 3. Publish to Kafka ---
				if err := PublishPriceTick(symbol, price); err != nil {
					log.Println("kafka publish error:", err)
				}

				// --- 4. Push to WebSocket clients (if enabled) ---
				//BroadcastPrice(symbol, price)

				// Still support legacy channel if caller uses it
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
