package exchange

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsBaseURL               = "wss://fstream.binance.com/stream?streams="
	maxStreamsPerConnection = 150
)

const symbolURL = "https://fapi.binance.com/fapi/v1/exchangeInfo"

type Symbol struct {
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	QuoteAsset string `json:"quoteAsset"`
}

type AggTrade struct {
	Symbol    string
	Quantity  float64
	Direction int // 1 = buy, -1 = sell
	Timestamp time.Time
	Price     float64
}

type aggTradeMessage struct {
	Data struct {
		Symbol string `json:"s"`
		Qty    string `json:"q"`
		Price  string `json:"p"`
		IsBuy  bool   `json:"m"` // true = sell
		Time   int64  `json:"T"`
	} `json:"data"`
}

var tradeHandler func(AggTrade)

func SetTradeHandler(handler func(AggTrade)) {
	tradeHandler = handler
}

func StartStream(symbols []string) {
	chunks := chunkSymbols(symbols, maxStreamsPerConnection)
	for _, group := range chunks {
		go connectToStream(group)
	}
}

func DownloadSymbols() []string {
	resp, err := http.Get(symbolURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var info struct {
		Symbols []Symbol `json:"symbols"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		panic(err)
	}

	var result []string
	for _, s := range info.Symbols {
		if s.Status == "TRADING" && s.QuoteAsset == "USDT" {
			result = append(result, s.Symbol)
		}
	}
	return result
}

func connectToStream(symbols []string) {
	url := wsBaseURL + buildStreamSuffix(symbols)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("WebSocket error: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WS read error: %v", err)
			return
		}

		var parsed aggTradeMessage
		if err := json.Unmarshal(msg, &parsed); err != nil {
			log.Printf("WS unmarshal error: %v", err)
			continue
		}

		qty, err := parseFloat(parsed.Data.Qty)
		if err != nil {
			log.Printf("Qty parse error: %v", err)
			continue
		}

		price, err := parseFloat(parsed.Data.Price)
		if err != nil {
			log.Printf("Price parse error: %v", err)
			continue
		}

		trade := AggTrade{
			Symbol:    parsed.Data.Symbol,
			Quantity:  qty,
			Price:     price,
			Direction: ifThenElse(parsed.Data.IsBuy, -1, 1),
			Timestamp: time.UnixMilli(parsed.Data.Time),
		}

		if tradeHandler != nil {
			tradeHandler(trade)
		}
	}
}

func buildStreamSuffix(symbols []string) string {
	var s []string
	for _, sym := range symbols {
		s = append(s, strings.ToLower(sym)+"@aggTrade")
	}
	return strings.Join(s, "/")
}

func chunkSymbols(symbols []string, size int) [][]string {
	var chunks [][]string
	for i := 0; i < len(symbols); i += size {
		end := i + size
		if end > len(symbols) {
			end = len(symbols)
		}
		chunks = append(chunks, symbols[i:end])
	}
	return chunks
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ifThenElse(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}
