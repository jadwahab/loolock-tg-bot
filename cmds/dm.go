package cmds

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

func HandleDMs(cfg config.Config, dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	challenge, paymail, sig, valid := helpers.IsValidChallengeResponse(update.Message.Text)
	if update.Message.Text != "" && valid {
		// If challenge is valid, send them the group link
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("You're validated! Join the group by clicking %s", cfg.Groups[config.Top100].Link)))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}

		HandleChallengeResponse(dbp, bot, update, challenge, paymail, sig)
	}

	switch update.Message.Text {

	case "/verify":
		userEntry, err := dbp.GetUserByTelegramID(update.Message.From.ID)
		if err != nil {
			log.Printf("Database error while fetching user: %v", err)
		} else if userEntry != nil { // user found in db
			// fmt.Println("Found entry:", userEntry.TelegramUsername)
			if userEntry.IsVerified {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You are already verified!"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				return
			}
		}
		SendNewUserChallenge(*update.Message.From, bot, update.Message.Chat.ID)

	case "/top100":
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("After you have verified yourself using the /verify command, "+
				"join the group by clicking here: %s", cfg.Groups[config.Top100].Link)))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}

	case "/leaderboard":
		const lbLimit = 100
		PrintLeaderboard(dbp, bot, update.Message.Chat.ID, lbLimit)

	default:
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Invalid command. Use /verify or /top or /leaderboard\n\n"+
				"/verify - Verify your identity\n"+
				"/top100 - Get access to the TOP 100 lockers group\n"+
				"/leaderboard - Get the TOP 100 leaderboard"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
	}
}
