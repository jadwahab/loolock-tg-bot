package helpers

import (
	"encoding/base64"
	"strings"

	"github.com/tonicpow/go-paymail"
)

func IsValidChallengeResponse(message string) (string, string, bool) {
	lines := strings.Split(strings.TrimSpace(message), "\n")

	if len(lines) != 2 {
		return "", "", false
	}

	paymailAddress := strings.TrimSpace(lines[0])
	signature := strings.TrimSpace(lines[1])

	s, err := paymail.ValidateAndSanitisePaymail(paymailAddress, false)
	if err != nil {
		return "", "", false
	}

	// Check if signature is a valid Base64 string
	_, err = base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return "", "", false
	}

	return s.Address, signature, true
}
