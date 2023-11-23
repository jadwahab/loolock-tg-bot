package helpers_test

import (
	"testing"

	"github.com/jadwahab/loolock-tg-bot/helpers"
)

func TestVerifyBSM(t *testing.T) {
	t.Run("Valid BSM verification", func(t *testing.T) {
		pubkey := "036dc74d02a42226b3756b2079ff66d5a8bec9282665ceac090482e111ecf7d03b"
		sig := "H1hqhMW5afsJvkyppztL8IqsNSIXq26iJvfhGx2aq3cJGEN7Sbxy4LUE27qJH0FuwIrqSoYoA72xWOUGJ7ojQKA="
		message := "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc|@|6837122638"
		valid := helpers.VerifyBSM(pubkey, sig, message)
		if !valid {
			t.Errorf("Expected BSM verification to be valid")
		}
	})

	t.Run("Invalid BSM verification - incorrect pubkey", func(t *testing.T) {
		invalidPubkey := "034356f9616d87858e783c96683b05b8e3586458ed53717598ab804198b3340c65"
		sig := "H1hqhMW5afsJvkyppztL8IqsNSIXq26iJvfhGx2aq3cJGEN7Sbxy4LUE27qJH0FuwIrqSoYoA72xWOUGJ7ojQKA="
		message := "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc|@|6837122638"
		invalid := helpers.VerifyBSM(invalidPubkey, sig, message)
		if invalid {
			t.Errorf("Expected BSM verification to be invalid for invalid pubkey")
		}
	})

	t.Run("Invalid BSM verification - incorrect signature", func(t *testing.T) {
		pubkey := "036dc74d02a42226b3756b2079ff66d5a8bec9282665ceac090482e111ecf7d03b"
		invalidSig := "IOMmw5Ed6gTotpjf/MDVReDO72LuoIkHMPr++qKhbcJ1Zh6VKKK/Te0oaQ41EBtmpO6KzTQDa7+Mal0rADIR1l8="
		message := "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc|@|6837122638"
		invalid := helpers.VerifyBSM(pubkey, invalidSig, message)
		if invalid {
			t.Errorf("Expected BSM verification to be invalid for invalid signature")
		}
	})

	t.Run("Invalid BSM verification - incorrect message", func(t *testing.T) {
		pubkey := "036dc74d02a42226b3756b2079ff66d5a8bec9282665ceac090482e111ecf7d03b"
		sig := "H1hqhMW5afsJvkyppztL8IqsNSIXq26iJvfhGx2aq3cJGEN7Sbxy4LUE27qJH0FuwIrqSoYoA72xWOUGJ7ojQKA="
		invalidMessage := "invalidMessage"
		invalid := helpers.VerifyBSM(pubkey, sig, invalidMessage)
		if invalid {
			t.Errorf("Expected BSM verification to be invalid for invalid message")
		}
	})

	t.Run("Valid BSM verification - Panda wallet", func(t *testing.T) {
		pubkey := "0356dd676b7d29bc665f0a0ca516caeaf80a3719fd7fbcb93d8a80e81f0f1ab5ad"
		sig := "H6SyY9OYYqMWRfnQJp7pu01ZGJqktthWv16x68V9mefYNWU25p9jjhT3jJXLM102lg5wFN29Kqwo3qcZw36/Hk8="
		message := "17DqbMhfHzLGjYqmiLAjhzAzKf3f1sK9Rc|@|6837122638"
		invalid := helpers.VerifyBSM(pubkey, sig, message)
		if invalid {
			t.Errorf("Expected BSM verification to be invalid for invalid message")
		}
	})
}
