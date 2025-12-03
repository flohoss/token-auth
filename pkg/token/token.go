package token

import (
	"crypto/sha256"
	"encoding/hex"
)

type Token struct {
	AllowedTokens []string
}

func New(allowedTokens []string) *Token {
	return &Token{
		AllowedTokens: allowedTokens,
	}
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (t *Token) Valid(value string, isHash bool) bool {
	if len(t.AllowedTokens) == 0 {
		return false
	}

	providedHash := value
	if !isHash {
		providedHash = HashToken(value)
	}

	for _, allowedToken := range t.AllowedTokens {
		allowedHash := HashToken(allowedToken)
		if providedHash == allowedHash {
			return true
		}
	}
	return false
}
