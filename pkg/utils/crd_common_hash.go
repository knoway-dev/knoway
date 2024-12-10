package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func CalcKeysHash(keys []string) string {
	if len(keys) == 0 {
		return ""
	}

	h := sha256.New()
	h.Write([]byte(strings.Join(keys, "/")))
	bs := h.Sum(nil)

	return hex.EncodeToString(bs)[:8]
}
