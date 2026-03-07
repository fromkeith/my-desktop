package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"hash/fnv"
)

func RandB64(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func Sha256Bytes(src string) []byte {
	hash := sha256.New()
	hash.Write([]byte(src))
	return hash.Sum(nil)
}

func HashToInt64(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	u := h.Sum64()
	// PostgreSQL advisory locks take signed BIGINT, so convert safely.
	return int64(u % (1 << 63))
}

func SetDiff(old, new []string) (added, removed []string) {
	newSet := make(map[string]bool)
	for _, item := range new {
		newSet[item] = true
	}

	existing := make(map[string]bool)
	for _, item := range old {
		existing[item] = true
		// this item was removed
		if _, exists := newSet[item]; !exists {
			removed = append(removed, item)
		}
	}

	// find new items
	for _, item := range new {
		if _, exists := existing[item]; !exists {
			added = append(added, item)
		}
	}
	return added, removed
}
