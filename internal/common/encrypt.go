package common

import (
	"crypto/md5"
	"encoding/hex"
)

func EncryptStringMD5(pass string) string {
	h := md5.New()
	h.Write([]byte(pass))
	return hex.EncodeToString(h.Sum(nil))
}
