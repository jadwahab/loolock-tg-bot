package cmds

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/tonicpow/go-paymail"
)

const signingBitcomPrefix = "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc" // whitelisted on https://relayauth.libsv.dev

func WelcomeMessage(cfg config.Config, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	messages := []string{
		"gm ELITES",
		"Only top 100 LooLockers are allowed.",
		fmt.Sprintf("You have %d min to make me group admin or else I will leave.", int(cfg.AdminTimeout)),
	}
	for _, messageText := range messages {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, messageText))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
	}

	// Start a timer to leave if not made admin
	go func(chatID int64) {
		time.Sleep(time.Duration(cfg.AdminTimeout) * time.Minute)
		if IsUserAdmin(bot, chatID, bot.Self.ID) {
			thankYouMsg := "Thank you for making me an admin!"
			_, err := bot.Send(tgbotapi.NewMessage(chatID, thankYouMsg))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		} else {
			leaveMsg := "I wasn't made an admin in time, so I'm out. ✌️"
			_, err := bot.Send(tgbotapi.NewMessage(chatID, leaveMsg))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
			bot.Request(tgbotapi.LeaveChatConfig{ChatID: chatID})
		}
	}(update.Message.Chat.ID)
}

func IsUserAdmin(bot *tgbotapi.BotAPI, chatID int64, userID int64) bool {
	chatAdminConfig := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}

	admins, err := bot.GetChatAdministrators(chatAdminConfig)
	if err != nil {
		log.Printf("Failed to get chat admins: %s", err)
		return false
	}

	for _, admin := range admins {
		if admin.User.ID == userID {
			return true
		}
	}
	return false
}

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
			// TODO: handle returning user
			// check db to see if they are back on leaderboard
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

type KickArgs struct {
	ChatID       int64
	UserID       int64
	UserName     string
	KickDuration time.Duration
}

func KickUser(bot *tgbotapi.BotAPI, ka *KickArgs) {
	_, err := bot.Request(tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: ka.ChatID,
			UserID: ka.UserID,
		},
		UntilDate: int64(time.Now().Add(ka.KickDuration).Unix()),
	})
	if err != nil {
		log.Printf("Failed to kick user: %s", err)
	}
	_, err = bot.Send(tgbotapi.NewMessage(ka.ChatID, fmt.Sprintf("@%s has just been kicked...", ka.UserName)))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
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
			// TODO: validate sig with public key
			log.Printf("SIG: %s", sig)
			validSig := update.Message.Text == userChallenge.Challenge
			if validSig {
				// TODO: update db with sig and verified data
				delete(challengeUserMap, update.Message.From.ID)
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the group!"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
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

func AddUser(arg1, arg2 string, dbp db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	s, err := paymail.ValidateAndSanitisePaymail(arg1, false)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid paymail."))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}
	paymailAddress := s.Address

	amountLocked, err := strconv.ParseFloat(arg2, 64)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid amount."))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	pubkey, err := GetPubKey(paymailAddress)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting public key for your paymail"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	err = dbp.UpsertUser(amountLocked, paymailAddress, pubkey)
	if err != nil {
		log.Printf("Failed to insert entry with %f, %s, %s: %s", amountLocked, paymailAddress, pubkey, err)
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to add user to DB"))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}
		return
	}

	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "User added successfully!"))
	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}
}

const pkiBaseURL = "https://relayx.io/bsvalias/id/"

type PKIResponseData struct {
	BsvAlias string `json:"bsvalias"`
	Handle   string `json:"handle"`
	PubKey   string `json:"pubkey"`
}

func GetPubKey(paymail string) (string, error) { // TODO: get public key for any paymail (not just relayx)
	resp, err := http.Get(pkiBaseURL + paymail)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	data := &PKIResponseData{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return "", err
	}
	return data.PubKey, nil
}

func AdminCommand(cmd string, dbp db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	commandArgs := strings.Fields(cmd)

	switch commandArgs[0] {
	case "/adduser":
		if len(commandArgs) == 3 {
			AddUser(commandArgs[1], commandArgs[2], dbp, bot, update)
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid /adduser format. Use /adduser <paymail> <amount>."))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}

	case "/leaderboard":
		leaderboard, err := dbp.GetLeaderboard()
		if err != nil {
			_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting leaderboard from DB"))
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
		_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resultString))
		if err != nil {
			log.Printf("Failed to send message: %s", err)
		}

	case "/refresh":

	default:
		if strings.HasPrefix(commandArgs[0], "/") {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Use /adduser or /leaderboard or /refresh"))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}
	}
}
