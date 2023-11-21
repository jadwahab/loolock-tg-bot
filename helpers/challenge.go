package helpers

import (
	"encoding/base64"
	"strings"

	"github.com/tonicpow/go-paymail"
)

type ChallengeResponse struct {
	Challenge string
	Signature string
	PubKey    string
	Paymail   string
}

func IsValidChallengeResponse(message string) (*ChallengeResponse, bool) {
	lines := strings.Split(strings.TrimSpace(message), "\n")

	if len(lines) < 3 || len(lines) > 4 {
		return nil, false
	}

	challenge := strings.TrimSpace(lines[0])
	signature := strings.TrimSpace(lines[1])
	// Check if signature is a valid Base64 string
	_, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, false
	}
	pubKey := strings.TrimSpace(lines[2])
	if len(pubKey) != 66 {
		return nil, false
	}

	var paymailAddress *paymail.SanitisedPaymail
	if len(lines) == 4 {
		paymailAddr := strings.TrimSpace(lines[3])
		paymailAddress, err = paymail.ValidateAndSanitisePaymail(paymailAddr, false)
		if err != nil {
			return nil, false
		}
	} else {
		paymailAddress = &paymail.SanitisedPaymail{
			Address: "",
		}
	}

	return &ChallengeResponse{
		Challenge: challenge,
		Signature: signature,
		PubKey:    pubKey,
		Paymail:   paymailAddress.Address,
	}, true
}
