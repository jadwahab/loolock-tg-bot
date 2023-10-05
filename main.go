package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const adminTimeout = 1    // Timeout for how many mins bot waits to be made admin before leaving
const kickTimeout = 24    // Timeout for how many hours a user is banned from group after being kicked
const responseTimeout = 5 // Timeout for how many mins to wait with no response before kicking

// Keeps track of user ID and their challenge string + number of attempts
var challengeUserMap = make(map[int64]*UserChallenge)

type UserChallenge struct {
	Challenge string
	Attempts  int
}

func main() {
	// db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	log.Fatalf("Error opening database: %q", err)
	// }
	// defer db.Close()
	// // Test the connection
	// err = db.Ping()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return nil, err
	// }

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN is required")
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
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
						fmt.Sprintf("You have %d min to make me group admin or else I will leave.", int(adminTimeout)),
					}
					for _, messageText := range messages {
						_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, messageText))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
					}

					// Start a timer to leave if not made admin
					go func(chatID int64) {
						time.Sleep(adminTimeout * time.Minute)
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
						time.Sleep(responseTimeout * time.Minute)

						if _, stillExists := challengeUserMap[userID]; stillExists {
							kickConfig := tgbotapi.KickChatMemberConfig{
								ChatMemberConfig: tgbotapi.ChatMemberConfig{
									ChatID: chatID,
									UserID: userID,
								},
								UntilDate: int64(time.Now().Add(kickTimeout * time.Hour).Unix()),
							}

							_, err := bot.Request(kickConfig)
							if err != nil {
								log.Printf("Failed to kick user: %s", err)
							}
							delete(challengeUserMap, userID) // Remove user from challenge map
						}
					}(newUser.ID, update.Message.Chat.ID)
				}
			}

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
				if userChallenge.Attempts >= 3 {
					_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Incorrect answer"))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
					}
					_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "HFSP"))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
					}
					delete(challengeUserMap, update.Message.From.ID)
					bot.Request(tgbotapi.KickChatMemberConfig{
						ChatMemberConfig: tgbotapi.ChatMemberConfig{
							ChatID: update.Message.Chat.ID,
							UserID: update.Message.From.ID,
						},
					})
				} else {
					_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Incorrect answer. You have %d attempts left.", 3-userChallenge.Attempts)))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
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

// func dbFunc(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if _, err := db.Exec("CREATE TABLE IF NOT EXISTS ticks (tick timestamp)"); err != nil {
// 			c.String(http.StatusInternalServerError,
// 				fmt.Sprintf("Error creating database table: %q", err))
// 			return
// 		}

// 		if _, err := db.Exec("INSERT INTO ticks VALUES (now())"); err != nil {
// 			c.String(http.StatusInternalServerError,
// 				fmt.Sprintf("Error incrementing tick: %q", err))
// 			return
// 		}

// 		rows, err := db.Query("SELECT tick FROM ticks")
// 		if err != nil {
// 			c.String(http.StatusInternalServerError,
// 				fmt.Sprintf("Error reading ticks: %q", err))
// 			return
// 		}

// 		defer rows.Close()
// 		for rows.Next() {
// 			var tick time.Time
// 			if err := rows.Scan(&tick); err != nil {
// 				c.String(http.StatusInternalServerError,
// 					fmt.Sprintf("Error scanning ticks: %q", err))
// 				return
// 			}
// 			c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", tick.String()))
// 		}
// 	}
// }
