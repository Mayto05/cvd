📊 CVD Bot (Go)

Бот для расчёта и визуализации Cumulative Volume Delta (CVD) на основе торговых данных Binance.
Реализован на языке Go, использует Binance WebSocket API и SQLite для хранения данных.
Может быть расширен до Telegram-бота для отправки графиков/сигналов.

⸻

🚀 Возможности
	•	Подключение к Binance через WebSocket.
	•	Поддержка нескольких торговых пар (по умолчанию: BTCUSDT).
	•	Подсчёт Cumulative Volume Delta (покупки – продажи).
	•	Сохранение данных в SQLite (storage/sqlite.go).
	•	Агрегация и обработка данных (exchange/aggregator.go).
	•	Лёгкая интеграция с Telegram-ботом для нотификаций.

⸻

📂 Структура проекта

.
├── exchange/
│   ├── aggregator.go   # Логика агрегации сделок и расчёта CVD
│   └── binance.go      # Коннектор к Binance WebSocket
├── storage/
│   ├── sqlite.go       # Работа с SQLite
│   └── model.go        # Модели данных (Trade, Candle, CVD и т.д.)
├── cmd/
│   └── main.go         # Точка входа: запуск бота
├── go.mod
└── README.md


⸻

⚙️ Установка

# Клонируем репозиторий
git clone https://github.com/username/cvd-bot.git
cd cvd-bot

# Ставим зависимости
go mod tidy


⸻

▶️ Запуск

go run cmd/main.go

По умолчанию бот подключается к Binance и начинает собирать сделки.
Данные сохраняются в SQLite (cvd.db в корне проекта).


⸻

📈 Использование
	•	Данные CVD можно забирать из БД и строить графики (через Python, TradingView Pine Script или Go-библиотеки для графиков).
	•	Возможные расширения:
	•	REST API для текущего CVD;
	•	Отправка графиков и сигналов в Telegram;
	•	Запуск в Docker или systemd.

⸻

🧑‍💻 TODO
	•	Сделать REST API для получения текущего CVD.
	•	Добавить Telegram-бота для уведомлений.
	•	Вынести конфигурацию в отдельный файл.
	•	Написать тесты.
