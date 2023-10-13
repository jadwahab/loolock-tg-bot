package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/cmds"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
	_ "github.com/lib/pq"
)

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
	bot.Debug = config.BotDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsPrivate() {
			cmds.HandleDMs(config, dbp, bot, update)
			continue
		}

		if update.Message.Chat.ID == config.Top100ChatID { // TOP 100 CHAT

			if len(update.Message.NewChatMembers) > 0 { // New user(s) join group
				for _, newUser := range update.Message.NewChatMembers {

					if newUser.ID == bot.Self.ID { // Bot
						cmds.WelcomeMessage(config, bot, update)

					} else { // Not bot
						log.Printf("New user joined ChatID: %d, UserID: %d, UserName: %s", update.Message.Chat.ID, newUser.ID, newUser.UserName)
						err = dbp.AddUserToGroupChatDB(update.Message.Chat.ID, update.Message.From.ID, update.Message.From.UserName)
						if err != nil {
							log.Printf("Failed to add user to group table: %s", err)
						}

						// kick if not on leaderboard
						lbes, err := dbp.GetLeaderboard(100) // TODO: top 100 for now
						if err != nil {
							log.Printf("Failed to get leaderboard: %s", err)
						}
						exists := helpers.UserExistsInLeaderboard(lbes, newUser.UserName)
						if !exists && newUser.ID != bot.Self.ID {
							helpers.KickUser(bot, &helpers.KickArgs{
								ChatID:       update.Message.Chat.ID,
								UserID:       newUser.ID,
								KickDuration: time.Duration(config.KickDuration),
								UserName:     newUser.UserName,
								DBP:          dbp,
							})
							if err != nil {
								log.Println("Error kicking member:", err)
							} else {
								log.Printf("Kicked %s (%d)\n", newUser.UserName, newUser.ID)
							}
						}

					}
				}
				continue
			}

			if update.Message.LeftChatMember != nil { // User leaves group
				leaver := update.Message.LeftChatMember
				err := dbp.RemoveUserFromGroupChatDB(update.Message.Chat.ID, leaver.ID)
				if err != nil {
					log.Printf("Failed to remove user %d from DB: %s", leaver.ID, err)
				}
			}

			if update.Message != nil {

				// check user exists in group_chat_users table and add if not
				userExists, err := dbp.UserExists(update.Message.Chat.ID, update.Message.From.ID)
				if err != nil {
					log.Printf("Failed to check if user exists: %s", err)
				}
				if !userExists {
					err = dbp.AddUserToGroupChatDB(update.Message.Chat.ID, update.Message.From.ID, update.Message.From.UserName)
					if err != nil {
						log.Printf("Failed to add user to group table: %s", err)
					}
				}

				if cmds.IsUserAdmin(bot, update.Message.Chat.ID, update.Message.From.ID) && update.Message.Text != "" {
					cmds.AdminCommand(update.Message.Text, dbp, bot, update)
				} else {
					user := update.Message.From
					// kick if not on leaderboard
					lbes, err := dbp.GetLeaderboard(100) // TODO: top 100 for now
					if err != nil {
						log.Printf("Failed to get leaderboard: %s", err)
					}
					exists := helpers.UserExistsInLeaderboard(lbes, user.UserName)
					if !exists && user.ID != bot.Self.ID {
						helpers.KickUser(bot, &helpers.KickArgs{
							ChatID:       update.Message.Chat.ID,
							UserID:       user.ID,
							KickDuration: time.Duration(config.KickDuration),
							UserName:     user.UserName,
							DBP:          dbp,
						})
						if err != nil {
							log.Println("Error kicking member:", err)
						} else {
							log.Printf("Kicked %s (%d)\n", user.UserName, user.ID)
						}
					}
				}
			}
			continue
		}

	}

	// Create a ticker and call the refresh function periodically
	ticker := time.NewTicker(time.Duration(config.RefreshPeriod) * time.Hour)
	go func() {
		for range ticker.C {
			err := helpers.RefreshLeaderboard(dbp)
			if err != nil {
				fmt.Println("Error refreshing leaderboard:", err)
			} else {
				fmt.Println("Leaderboard refreshed!")
			}
		}
	}()

}
