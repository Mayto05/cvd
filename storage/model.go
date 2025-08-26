package storage

import "time"

type MinuteCVD struct {
	Timestamp time.Time
	Symbol    string
	Value     float64
}