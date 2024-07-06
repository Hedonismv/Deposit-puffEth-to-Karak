package formatter

import "strings"

// PrivateKeyToHex This function converts default private key with 0x... prefix to secp256k1 private key
func PrivateKeyToHex(privateKey string) string {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
		return privateKey

	}
	return privateKey
}
