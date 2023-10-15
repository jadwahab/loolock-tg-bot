package cmds

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

func PrintLeaderboard(dbp *db.DBParams, bot *tgbotapi.BotAPI, chatID int64, lbLimit int) {
	_, err := bot.Send(tgbotapi.NewMessage(chatID, "Fetching and updating leaderboard..."))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
	err = helpers.RefreshLeaderboard(dbp)
	if err != nil {
		log.Println(err.Error())
		_, err = bot.Send(tgbotapi.NewMessage(chatID, "Error getting leaderboard from API"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	leaderboard, err := dbp.GetLeaderboard(lbLimit)
	if err != nil {
		log.Println(err.Error())
		_, err = bot.Send(tgbotapi.NewMessage(chatID, "Error getting leaderboard from DB"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	var sb strings.Builder
	for i, user := range leaderboard {
		sb.WriteString(fmt.Sprintf("%d- %s - %f\n", i+1, user.Paymail, user.AmountLocked))
	}
	resultString := sb.String()
	_, err = bot.Send(tgbotapi.NewMessage(chatID, resultString))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
}
