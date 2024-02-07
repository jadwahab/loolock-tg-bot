package admin

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/tonicpow/go-paymail"
)

const adminChatID = 491301086

func Notify(bot *tgbotapi.BotAPI, message string) error {
	log.Println("Sending message to admin: ", message)
	_, err := bot.Send(tgbotapi.NewMessage(adminChatID, message))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
	return nil
}

func AddUser(arg1, arg2 string, dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	s, err := paymail.ValidateAndSanitisePaymail(arg1, false)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid paymail."))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}
	paymailAddress := s.Address

	amountLocked, err := strconv.ParseFloat(arg2, 64)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid amount."))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	err = dbp.UpsertUser(amountLocked, paymailAddress)
	if err != nil {
		log.Printf("Failed to insert entry with %f, %s: %s", amountLocked, paymailAddress, err)
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to add user to DB"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "User added successfully!"))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
}
