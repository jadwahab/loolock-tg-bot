package helpers

import (
	"encoding/base64"
	"strings"

	"github.com/tonicpow/go-paymail"
)

func IsValidChallengeResponse(message string) (string, string, string, bool) {
	lines := strings.Split(strings.TrimSpace(message), "\n")

	if len(lines) != 3 {
		return "", "", "", false
	}

	challenge := strings.TrimSpace(lines[0])
	paymailAddress := strings.TrimSpace(lines[1])
	signature := strings.TrimSpace(lines[2])

	s, err := paymail.ValidateAndSanitisePaymail(paymailAddress, false)
	if err != nil {
		return "", "", "", false
	}

	// Check if signature is a valid Base64 string
	_, err = base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return "", "", "", false
	}

	return challenge, s.Address, signature, true
}
