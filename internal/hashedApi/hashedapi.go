package hashedapi

import (
	"crypto/md5"
	"encoding/hex"
)



func HashApi(token string) string{
	hash := md5.Sum([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])
	return hashedToken
}