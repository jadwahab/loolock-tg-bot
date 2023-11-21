package cmds

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

func HandleDMs(cfg config.Config, dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Printf("Received DM from [%s:%d] %s", update.Message.From.UserName, update.Message.From.ID, update.Message.Text)

	// TODO: handle challenge with pubkey not paymail
	challengeResponse, valid := helpers.IsValidChallengeResponse(update.Message.Text)
	if update.Message.Text != "" && valid {
		HandleChallengeResponse(cfg, dbp, bot, update, challengeResponse)
		return
	}

	const lbLimit = 100
	const top100ChatID = -1001984238822

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
		PrintLeaderboard(dbp, bot, update.Message.Chat.ID, lbLimit)

	// unexposed commands:
	case "/announce":
		_, err := bot.Send(tgbotapi.NewMessage(top100ChatID, "GM elites!\n\n"+
			"Leaderboard calculations have now gone from using just amount LOCKED "+
			"to a combitation of amount LOCKED + LIKED!\n\nThis means the leaderbaord "+
			"will change so don't be surprised if you get kicked - go lock up more!"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}

	case "/kickintruders":
		users, err := dbp.GetUniqueUsers(top100ChatID)
		if err != nil {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to fetch users from database"))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}
		lbes, err := dbp.GetLeaderboard(true, lbLimit, "both")
		if err != nil {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to get leaderboard"))
			if err != nil {
				log.Printf("Failed to get leaderboard: %s", err)
			}
		}
		for _, user := range users {
			exists := helpers.UserExistsInLeaderboard(lbes, user.UserID)
			if !exists && user.UserID != bot.Self.ID {
				helpers.KickUser(bot, &helpers.KickArgs{
					ChatID:       top100ChatID,
					UserID:       user.UserID,
					KickDuration: time.Duration(cfg.KickDuration),
					UserName:     user.UserName,
					DBP:          dbp,
					KickMessage:  "You were kicked because you are no longer in the top 100 lockers!",
				})
			}
		}

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
