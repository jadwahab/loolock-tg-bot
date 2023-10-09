package helpers

import "github.com/bitcoinschema/go-bitcoin/v2"

/*
Examples:
	const pubkey = "02320bd4c9d15ded4645828c2e9e8788630c7cc253e8572a586c14843ad14c5ae6"
	const sig = "IJDiGEdovFRf/U2WtJ6WJz59eBupAuZDJKXe0/O1aJvAYSF4xGW2ZllIUX6cybm51Uv5f1GRID41v7bcIVr4Jrk="
	const msg = "1RELAYTEST|test"
*/

func VerifyBSM(pubkey, sig, message string) bool {
	// Get an address from private key
	// the compressed flag must match the flag provided during signing
	address, err := bitcoin.GetAddressFromPubKeyString(pubkey, true)
	if err != nil {
		return false
	}

	// Verify the signature
	if err = bitcoin.VerifyMessage(address.AddressString, sig, message); err != nil {
		return false
	} else {
		return true
	}
}
