package cmds

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

const signingBitcomPrefix = "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc" // whitelisted on https://relayauth.libsv.dev

type UserChallenge struct {
	Challenge string
	Attempts  int
}

func SendNewUserChallenge(cfg config.Config, dbp *db.DBParams, newUser tgbotapi.User,
	bot *tgbotapi.BotAPI, update tgbotapi.Update, challengeUserMap map[int64]string) {

	userEntry, err := dbp.GetUserByTelegramUsername(newUser.UserName)
	if err != nil {
		log.Printf("Database error while fetching user: %v", err)
	} else if userEntry != nil { // user found in db
		// fmt.Println("Found entry:", userEntry.TelegramUsername)
		if userEntry.IsVerified {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome back fam!"))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}

			onLeaderboard, err := dbp.IsPaymailInTop100(userEntry.Paymail)
			if err != nil {
				log.Printf("Failed to check leaderboard DB: %s", err)
			}

			if onLeaderboard {
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the group @"+update.Message.From.UserName+"!"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
			} else {
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You fell off the leaderboard!\nNGMI..."))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				helpers.KickUser(bot, &helpers.KickArgs{
					ChatID:       update.Message.Chat.ID,
					UserID:       update.Message.From.ID,
					KickDuration: time.Duration(cfg.KickDuration),
					UserName:     update.Message.From.UserName,
					DBP:          dbp,
				})
			}
			return
		}
	}

	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
		"Welcome @%s\n\n"+
			"Only top 100 LooLockers are allowed.\n\n"+
			"To prove that you are on the leaderboard, please sign this message "+
			"and then send 1 message with challenge message I will send you in "+
			"separate message on 1st line, relay paymail on 2nd and signature on 3rd.\n"+ // TODO: change to pub key
			"Use this website to sign: https://relayauth.libsv.dev/"+
			"\n\nExample:\n\n"+
			"17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc|unfudabledragon|FQ5im\njek@relayx.io\n"+
			"IJDiGEdovFRf/U2WtJ6WJz59eBupAuZDJKXe0/O1aJvAYSF4xGW2ZllIUX6cybm51Uv5f1GRID41v7bcIVr4Jrk=",
		newUser.UserName)))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}

	challenge := signingBitcomPrefix + "|" + newUser.UserName + "|" + generateRandomString(5)
	// Store challenge for this user
	challengeUserMap[newUser.ID] = challenge

	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, challenge))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}

	// Start a timer to kick the user if they don't respond in `responseTimeout` minutes
	go func(userID int64, chatID int64) {
		time.Sleep(time.Duration(cfg.ResponseTimeout) * time.Minute)

		if _, stillExists := challengeUserMap[userID]; stillExists {
			_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "No response from @"+newUser.UserName))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
			helpers.KickUser(bot, &helpers.KickArgs{
				ChatID:       chatID,
				UserID:       userID,
				KickDuration: time.Duration(cfg.KickDuration),
				UserName:     newUser.UserName,
				DBP:          dbp,
			})
			delete(challengeUserMap, userID) // Remove user from challenge map
		}
	}(newUser.ID, update.Message.Chat.ID)

}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	return string(randomString)
}

func HandleChallengeResponse(cfg config.Config, dbp *db.DBParams,
	bot *tgbotapi.BotAPI, update tgbotapi.Update, challengeUserMap map[int64]string,
	paymail, sig string) {

	exists, err := dbp.PaymailExists(paymail)
	if err != nil {
		log.Printf("Database error while fetching paymail: %v", err)
		return
	}

	if !exists { // paymail not found
		_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Your paymail is not on the leaderboard, gtfo"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		helpers.KickUser(bot, &helpers.KickArgs{
			ChatID:       update.Message.Chat.ID,
			UserID:       update.Message.From.ID,
			KickDuration: time.Duration(cfg.KickDuration),
			UserName:     update.Message.From.UserName,
			DBP:          dbp,
		})
		delete(challengeUserMap, update.Message.From.ID)
		return

	} else { // paymail found
		if challenge, exists := challengeUserMap[update.Message.From.ID]; exists {

			pubkey, err := apis.GetPubKey(paymail)
			if err != nil {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting public key for your paymail"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				return
			}

			if helpers.VerifyBSM(pubkey, sig, challenge) { // sig verified
				err := dbp.UpdateVerifiedUser(paymail, update.Message.From.UserName, challenge, pubkey, sig)
				if err != nil {
					log.Printf("Failed to update verified user in leaderboard table: %s", err)
				}
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the group @"+update.Message.From.UserName+"!"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				delete(challengeUserMap, update.Message.From.ID)

			} else {
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Incorrect answer. Sign the message: %s", challenge)))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
			}
		}
	}
}
