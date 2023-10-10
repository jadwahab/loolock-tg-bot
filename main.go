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

// Keeps track of user ID and their challenge string + number of attempts
var challengeUserMap = make(map[int64]*cmds.UserChallenge)

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

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsPrivate() { // Skip private chats
			continue
		}

		if len(update.Message.NewChatMembers) > 0 { // New user(s) join group
			for _, newUser := range update.Message.NewChatMembers {
				if newUser.ID == bot.Self.ID { // Bot
					cmds.WelcomeMessage(config, bot, update)
					// Get list of all admins in the chat
					chatAdminConfig := tgbotapi.ChatAdministratorsConfig{
						ChatConfig: tgbotapi.ChatConfig{
							ChatID: update.Message.Chat.ID,
						},
					}

					admins, err := bot.GetChatAdministrators(chatAdminConfig)
					if err != nil {
						log.Printf("Failed to get chat admins: %s", err)
						continue
					}

					// Loop through all admins and send them a challenge
					for _, admin := range admins {
						// Skip if the admin is the bot itself
						if admin.User.ID == bot.Self.ID {
							continue
						}

						// Send challenge to the admin
						cmds.SendNewUserChallenge(config, dbp, *admin.User, bot, update, challengeUserMap)
					}

				} else { // Not bot
					cmds.SendNewUserChallenge(config, dbp, newUser, bot, update, challengeUserMap)
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

			if _, exists := challengeUserMap[update.Message.From.ID]; exists { // User sent challenge response
				paymail, sig, valid := helpers.IsValidChallengeResponse(update.Message.Text)
				if update.Message.Text != "" && valid {
					cmds.HandleChallengeResponse(config, dbp, bot, update, challengeUserMap, paymail, sig)
				} else {
					_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid format, try again."))
					if err != nil {
						log.Printf("Failed to send message: %s", err)
					}
				}
			}

			if cmds.IsUserAdmin(bot, update.Message.Chat.ID, update.Message.From.ID) && update.Message.Text != "" {
				cmds.AdminCommand(update.Message.Text, dbp, bot, update)
				continue
			}
		}

	}
}
