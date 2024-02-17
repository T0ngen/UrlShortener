package generateapi

import (
	"crypto/rand"
	"encoding/hex"
)




func GenerateAPIKey() (string, error) {
    bytes := make([]byte, 12)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    key := hex.EncodeToString(bytes)
    return key, nil
}