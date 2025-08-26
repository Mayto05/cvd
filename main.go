package main

import (
	"database/sql"
	"log"

	"cvd-bot/exchange"
	"cvd-bot/storage"
	"cvd-bot/telegram"
)

func main() {
	db := initDatabase()
	defer db.Close()

	startExchange()

	telegram.StartTelegramBot("8441208120:AAGNAU0gyuuJhY5-AaRzW5CZ6S8XHBHbdBo")
}

func initDatabase() *sql.DB {
	db := storage.InitDB("cvd.db")
	return db
}

func startExchange() {
	exchange.StartAggregator()

	log.Println("Connecting to Binance...")
	symbols := exchange.DownloadSymbols()
	exchange.StartStream(symbols)
}
