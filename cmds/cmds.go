package cmds

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
)

func WelcomeMessage(cfg config.Config, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	messages := []string{
		"gm ELITES",
		"Only top 100 LooLockers are allowed.",
		fmt.Sprintf("You have %d min to make me group admin or else I will leave.", int(cfg.AdminTimeout)),
	}
	for _, messageText := range messages {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, messageText))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
	}

	// Start a timer to leave if not made admin
	go func(chatID int64) {
		time.Sleep(time.Duration(cfg.AdminTimeout) * time.Minute)
		if !IsUserAdmin(bot, chatID, bot.Self.ID) {
			leaveMsg := "I wasn't made an admin in time, so I'm out. ✌️"
			_, err := bot.Send(tgbotapi.NewMessage(chatID, leaveMsg))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
			bot.Request(tgbotapi.LeaveChatConfig{ChatID: chatID})
		}
	}(update.Message.Chat.ID)
}

func IsUserAdmin(bot *tgbotapi.BotAPI, chatID int64, userID int64) bool {
	chatAdminConfig := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}

	admins, err := bot.GetChatAdministrators(chatAdminConfig)
	if err != nil {
		log.Printf("Failed to get chat admins: %s", err)
		return false
	}

	for _, admin := range admins {
		if admin.User.ID == userID {
			return true
		}
	}
	return false
}

type KickArgs struct {
	ChatID       int64
	UserID       int64
	UserName     string
	KickDuration time.Duration
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
	} else {
		_, err = bot.Send(tgbotapi.NewMessage(ka.ChatID, fmt.Sprintf("@%s has just been kicked...", ka.UserName)))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
	}
}
