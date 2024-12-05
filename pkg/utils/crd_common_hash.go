package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func CalcKeysHash(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	h := md5.New()
	h.Write([]byte(strings.Join(keys, "/")))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)[:8]
}
