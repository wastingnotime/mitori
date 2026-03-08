package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func newID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + "_" + time.Now().Format("20060102150405") + "_" + hex.EncodeToString(b)
}
