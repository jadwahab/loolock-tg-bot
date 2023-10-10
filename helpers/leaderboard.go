package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/db"
)

func Refresh(dbp db.DBParams, bot *tgbotapi.BotAPI, update tgbotapi.Update) {

}

type Bitcoiner struct {
	Index             int     `json:"index"`
	Handle            string  `json:"handle"`
	CreatedAt         string  `json:"created_at"`
	TotalAmountLocked float64 `json:"totalAmountLocked"`
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

	var bitcoiners []Bitcoiner
	if err := json.Unmarshal(body, &bitcoiners); err != nil {
		return nil, err
	}

	// not needed since already ordered in api response
	// sort.Slice(bitcoiners, func(i, j int) bool {
	// 	return bitcoiners[i].TotalAmountLocked > bitcoiners[j].TotalAmountLocked
	// })

	if len(bitcoiners) > 100 {
		bitcoiners = bitcoiners[:100]
	}

	return bitcoiners, nil
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
