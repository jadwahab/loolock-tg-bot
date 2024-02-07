package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/cmds/admin"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
)

func HandleNewUser(dbp *db.DBParams, cfg config.Config, bot *tgbotapi.BotAPI, newUser tgbotapi.User) error {
	chatID := cfg.Groups[config.TopLockers].ChatID

	// TODO: remove later
	c := tgbotapi.SetChatTitleConfig{
		ChatID: chatID,
		Title:  "TOP 50 Lockers",
	}
	if _, err := bot.Request(c); err != nil {
		log.Println("Error setting chat title:", err)
	} else {
		log.Println("Chat title changed successfully!")
	}
	//

	lbe, err := dbp.GetUserByTelegramID(newUser.ID)
	if err != nil {
		return err
	}
	if lbe == nil || !lbe.IsVerified {
		KickUser(bot, &KickArgs{
			ChatID:       chatID,
			UserID:       newUser.ID,
			KickDuration: time.Duration(cfg.KickDuration),
			UserName:     newUser.UserName,
			DBP:          dbp,
			KickMessage: "You were kicked because you are not verified.\n\n" +
				"DM /verify to @loolockbot to verify a link between your onchain identity and your telegram userID!",
		})
		return nil
	}

	// User exists and is verified
	members, err := dbp.GetCurrentMembers(chatID)
	if err != nil {
		return err
	}
	members = append(members, db.ChatUser{
		UserID:   newUser.ID,
		UserName: newUser.UserName,
		TgName:   sql.NullString{String: strings.TrimSpace(newUser.FirstName + " " + newUser.LastName)},
	})

	chatLimit := cfg.Groups[config.TopLockers].Limit
	newUserKicked := false

	if len(members) > chatLimit { // overflow
		membersToKick, err := DetermineMembersToKick(dbp, members, chatLimit)
		if err != nil {
			return err
		}

		// FIXME: remove if all is good
		if len(membersToKick) > 1 {
			err = admin.Notify(bot, fmt.Sprintf("membersToKick: %v. Length %d", membersToKick, len(membersToKick)))
			log.Println(err)
		}

		// Kick the overflowing member
		for _, km := range membersToKick {
			if km.UserID != bot.Self.ID {
				KickUser(bot, &KickArgs{
					ChatID:      chatID,
					UserID:      km.UserID,
					UserName:    km.UserName,
					DBP:         dbp,
					KickMessage: "You were kicked because the group has reached the maximum number of members, go lock up more!",
				})
			}
			if km.UserID == newUser.ID {
				newUserKicked = true
			}
		}
	}

	if newUserKicked {
		log.Printf("New user kicked from ChatID: %d, UserID: %d, UserName: %s", chatID, newUser.ID, newUser.UserName)
	} else {
		log.Printf("New user joined ChatID: %d, UserID: %d, UserName: %s", chatID, newUser.ID, newUser.UserName)
	}
	err = dbp.AddUserToGroupChatDB(chatID, newUser.ID, newUser.UserName, strings.TrimSpace(newUser.FirstName+" "+newUser.LastName), !newUserKicked)
	if err != nil {
		return err
	}

	return nil
}
