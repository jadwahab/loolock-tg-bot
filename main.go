package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

			challenge, paymail, sig, valid := helpers.IsValidChallengeResponse(update.Message.Text)
			if update.Message.Text != "" && valid {
				// If challenge is valid, send them the group link
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You're validated! Join the group by clicking [here](GROUP_LINK_HERE)"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}

				cmds.HandleChallengeResponse(dbp, bot, update, challenge, paymail, sig)
			}

			switch update.Message.Text {

			case "/verify":
				userEntry, err := dbp.GetUserByTelegramUsername(update.Message.From.UserName)
				if err != nil {
					log.Printf("Database error while fetching user: %v", err)
				} else if userEntry != nil { // user found in db
					// fmt.Println("Found entry:", userEntry.TelegramUsername)
					if userEntry.IsVerified {
						_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You are already verified!"))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
						continue
					}
				}
				cmds.SendNewUserChallenge(*update.Message.From, bot, update.Message.Chat.ID)

			case "/top100":
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("After you have verified yourself using the /verify command, "+
						"join the group by clicking here: %s", config.GroupLink)))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}

			default:
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					"Invalid command. Use /verify or /top100\n\n"+
						"/verify - Verify your identity\n"+
						"/top100 - Get access to the TOP 100 lockers group"))
				if err != nil {
					log.Printf("Failed to send message: %s", err)
				}
			}

			continue
		}

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

					// TODO: check if user is verified else kick them

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
				continue
			}
		}

	}

	// TODO: fix periodic refresh
	// // Create a ticker and call the refresh function periodically
	// ticker := time.NewTicker(time.Duration(config.RefreshPeriod) * time.Hour)
	// go func() {
	// 	for range ticker.C {
	// 		err := helpers.Refresh(config, dbp, bot)
	// 		if err != nil {
	// 			fmt.Println("Error refreshing leaderboard:", err)
	// 		} else {
	// 			fmt.Println("Leaderboard refreshed!")
	// 		}
	// 	}
	// }()

}
