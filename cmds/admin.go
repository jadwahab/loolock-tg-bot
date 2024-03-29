package cmds

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
)

func AdminCommand(cfg config.Config, cmd string, dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	commandArgs := strings.Fields(cmd)

	switch commandArgs[0] {
	case "/leaderboard":
		PrintLeaderboard(dbp, bot, update.Message.Chat.ID, cfg.Groups[config.TopLockers].Limit)

	default:
		if strings.HasPrefix(commandArgs[0], "/") {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Use /leaderboard"))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}
	}
}
