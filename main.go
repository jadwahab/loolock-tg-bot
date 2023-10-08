package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
	_ "github.com/lib/pq"
)

const signingBitcomPrefix = "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc"

// Keeps track of user ID and their challenge string + number of attempts
var challengeUserMap = make(map[int64]*UserChallenge)

type UserChallenge struct {
	Challenge string
	Attempts  int
}

func main() {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	defer conn.Close()
	// Test the connection
	err = conn.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %q", err)
	}

	dbp := db.DBParams{
		DB: conn,
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN is required")
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = config.BotDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsPrivate() { // Skip private chats
			continue
		}

		if len(update.Message.NewChatMembers) > 0 {
			for _, newUser := range update.Message.NewChatMembers {
				if newUser.ID == bot.Self.ID { // Bot joins new group
					// Send welcome msgs
					messages := []string{
						"gm ELITES",
						"Only top 100 LooLockers are allowed.",
						fmt.Sprintf("You have %d min to make me group admin or else I will leave.", int(config.AdminTimeout)),
					}
					for _, messageText := range messages {
						_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, messageText))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
					}

					// Start a timer to leave if not made admin
					go func(chatID int64) {
						time.Sleep(time.Duration(config.AdminTimeout) * time.Minute)
						if isBotAdmin(bot, chatID) {
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
					break

				} else { // A new member joined, and it's not the bot, send challenge

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
							break
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
						time.Sleep(time.Duration(config.ResponseTimeout) * time.Minute)
						_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "No response from @"+newUser.UserName))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}

						if _, stillExists := challengeUserMap[userID]; stillExists {
							KickUser(bot, &KickArgs{
								ChatID:       chatID,
								UserID:       userID,
								KickDuration: time.Duration(config.KickDuration),
								UserName:     newUser.UserName,
							})
							delete(challengeUserMap, userID) // Remove user from challenge map
						}
					}(newUser.ID, update.Message.Chat.ID)
				}
			}

			continue
		}

		// User sent challenge response
		paymail, sig, valid := helpers.IsValidChallengeResponse(update.Message.Text)
		if update.Message != nil && update.Message.Text != "" && valid {

			leaderboardEntry, err := dbp.GetEntryByPaymail(paymail)
			if err != nil {
				log.Printf("Database error while fetching paymail: %v", err)
				continue

			} else if leaderboardEntry == nil { // paymail not found
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Your paymail is not on the leaderboard, gtfo"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				KickUser(bot, &KickArgs{
					ChatID:       update.Message.Chat.ID,
					UserID:       update.Message.From.ID,
					KickDuration: time.Duration(config.KickDuration),
					UserName:     update.Message.From.UserName,
				})
				delete(challengeUserMap, update.Message.From.ID)
				continue

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
								KickDuration: time.Duration(config.KickDuration),
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

		} else {
			_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid format, try again."))
			if err != nil {
				log.Printf("Failed to send message: %s", err)
			}
		}

	}
}

func isBotAdmin(bot *tgbotapi.BotAPI, chatID int64) bool {
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
		if admin.User.ID == bot.Self.ID {
			return true
		}
	}
	return false
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

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	return string(randomString)
}
