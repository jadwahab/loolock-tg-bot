package cmds

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

const signingBitcomPrefix = "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc" // whitelisted on https://relayauth.libsv.dev

func SendNewUserChallenge(newUser tgbotapi.User, bot *tgbotapi.BotAPI, chatID int64) {
	challenge := signingBitcomPrefix + "|@" + newUser.UserName + "|" + fmt.Sprintf("%d", newUser.ID)

	_, err := bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"You must verify a link between your telegram username and your onchain identity (currently RelayX handle).\n\n"+
			"Sign a message with your key. Use this website to sign: https://relayauth.libsv.dev?userInput=%s "+
			"and then copy and paste the result here.",
		challenge)))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
}

func HandleChallengeResponse(cfg config.Config, dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update, challenge, paymail, sig string) {
	pubkey, err := apis.GetPubKey(paymail)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting public key for your paymail"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	if helpers.VerifyBSM(pubkey, sig, challenge) { // sig verified
		err := dbp.UpdateVerifiedUser(paymail, update.Message.From.UserName, challenge, pubkey, sig, update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to update verified user in leaderboard table: %s", err)
		}
		// If challenge is valid, send them the group link
		_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Successfully verified! Join the group by clicking %s", cfg.Groups[config.Top100].Link)))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		log.Printf("Successfully verified user: %s, %d", update.Message.From.UserName, update.Message.From.ID)

	} else {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid signature."))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
	}
}
