package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"main/wordliebot"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	tgbot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", tgbot.Self.UserName)

	log.Printf("Preparing the dictionary...")
	wordliebot := wordliebot.NewWordlieBot(tgbot, "dictionary.txt")
	log.Printf("Bot is ready!")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := wordliebot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			go wordliebot.ProcessIncomingMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go wordliebot.ProcessIncomingCallbackQuery(update.CallbackQuery)
		}
	}
}
