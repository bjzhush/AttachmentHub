package publicid

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	Length = 12
)

var (
	ErrInvalidID = errors.New("invalid public id")
)

func FromAttachment(id int64, fileSHA string, storedName string, attempt int) (string, error) {
	if id <= 0 {
		return "", ErrInvalidID
	}

	seed := fmt.Sprintf("%d:%s:%s:%d", id, strings.TrimSpace(fileSHA), strings.TrimSpace(storedName), attempt)
	sum := sha256.Sum256([]byte(seed))
	hash := strings.ToUpper(hex.EncodeToString(sum[:]))
	if len(hash) < Length {
		return "", ErrInvalidID
	}
	return hash[:Length], nil
}

func Normalize(id string) (string, error) {
	cleaned := strings.ToUpper(strings.TrimSpace(id))
	if len(cleaned) != Length {
		return "", ErrInvalidID
	}

	for i := 0; i < len(cleaned); i++ {
		c := cleaned[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return "", ErrInvalidID
		}
	}

	return cleaned, nil
}
