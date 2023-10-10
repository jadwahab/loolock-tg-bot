package cmds

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/tonicpow/go-paymail"
)

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
