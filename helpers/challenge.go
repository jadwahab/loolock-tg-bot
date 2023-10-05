package helpers

import (
	"encoding/base64"
	"strings"
)

func IsValidChallengeResponse(message string) (string, string, bool) {
	lines := strings.Split(strings.TrimSpace(message), "\n")

	if len(lines) != 2 {
		return "", "", false
	}

	paymail := strings.TrimSpace(lines[0])
	signature := strings.TrimSpace(lines[1])

	// Basic check for paymail format
	if !strings.Contains(paymail, "@") {
		return "", "", false
	}

	// Check if signature is a valid Base64 string
	_, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return "", "", false
	}

	return paymail, signature, true
}
