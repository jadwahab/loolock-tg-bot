package helpers

import (
	"fmt"
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
}

func KickUser(bot *tgbotapi.BotAPI, ka *KickArgs) {
	_, err := bot.Request(tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: ka.ChatID,
			UserID: ka.UserID,
		},
		UntilDate: int64(time.Now().Add(ka.KickDuration).Unix()),
	})
	if err != nil {
		log.Printf("Failed to kick user: %s", err)
		return
	}

	// Immediately unban the user after kicking
	_, err = bot.Request(tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: ka.ChatID,
			UserID: ka.UserID,
		},
	})
	if err != nil {
		log.Printf("Failed to unban user: %s", err)
	}

	_, err = bot.Send(tgbotapi.NewMessage(ka.ChatID, fmt.Sprintf("@%s has just been kicked...", ka.UserName)))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
	err = ka.DBP.RemoveUserFromGroupChatDB(ka.ChatID, ka.UserID)
	if err != nil {
		log.Printf("Failed to remove user from group chat DB: %s", err)
	}
}
