package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a 24-hex-character random ID (12 bytes).
// In a real project, you might use ULIDs/UUIDs. We keep stdlib only.
func New() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// For simplicity, panic on failure. Consider returning an error in prod.
		panic("id generation failed: " + err.Error())
	}
	return hex.EncodeToString(buf[:])
}
