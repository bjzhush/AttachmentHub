package publicid

import (
	"errors"
	"strings"
)

const (
	Length = 8
	base   = 36
)

var (
	ErrInvalidID      = errors.New("invalid public id")
	ErrSourceTooLarge = errors.New("source id exceeds 8-char base36 range")
	alphabet          = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func FromInt64(id int64) (string, error) {
	if id <= 0 {
		return "", ErrInvalidID
	}

	var buf [Length]byte
	for i := Length - 1; i >= 0; i-- {
		buf[i] = alphabet[id%base]
		id /= base
	}

	if id > 0 {
		return "", ErrSourceTooLarge
	}

	return string(buf[:]), nil
}

func Normalize(id string) (string, error) {
	cleaned := strings.ToUpper(strings.TrimSpace(id))
	if len(cleaned) != Length {
		return "", ErrInvalidID
	}

	for i := 0; i < len(cleaned); i++ {
		c := cleaned[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			return "", ErrInvalidID
		}
	}

	return cleaned, nil
}
