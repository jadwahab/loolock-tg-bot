package helpers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/jadwahab/loolock-tg-bot/db"
)

func Refresh(dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	err := RefreshLeaderboard(dbp)
	if err != nil {
		return err
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
