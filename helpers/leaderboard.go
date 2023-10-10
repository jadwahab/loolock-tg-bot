package helpers

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
)

func Refresh(cfg config.Config, dbp *db.DBParams, bot *tgbotapi.BotAPI, chatID int64) error {
	err := RefreshLeaderboard(dbp)
	if err != nil {
		return err
	}

	lbes, err := dbp.GetLeaderboard()
	if err != nil {
		return err
	}

	count, err := bot.GetChatMembersCount(tgbotapi.ChatMemberCountConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	})
	if err != nil {
		log.Println("Error fetching member count:", err)
	}

	users, err := dbp.GetUniqueUsers(chatID)
	if err != nil {
		log.Println("Error getting unique user IDs:", err)
	}

	if count != len(users) {
		log.Println("ERROR - count of members not equal to user ID list:")
	}

	for _, user := range users {
		exists := UserExistsInLeaderboard(lbes, user.UserName)
		if !exists && user.UserID != bot.Self.ID {
			KickUser(bot, &KickArgs{
				ChatID:       chatID,
				UserID:       user.UserID,
				KickDuration: time.Duration(cfg.KickDuration),
				UserName:     user.UserName,
				DBP:          dbp,
			})
			if err != nil {
				log.Println("Error kicking member:", err)
			} else {
				log.Printf("Kicked %s (%d)\n", user.UserName, user.UserID)
			}
		}
	}

	return nil
}

func RefreshLeaderboard(dbp *db.DBParams) error {
	top100, err := apis.GetTop100Bitcoiners()
	if err != nil {
		return err
	}

	return dbp.BatchUpsert(top100)
}

func UserExistsInLeaderboard(leaderboard []db.LeaderboardEntry, username string) bool {
	for _, entry := range leaderboard {
		if entry.TelegramUsername == username {
			return true
		}
	}
	return false
}
