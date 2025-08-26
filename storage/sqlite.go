package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(path string) *sql.DB {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS minute_cvd (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME,
		symbol TEXT,
		value REAL
	);
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	return db
}

func DB() *sql.DB {
	return db
}

func SaveMinuteCVD(cvd MinuteCVD) {
	stmt := `INSERT INTO minute_cvd (timestamp, symbol, value) VALUES (?, ?, ?)`
	_, err := db.Exec(stmt, cvd.Timestamp.Format(time.RFC3339), cvd.Symbol, cvd.Value)
	if err != nil {
		log.Printf("Failed to insert CVD: %v", err)
	}
}

func GetMinuteCVDBySymbol(symbol string) ([]MinuteCVD, error) {
	rows, err := db.Query(`
		SELECT timestamp, symbol, value FROM minute_cvd
		WHERE symbol = ? ORDER BY timestamp DESC LIMIT 100
	`, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MinuteCVD
	for rows.Next() {
		var m MinuteCVD
		var ts string
		if err := rows.Scan(&ts, &m.Symbol, &m.Value); err != nil {
			return nil, err
		}
		m.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		result = append(result, m)
	}
	return result, nil
}

func GetTopCVD(requestedMinutes int, limit int) ([]struct {
	Symbol string
	SumCVD float64
}, error) {
	// Узнаём минимальный timestamp в базе
	var minTimeStr string
	err := db.QueryRow(`SELECT MIN(timestamp) FROM minute_cvd`).Scan(&minTimeStr)
	if err != nil || minTimeStr == "" {
		return nil, fmt.Errorf("нет данных в базе")
	}

	// Конвертируем время из строки
	minTime, err := time.Parse(time.RFC3339, minTimeStr)
	if err != nil {
		return nil, err
	}

	// Сколько минут реально доступно?
	availableMinutes := int(time.Since(minTime).Minutes())
	if availableMinutes < requestedMinutes {
		requestedMinutes = availableMinutes
	}

	durationStr := "-" + strconv.Itoa(requestedMinutes) + " minutes"

	query := `
		SELECT symbol, SUM(value) as sum_cvd
		FROM minute_cvd
		WHERE timestamp >= datetime('now', ?)
		GROUP BY symbol
		ORDER BY ABS(sum_cvd) DESC
		LIMIT ?
	`
	rows, err := db.Query(query, durationStr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Symbol string
		SumCVD float64
	}
	for rows.Next() {
		var r struct {
			Symbol string
			SumCVD float64
		}
		if err := rows.Scan(&r.Symbol, &r.SumCVD); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Возвращает сумму CVD за последние duration минут по symbol
func GetSumCVDBySymbolAndDuration(symbol string, duration int) (float64, error) {
	query := `
	SELECT COALESCE(SUM(value), 0) FROM minute_cvd
	WHERE symbol = ? AND timestamp >= datetime('now', ?)
	`
	durationStr := fmt.Sprintf("-%d minutes", duration)
	var sum float64
	err := db.QueryRow(query, symbol, durationStr).Scan(&sum)
	return sum, err
}
