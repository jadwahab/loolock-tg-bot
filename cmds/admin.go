package cmds

import (
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
	"github.com/tonicpow/go-paymail"
)

func AdminCommand(cmd string, dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	commandArgs := strings.Fields(cmd)

	switch commandArgs[0] {
	case "/adduser":
		if len(commandArgs) == 3 {
			AddUser(commandArgs[1], commandArgs[2], dbp, bot, update)
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid /adduser format. Use /adduser <paymail> <amount>."))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}

	case "/leaderboard":
		PrintLeaderboard(dbp, bot, update.Message.Chat.ID)

	case "/refresh":
		helpers.Refresh(config.Config{KickDuration: 0}, dbp, bot, update.Message.Chat.ID)

	default:
		if strings.HasPrefix(commandArgs[0], "/") {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Use /adduser or /leaderboard or /refresh"))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}
	}
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
