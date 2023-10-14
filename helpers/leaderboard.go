package helpers

import (
	"errors"
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

	lbes, err := dbp.GetLeaderboard(100) // TODO: top 100 for now
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
	bitcoiners, err := apis.GetBitcoiners()
	if err != nil {
		return err
	}

	if len(bitcoiners) != 100 {
		return errors.New("error getting enough bitcoiners from API")
	}

	return dbp.BatchUpsert(bitcoiners)
}

func UserExistsInLeaderboard(leaderboard []db.LeaderboardEntry, username string) bool {
	for _, entry := range leaderboard {
		if entry.TelegramUsername.Valid {
			if entry.TelegramUsername.String == username && entry.IsVerified {
				return true
			}
		}
	}
	return false
}
