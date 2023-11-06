package apis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type LockedBitcoiner struct {
	Index             int     `json:"index"`
	Handle            string  `json:"handle"`
	CreatedAt         string  `json:"created_at"`
	TotalAmountLocked float64 `json:"totalAmountLocked"`
}

type LockedBitcoinersResponse struct {
	Bitcoiners []LockedBitcoiner `json:"rankedBitcoiners"`
}

const LockedLeaderboardAPIEndpoint = "https://www.hodlocker.com/api/bitcoiners"

func GetBitcoinersLocked() ([]LockedBitcoiner, error) {
	resp, err := http.Get(LockedLeaderboardAPIEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response LockedBitcoinersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Bitcoiners, nil
}

type LikedBitcoiner struct {
	Index                    int     `json:"index"`
	Handle                   string  `json:"handle"`
	CreatedAt                string  `json:"created_at"`
	TotalLockLikedFromOthers float64 `json:"totalLockLikedFromOthers"`
}

type LikedBitcoinersResponse struct {
	Bitcoiners []LikedBitcoiner `json:"rankedBitcoiners"`
}

const LikedLeaderboardAPIEndpoint = "https://www.hodlocker.com/api/bitcoiners/liked"

func GetBitcoinersLiked() ([]LikedBitcoiner, error) {
	resp, err := http.Get(LikedLeaderboardAPIEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response LikedBitcoinersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Bitcoiners, nil
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
