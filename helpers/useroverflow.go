package helpers

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/cmds/admin"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/pkg/errors"
)

func HandleUserOverflow(dbp *db.DBParams, cfg config.Config, bot *tgbotapi.BotAPI, chatID int64, chatLimit int) error {
	members, err := dbp.GetCurrentMembers(cfg.Groups[config.TopLockers].ChatID)
	if err != nil {
		err = admin.Notify(bot, "Failed to get current members: "+err.Error())
		if err != nil {
			log.Printf("Failed to notify admin: %s", err)
		}
	}

	if len(members) < chatLimit {
		return errors.New("members count is less than chat limit")
	}

	membersToKick, err := DetermineMembersToKick(dbp, members, chatLimit)
	if err != nil {
		return err
	}

	// Kick the overflowing members
	for _, km := range membersToKick {
		if km.UserID != bot.Self.ID {
			KickUser(bot, &KickArgs{
				ChatID:       chatID,
				UserID:       km.UserID,
				UserName:     km.UserName,
				KickDuration: time.Duration(cfg.KickDuration),
				DBP:          dbp,
				KickMessage:  "You were kicked because the group has reached the maximum number of members!",
			})
		}
	}

	return nil
}
