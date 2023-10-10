package cmds

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

const signingBitcomPrefix = "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc" // whitelisted on https://relayauth.libsv.dev

type UserChallenge struct {
	Challenge string
	Attempts  int
}

func SendNewUserChallenge(cfg config.Config, dbp db.DBParams, newUser tgbotapi.User,
	bot *tgbotapi.BotAPI, update tgbotapi.Update, challengeUserMap map[int64]*UserChallenge) {

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
				err = dbp.AddUserToGroupChatDB(update.Message.Chat.ID, update.Message.From.ID, update.Message.From.UserName)
				if err != nil {
					log.Printf("Failed to add user to group table: %s", err)
				}
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the group @"+update.Message.From.UserName+"!"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
			} else {
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You fell off the leaderboard!\nNGMI..."))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				KickUser(bot, &KickArgs{
					ChatID:       update.Message.Chat.ID,
					UserID:       update.Message.From.ID,
					KickDuration: time.Duration(cfg.KickDuration),
					UserName:     update.Message.From.UserName,
				})
			}
			return
		}
	}

	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
		"Welcome @%s\n\n"+
			"Only top 100 LooLockers are allowed.\n\n"+
			"To prove that you are on the leaderboard, please sign this message "+
			"and then send 1 message with relay paymail on 1st line and signature on 2nd.\n"+ // TODO: change to pub key
			"Use this website to sign: https://relayauth.libsv.dev/"+
			"\n\nExample:\n\n"+
			"jek@relayx.io\nIJDiGEdovFRf/U2WtJ6WJz59eBupAuZDJKXe0/O1aJvAYSF4xGW2ZllIUX6cybm51Uv5f1GRID41v7bcIVr4Jrk=",
		newUser.UserName)))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}

	challenge := signingBitcomPrefix + "|" + newUser.UserName + "|" + generateRandomString(5)
	// Store challenge for this user
	challengeUserMap[newUser.ID] = &UserChallenge{Challenge: challenge, Attempts: 0}

	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, challenge))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}

	// Start a timer to kick the user if they don't respond in `responseTimeout` minutes
	go func(userID int64, chatID int64) {
		time.Sleep(time.Duration(cfg.ResponseTimeout) * time.Minute)
		_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "No response from @"+newUser.UserName))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}

		if _, stillExists := challengeUserMap[userID]; stillExists {
			KickUser(bot, &KickArgs{
				ChatID:       chatID,
				UserID:       userID,
				KickDuration: time.Duration(cfg.KickDuration),
				UserName:     newUser.UserName,
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

func HandleChallengeResponse(cfg config.Config, dbp db.DBParams,
	bot *tgbotapi.BotAPI, update tgbotapi.Update, challengeUserMap map[int64]*UserChallenge,
	paymail, sig string) {

	leaderboardEntry, err := dbp.GetEntryByPaymail(paymail)
	if err != nil {
		log.Printf("Database error while fetching paymail: %v", err)
		return

	} else if leaderboardEntry == nil { // paymail not found
		_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Your paymail is not on the leaderboard, gtfo"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		KickUser(bot, &KickArgs{
			ChatID:       update.Message.Chat.ID,
			UserID:       update.Message.From.ID,
			KickDuration: time.Duration(cfg.KickDuration),
			UserName:     update.Message.From.UserName,
		})
		delete(challengeUserMap, update.Message.From.ID)
		return

	} else { // paymail found
		if userChallenge, exists := challengeUserMap[update.Message.From.ID]; exists {

			if helpers.VerifyBSM(leaderboardEntry.PublicKey, sig, userChallenge.Challenge) { // sig verified
				err := dbp.UpdateVerifiedUser(paymail, update.Message.From.UserName, userChallenge.Challenge, sig)
				if err != nil {
					log.Printf("Failed to update verified user in leaderboard table: %s", err)
				}
				err = dbp.AddUserToGroupChatDB(update.Message.Chat.ID, update.Message.From.ID, update.Message.From.UserName)
				if err != nil {
					log.Printf("Failed to add user to group table: %s", err)
				}
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the group @"+update.Message.From.UserName+"!"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				delete(challengeUserMap, update.Message.From.ID)

			} else {
				userChallenge.Attempts++
				if userChallenge.Attempts >= 3 { // out of attempts
					_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Incorrect answer"))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
					}
					_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "HFSP"))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
					}
					delete(challengeUserMap, update.Message.From.ID)
					KickUser(bot, &KickArgs{
						ChatID:       update.Message.Chat.ID,
						UserID:       update.Message.From.ID,
						KickDuration: time.Duration(cfg.KickDuration),
						UserName:     update.Message.From.UserName,
					})
				} else { // try again + increment attempts
					_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Incorrect answer. You have %d attempts left.", 3-userChallenge.Attempts)))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
					}
				}
			}
		}
	}
}
