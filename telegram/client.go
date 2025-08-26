package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"cvd-bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartTelegramBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "topcvd":
			handleTopCVD(bot, update.Message)
		case "cvd":
			handleCVD(bot, update.Message)
		default:
			handleUnknownCommand(bot, update.Message)
		}
	}
}

func handleTopCVD(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	duration := 10
	if args != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(args)); err == nil && val > 0 {
			duration = val
		}
	}

	result, err := storage.GetTopCVD(duration, 20)
	if err != nil {
		reply(bot, msg.Chat.ID, "Ошибка при получении данных: "+err.Error())
		return
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📊 Топ CVD за последние %d минут:\n\n", duration))

	for _, r := range result {
		sign := "+"
		if r.SumCVD < 0 {
			sign = "-"
		}
		amount := formatNumber(abs(r.SumCVD))
		b.WriteString(fmt.Sprintf("%s, %s%s $\n", r.Symbol, sign, amount))
	}

	reply(bot, msg.Chat.ID, b.String())
}

func handleCVD(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	if len(args) != 2 {
		reply(bot, msg.Chat.ID, "Использование: /cvd <тикер> <минуты>\nПример: /cvd btc 60")
		return
	}

	inputSymbol := strings.ToUpper(args[0])
	// Добавляем "USDT" если нет
	if !strings.HasSuffix(inputSymbol, "USDT") {
		inputSymbol += "USDT"
	}

	duration, err := strconv.Atoi(args[1])
	if err != nil || duration <= 0 {
		reply(bot, msg.Chat.ID, "Неверный период. Укажите положительное число минут.")
		return
	}

	sumCVD, err := storage.GetSumCVDBySymbolAndDuration(inputSymbol, duration)
	if err != nil {
		reply(bot, msg.Chat.ID, "Ошибка при получении данных: "+err.Error())
		return
	}

	sign := "+"
	if sumCVD < 0 {
		sign = "-"
	}

	formatted := formatNumber(abs(sumCVD))
	reply(bot, msg.Chat.ID, fmt.Sprintf("CVD за последние %d минут по %s: %s%s $", duration, inputSymbol, sign, formatted))
}

func handleUnknownCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	reply(bot, msg.Chat.ID, "Неизвестная команда")
}

func reply(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func formatNumber(n float64) string {
	s := strconv.FormatFloat(n, 'f', 0, 64)
	nRunes := []rune(s)

	var result []rune
	count := 0
	for i := len(nRunes) - 1; i >= 0; i-- {
		result = append([]rune{nRunes[i]}, result...)
		count++
		if count%3 == 0 && i != 0 {
			result = append([]rune{' '}, result...)
		}
	}
	return string(result)
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
