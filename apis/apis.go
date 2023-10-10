package apis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
