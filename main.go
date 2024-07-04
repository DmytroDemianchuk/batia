package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// MaxConsecutiveVoiceSends defines the maximum number of consecutive sends for the same voice message
const MaxConsecutiveVoiceSends = 3

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env not loaded")
	}

	botToken := os.Getenv("TG_API_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TG_API_BOT_TOKEN is not set")
	}

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)
	bh, _ := th.NewBotHandler(bot, updates)

	defer bh.Stop()
	defer bot.StopLongPolling()

	// Initialize map to keep track of the consecutive sends for each voice message per chat
	consecutiveVoiceSendCount := make(map[int64]int)
	lastVoiceSent := make(map[int64]string)

	// Handler for command "start"
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		chatID := update.Message.Chat.ID

		// Create custom keyboard with buttons
		keyboard := tu.Keyboard(
			tu.KeyboardRow(
				tu.KeyboardButton("що є поїсти ?"),
				tu.KeyboardButton("я долбойоб ?"),
			),
			tu.KeyboardRow(
				tu.KeyboardButton("де заробить грошей ?"),
				tu.KeyboardButton("А шо не так ?"),
			),
			tu.KeyboardRow(
				tu.KeyboardButton("шо робить з цим підром ?"),
				tu.KeyboardButton("а Сухлякі шо ?"),
			),
			tu.KeyboardRow(
				tu.KeyboardButton("як ти ?"),
				tu.KeyboardButton("горілка"),
			),
			tu.KeyboardRow(
				tu.KeyboardButton("приучила"),
				tu.KeyboardButton("єбало"),
			),
			tu.KeyboardRow(
				tu.KeyboardButton("хуйло"),
				tu.KeyboardButton("мене кінув шеф на гроші"),
			),
		)

		// Prepare message with custom keyboard
		message := tu.Message(
			tu.ID(chatID),
			"Натискай на кнопки, щоб отримати голосове повідомлення",
		).WithReplyMarkup(keyboard)

		// Send message with custom keyboard
		_, err := bot.SendMessage(message)
		if err != nil {
			log.Println("Failed to send message:", err)
			return
		}

	}, th.CommandEqual("start"))

	// Function to send voice message and handle consecutive send count
	sendVoice := func(bot *telego.Bot, update telego.Update, voiceFile string) {
		if update.Message == nil {
			return
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		// Check if the same voice message was sent consecutively
		if lastVoiceSent[chatID] == text {
			consecutiveVoiceSendCount[chatID]++
		} else {
			consecutiveVoiceSendCount[chatID] = 1
		}

		// Update the last voice message sent
		lastVoiceSent[chatID] = text

		// Check if the consecutive send count exceeds the maximum allowed
		if consecutiveVoiceSendCount[chatID] > MaxConsecutiveVoiceSends {
			errorMessage := tu.Message(
				tu.ID(chatID),
				"Вибери інше голосове повідомлення",
			)
			_, err := bot.SendMessage(errorMessage)
			if err != nil {
				log.Println("Failed to send error message:", err)
			}
			return
		}

		// Open file for upload
		voice, err := os.Open(voiceFile)
		if err != nil {
			log.Println("Failed to open voice file:", err)
			return
		}
		defer voice.Close()

		// Set parameters for sending voice message
		params := &telego.SendVoiceParams{
			ChatID: tu.ID(chatID),
			Voice:  tu.File(voice),
		}

		// Send voice message
		_, err = bot.SendVoice(params)
		if err != nil {
			log.Println("Failed to send voice message:", err)
			return
		}
	}

	// Handler for text messages
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		if update.Message == nil {
			return
		}

		text := update.Message.Text

		// Map of text triggers to voice files
		voiceResponses := map[string]string{
			"що є поїсти ?":            "voice/2.ogg",
			"я долбойоб ?":             "voice/two_downs.ogg",
			"де заробить грошей ?":     "voice/krad_.ogg",
			"А шо не так ?":            "voice/hatyla.ogg",
			"горілка":                  "voice/kol_ka.ogg",
			"як ти ?":                  "voice/worse.ogg",
			"а Сухлякі шо ?":           "voice/suchlak.ogg",
			"шо робить з цим підром ?": "voice/axe.ogg",
			"єбало":                    "voice/ebalo.ogg",
			"хуйло":                    "voice/huilo.ogg",
			"приучила":                 "voice/priuchila.ogg",
			"мене кінув шеф на гроші":  "voice/verovka.ogg",
		}

		// Check if the text matches any trigger
		if voiceFile, found := voiceResponses[text]; found {
			sendVoice(bot, update, voiceFile)
		}

	}, th.AnyMessage())

	bh.Start()
}
