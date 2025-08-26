package exchange

import (
	"sync"
	"time"

	"cvd-bot/storage"
)

type MinuteCVD struct {
	Timestamp time.Time
	Symbol    string
	Value     float64
}

var (
	cvdPerSymbol = make(map[string]float64)
	mutex        sync.Mutex

	minuteAgg = make(map[string]float64)
	lastFlush = make(map[string]time.Time)
)

func StartAggregator() {
	SetTradeHandler(handleTrade)
}

func handleTrade(t AggTrade) {
	mutex.Lock()
	defer mutex.Unlock()

	rounded := t.Timestamp.Truncate(time.Minute)
	key := t.Symbol + rounded.Format("200601021504")

	// Рассчитываем значение CVD в долларах с направлением
	value := float64(t.Direction) * t.Quantity * t.Price

	// Обновляем агрегаты
	cvdPerSymbol[t.Symbol] += value
	minuteAgg[key] += value

	last, ok := lastFlush[t.Symbol]
	if !ok {
		last = time.Time{} // нулевое значение
	}

	// Если текущая минута больше последней зафиксированной, то сохраним данные за прошлую
	if last.IsZero() || rounded.After(last) {
		if !last.IsZero() {
			prevKey := t.Symbol + last.Format("200601021504")
			val, ok := minuteAgg[prevKey]
			if !ok {
				val = 0
			}

			cvd := storage.MinuteCVD{
				Timestamp: last,
				Symbol:    t.Symbol,
				Value:     val,
			}
			storage.SaveMinuteCVD(cvd)
			delete(minuteAgg, prevKey)
		}
		lastFlush[t.Symbol] = rounded
	}
}
