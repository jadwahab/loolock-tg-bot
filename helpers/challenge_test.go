package helpers_test

import (
	"reflect"
	"testing"

	"github.com/jadwahab/loolock-tg-bot/helpers"
)

func TestIsValidChallengeResponse(t *testing.T) {
	// Valid challenge response
	validMessage := `challenge
	signature
	publicKey
	paymailAddress`
	expectedValidResponse := &helpers.ChallengeResponse{
		Challenge: "challenge",
		Signature: "signature",
		PubKey:    "publicKey",
		Paymail:   "paymailAddress",
	}
	validResponse, valid := helpers.IsValidChallengeResponse(validMessage)
	if !valid || !reflect.DeepEqual(validResponse, expectedValidResponse) {
		t.Errorf("Expected valid response: %v, got: %v", expectedValidResponse, validResponse)
	}

	// Invalid challenge response - less than 3 lines
	invalidMessage1 := `challenge
	signature`
	_, valid = helpers.IsValidChallengeResponse(invalidMessage1)
	if valid {
		t.Errorf("Expected invalid response for message: %s", invalidMessage1)
	}

	// Invalid challenge response - more than 4 lines
	invalidMessage2 := `challenge
	signature
	publicKey
	paymailAddress
	extraLine`
	_, valid = helpers.IsValidChallengeResponse(invalidMessage2)
	if valid {
		t.Errorf("Expected invalid response for message: %s", invalidMessage2)
	}

	// Invalid challenge response - invalid signature
	invalidMessage3 := `challenge
	invalidSignature
	publicKey
	paymailAddress`
	_, valid = helpers.IsValidChallengeResponse(invalidMessage3)
	if valid {
		t.Errorf("Expected invalid response for message: %s", invalidMessage3)
	}

	// Invalid challenge response - invalid public key length
	invalidMessage4 := `challenge
	signature
	invalidPublicKey
	paymailAddress`
	_, valid = helpers.IsValidChallengeResponse(invalidMessage4)
	if valid {
		t.Errorf("Expected invalid response for message: %s", invalidMessage4)
	}

	// Invalid challenge response - invalid paymail address
	invalidMessage5 := `challenge
	signature
	publicKey
	invalidPaymailAddress`
	_, valid = helpers.IsValidChallengeResponse(invalidMessage5)
	if valid {
		t.Errorf("Expected invalid response for message: %s", invalidMessage5)
	}
}
