package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/solchef/crypto-options-backend/models"
)

type Candle struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

func GetCurrentPriceByAPI(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var data map[string]string
	json.Unmarshal(body, &data)

	price, _ := strconv.ParseFloat(data["price"], 64)
	return price, nil
}

func GetPriceHistory(symbol string) ([]models.Candle, error) {
	// Binance requires uppercase with no spaces
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	fmt.Print(symbol)
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=5m&limit=24", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Check if Binance returned error JSON
	var errResp map[string]interface{}
	if json.Unmarshal(body, &errResp) == nil {
		if msg, ok := errResp["msg"]; ok {
			fmt.Print(msg)
			return nil, fmt.Errorf("binance error: %v", errResp)
		}
	}

	// Binance returns array of arrays
	var raw [][]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	candles := make([]models.Candle, len(raw))
	for i, k := range raw {
		candles[i] = models.Candle{
			OpenTime:  int64(k[0].(float64)),
			Open:      atof(k[1]),
			High:      atof(k[2]),
			Low:       atof(k[3]),
			Close:     atof(k[4]),
			Volume:    atof(k[5]),
			CloseTime: int64(k[6].(float64)),
		}
	}

	return candles, nil
}

func atof(v interface{}) float64 {
	str, ok := v.(string)
	if !ok {
		return 0
	}
	val, _ := strconv.ParseFloat(str, 64)
	return val
}
