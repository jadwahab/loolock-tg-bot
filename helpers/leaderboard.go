package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/tonicpow/go-paymail"
)

func Refresh(dbp *db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	err := RefreshLeaderboard(dbp)
	if err != nil {
		return err
	}

	return nil
}

func RefreshLeaderboard(dbp *db.DBParams) error {
	top100, err := GetTop100Bitcoiners()
	if err != nil {
		return err
	}

	// Prepare data
	var valueStrings []string
	var valueArgs []interface{}
	i := 1
	for _, bitcoiner := range top100 {
		s, err := paymail.ValidateAndSanitisePaymail("1"+bitcoiner.Handle, false) // TODO: harden instead of prepending '1'
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i, i+1, i+2, i+3))
		valueArgs = append(valueArgs, bitcoiner.TotalAmountLocked, s.Address, time.Now(), time.Now())
		i += 4
	}

	sqlStatement := `
		INSERT INTO leaderboard (amount_locked, paymail, created_at, updated_at) 
		VALUES %s
		ON CONFLICT (paymail)
		DO UPDATE SET amount_locked = EXCLUDED.amount_locked, updated_at = NOW();
	`
	sqlStatement = fmt.Sprintf(sqlStatement, strings.Join(valueStrings, ","))
	_, err = dbp.DB.Exec(sqlStatement, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

type Bitcoiner struct {
	Index             int     `json:"index"`
	Handle            string  `json:"handle"`
	CreatedAt         string  `json:"created_at"`
	TotalAmountLocked float64 `json:"totalAmountLocked"`
}

type BitcoinersResponse struct {
	Bitcoiners []Bitcoiner `json:"bitcoiners"`
}

const LeaderboardAPIEndpoint = "https://www.hodlocker.com/api/bitcoiners"

func GetTop100Bitcoiners() ([]Bitcoiner, error) {
	resp, err := http.Get(LeaderboardAPIEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response BitcoinersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	bitcoiners := response.Bitcoiners
	if len(bitcoiners) > 100 {
		bitcoiners = bitcoiners[:100]
	}

	return bitcoiners, nil
}
