package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
				tu.KeyboardButton("що є поїсти ?"), // Added button "Горошок"
				tu.KeyboardButton("я долбойоб ?"),  // Added button "Дауни"
			),
			tu.KeyboardRow(
				tu.KeyboardButton("де заробить грошей ?"), // Added button "Іди крадь"
				tu.KeyboardButton("А шо не так ?"),        // Added button "Гатила"
			),
			tu.KeyboardRow(
				tu.KeyboardButton("шо робить з цим підром ?"), // Added button "Іди крадь"
				tu.KeyboardButton("а Сухлякі шо ?"),           // Added button "Гатила"
			),
			tu.KeyboardRow(
				tu.KeyboardButton("як ти ?"), // Added button "Іди крадь"
				tu.KeyboardButton("горілка"), // Added button "Гатила"
			),
			tu.KeyboardRow(
				tu.KeyboardButton("приучила"), // Added button "Іди крадь"
				tu.KeyboardButton("єбало"),    // Added button "Гатила"
			),
			tu.KeyboardRow(
				tu.KeyboardButton("хуйло"),                   // Added button "Іди крадь"
				tu.KeyboardButton("мене кінув шеф на гроші"), // Added button "Гатила"
			),
			tu.KeyboardRow(
				tu.KeyboardButton("Реферальне посилання"), // Added referral button
			),
		)

		// Prepare message with custom keyboard
		message := tu.Message(
			tu.ID(chatID),
			"Натискай на кнопки, щоб отримати голосове повідомлення або скористайтеся Реферальним посиланням",
		).WithReplyMarkup(keyboard)

		// Send message with custom keyboard
		_, err := bot.SendMessage(message)
		if err != nil {
			log.Println("Failed to send message:", err)
			return
		}

	}, th.CommandEqual("start"))

	// Function to send voice message and handle consecutive send count
	sendVoice := func(bot *telego.Bot, update telego.Update, voiceFile, button string) {
		if update.Message == nil {
			return
		}

		chatID := update.Message.Chat.ID

		// Check if the same voice message was sent consecutively
		if lastVoiceSent[chatID] == button {
			consecutiveVoiceSendCount[chatID]++
		} else {
			consecutiveVoiceSendCount[chatID] = 1
		}

		// Update the last voice message sent
		lastVoiceSent[chatID] = button

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

		// Check if the text matches any trigger
		switch text {
		case "що є поїсти ?":
			sendVoice(bot, update, "voice/2.ogg", "що є поїсти ?")
		case "я долбойоб ?":
			sendVoice(bot, update, "voice/two_downs.ogg", "я долбойоб ?")
		case "де заробить грошей ?":
			sendVoice(bot, update, "voice/krad_.ogg", "де заробить грошей ?")
		case "А шо не так ?":
			sendVoice(bot, update, "voice/hatyla.ogg", "А шо не так ?")
		case "горілка":
			sendVoice(bot, update, "voice/kol_ka.ogg", "горілка")
		case "як ти ?":
			sendVoice(bot, update, "voice/worse.ogg", "як ти ?")
		case "а Сухлякі шо ?":
			sendVoice(bot, update, "voice/suchlak.ogg", "а Сухлякі шо ?")
		case "шо робить з цим підром ?":
			sendVoice(bot, update, "voice/axe.ogg", "шо робить з цим підром ?")
		case "єбало":
			sendVoice(bot, update, "voice/ebalo.ogg", "єбало")
		case "хуйло":
			sendVoice(bot, update, "voice/huilo.ogg", "хуйло")
		case "приучила":
			sendVoice(bot, update, "voice/priuchila.ogg", "приучила")
		case "мене кінув шеф на гроші":
			sendVoice(bot, update, "voice/verovka.ogg", "мене кінув шеф на гроші")
		case "Реферальне посилання":
			sendReferralLink(bot, update.Message.Chat.ID)
		}

	}, th.AnyMessage())

	bh.Start()
}

// Function to send referral link
func sendReferralLink(bot *telego.Bot, chatID int64) {
	// Generate a unique referral link (example: using chat ID as a parameter)
	referralLink := "https://example.com/telegram-bot/referral?user_id=" + strconv.FormatInt(chatID, 10)

	// Message with referral link
	message := tu.Message(
		tu.ID(chatID),
		"Скористайтеся цим посиланням для продовження на вашому телефоні: "+referralLink,
	)

	// Send message with referral link
	_, err := bot.SendMessage(message)
	if err != nil {
		log.Println("Failed to send referral link:", err)
	}
}
