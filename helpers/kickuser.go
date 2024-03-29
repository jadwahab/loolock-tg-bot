package helpers

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
)

type KickArgs struct {
	ChatID       int64
	UserID       int64
	UserName     string
	KickDuration time.Duration
	DBP          *db.DBParams
	KickMessage  string
}

func KickUser(bot *tgbotapi.BotAPI, ka *KickArgs) {
	if ka.KickDuration == 0 {
		ka.KickDuration = 1
	}

	_, err := bot.Request(tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: ka.ChatID,
			UserID: ka.UserID,
		},
		UntilDate: int64(time.Now().Add(time.Duration(ka.KickDuration) * time.Minute).Unix()),
	})
	if err != nil {
		log.Printf("Failed to kick user: %s", err)
		return
	} else {
		log.Printf("Kicked %s (%d)\n", ka.UserName, ka.UserID)
	}

	// Immediately unban the user after kicking
	_, err = bot.Request(tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: ka.ChatID,
			UserID: ka.UserID,
		},
		OnlyIfBanned: true,
	})
	if err != nil {
		log.Printf("Failed to unban user: %s", err)
	}

	err = ka.DBP.UpdateUserLeftAt(ka.ChatID, ka.UserID)
	if err != nil {
		log.Printf("Failed to update user left in DB: %s", err)
	}

	// Send a message to the user
	_, err = bot.Send(tgbotapi.NewMessage(ka.UserID, ka.KickMessage))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
}
