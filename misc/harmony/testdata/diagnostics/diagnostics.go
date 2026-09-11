package diagnostics

import "encoding/hex"

func RoundTrip(data []byte) ([]byte, error) {
	return hex.DecodeString(hex.EncodeToString(data))
}
