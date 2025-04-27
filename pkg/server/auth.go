package server

import (
	"crypto/sha256"
	"encoding/hex"
)

func sha256Hash(payload []byte) string {
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}
