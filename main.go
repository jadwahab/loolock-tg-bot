package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
	_ "github.com/lib/pq"
)

// Keeps track of user ID and their challenge string + number of attempts
var challengeUserMap = make(map[int64]*UserChallenge)

type UserChallenge struct {
	Challenge string
	Attempts  int
}

func main() {
	config, err := LoadConfig("config.yaml")
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
						if !isBotAdmin(bot, chatID) {
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
						continue
					}

					if userEntry.IsVerified {
						_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome back fam!"))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
						break
					}

					// If the new member's telegram username is not found in the database
					if userEntry == nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please provide your paymail to verify your status on the leaderboard.")
						_, err = bot.Send(msg)
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
					}

					challenge := "1RELAYTEST|" + generateRandomString(20)
					challengeUserMap[newUser.ID] = &UserChallenge{Challenge: challenge, Attempts: 0} // Store challenge for this user

					messages := []string{
						"Welcome " + newUser.UserName,
						"Only top 100 LooLockers are allowed. Sign this message to prove you are on the leaderboard:",
						challenge,
					}

					for _, messageText := range messages {
						_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, messageText))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
					}

					// Start a timer to kick the user if they don't respond in `responseTimeout` minutes
					go func(userID int64, chatID int64) {
						time.Sleep(time.Duration(config.ResponseTimeout) * time.Minute)

						if _, stillExists := challengeUserMap[userID]; stillExists {
							KickUser(bot, chatID, userID, time.Duration(config.KickDuration))
							delete(challengeUserMap, userID) // Remove user from challenge map
						}
					}(newUser.ID, update.Message.Chat.ID)
				}
			}

			continue
		}

		// Placeholder check for a paymail format; this is rudimentary and might need more refinement.
		if len(update.Message.Text) > 5 && strings.Contains(update.Message.Text, "@") {
			userPaymail := update.Message.Text
			leaderboardEntry, err := dbp.GetEntryByPaymail(userPaymail)

			if err != nil {
				log.Printf("Database error while fetching paymail: %v", err)
				continue
			}

			if leaderboardEntry == nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, that paymail is not recognized. You will be removed from the group.")
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
				KickUser(bot, update.Message.Chat.ID, update.Message.From.ID, time.Duration(config.KickDuration))
				continue
			}

			// Handle responses to the challenge question
			if userChallenge, exists := challengeUserMap[update.Message.From.ID]; exists {
				if update.Message.Text == userChallenge.Challenge {
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
						KickUser(bot, update.Message.Chat.ID, update.Message.From.ID, time.Duration(config.KickDuration))
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

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	return string(randomString)
}

func KickUser(bot *tgbotapi.BotAPI, chatID int64, userID int64, kickDuration time.Duration) {
	_, err := bot.Request(tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate: int64(time.Now().Add(kickDuration).Unix()),
	})
	if err != nil {
		log.Printf("Failed to kick user: %s", err)
	}
}
