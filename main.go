package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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

		if len(update.Message.NewChatMembers) > 0 { // New user(s) join group
			for _, newUser := range update.Message.NewChatMembers {
				if newUser.ID == bot.Self.ID { // Bot
					cmds.WelcomeMessage(config, bot, update)
				} else { // Not bot
					cmds.SendNewUserChallenge(config, dbp, newUser, bot, update, challengeUserMap)
				}
			}
			continue
		}

		if update.Message != nil {

			if cmds.IsUserAdmin(bot, update.Message.Chat.ID, update.Message.From.ID) {
				commandArgs := strings.Fields(update.Message.Text)

				switch commandArgs[0] {
				case "/adduser":
					if len(commandArgs) == 3 {
						cmds.AddUser(commandArgs[1], commandArgs[2], dbp, bot, update)
					} else {
						_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid /adduser format. Use /adduser <paymail> <amount>."))
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
						continue
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
						_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid command. Use /adduser or /leaderboard or /refresh"))
						if err != nil {
							log.Printf("Failed to send message: %s", err)
						}
					}
				}
				continue
			}

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
		}

	}
}
