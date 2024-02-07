package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/cmds"
	"github.com/jadwahab/loolock-tg-bot/cmds/admin"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
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

	dbp := &db.DBParams{
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
	bot.Debug = cfg.BotDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Create a ticker and refresh leaderboard periodically
	ticker := time.NewTicker(time.Duration(cfg.RefreshPeriod) * time.Hour)
	go func() {
		for range ticker.C {
			err := helpers.RefreshLeaderboard(dbp)
			if err != nil {
				log.Println("Error refreshing leaderboard:", err)
			} else {
				log.Println("Leaderboard refreshed!")
			}
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsPrivate() {
			cmds.HandleDMs(cfg, dbp, bot, update)
			continue
		}

		if update.Message.Chat.ID == cfg.Groups[config.Top100].ChatID { // TOP 100 CHAT

			if len(update.Message.NewChatMembers) > 0 { // New user(s) join group
				for _, newUser := range update.Message.NewChatMembers {

					if newUser.ID == bot.Self.ID { // Bot
						cmds.WelcomeMessage(cfg, bot, update)

					} else { // Not bot
						err := helpers.HandleNewUser(dbp, cfg, bot, newUser)
						if err != nil {
							err = admin.Notify(bot, "Failed to handle new user: "+err.Error())
							if err != nil {
								log.Printf("Failed to notify admin: %s", err)
							}
						}
					}
				}
				continue
			}

			if update.Message.LeftChatMember != nil { // User leaves group
				leaver := update.Message.LeftChatMember
				err := dbp.UpdateUserLeftAt(update.Message.Chat.ID, leaver.ID)
				if err != nil {
					err = admin.Notify(bot, fmt.Sprintf("Failed to update left_at for user %d in DB: %s", leaver.ID, err))
					if err != nil {
						log.Printf("Failed to notify admin: %s", err)
					}
				}
				log.Printf("User left the group with ID: %d, UserName: %s", leaver.ID, leaver.UserName)
			}

			if update.Message != nil { // Message sent on group
				log.Printf("Message from %d, %s: %s", update.Message.From.ID, update.Message.From.UserName, update.Message.Text)

				// // TODO: remove?
				// // check user exists in group_chat_users table and add if not
				// userExists, err := dbp.UserExists(update.Message.Chat.ID, update.Message.From.ID)
				// if err != nil {
				// 	log.Printf("Failed to check if user exists: %s", err)
				// }
				// if !userExists {
				// 	user := update.Message.From
				// 	err = dbp.AddUserToGroupChatDB(update.Message.Chat.ID, user.ID, user.UserName, strings.TrimSpace(user.FirstName+" "+user.LastName), true)
				// 	if err != nil {
				// 		log.Printf("Failed to add user to group table: %s", err)
				// 	}
				// }

				if cmds.IsUserAdmin(bot, update.Message.Chat.ID, update.Message.From.ID) && update.Message.Text != "" {
					cmds.AdminCommand(update.Message.Text, dbp, bot, update)
				}
			}
			continue
		} // end TOP 100 CHAT

	}

}
