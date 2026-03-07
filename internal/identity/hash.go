package identity

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashLength is the number of hex characters in a test ID hash.
// 16 hex chars = 64 bits, providing practical uniqueness for test identity.
const HashLength = 16

// GenerateID produces a deterministic test ID from a canonical identity string.
// The ID is a truncated SHA-256 hash, hex-encoded.
//
// Properties:
//   - deterministic: same canonical identity always produces the same ID
//   - stable: independent of traversal order or runtime state
//   - compact: 16 hex characters
func GenerateID(canonical string) string {
	h := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(h[:])[:HashLength]
}
