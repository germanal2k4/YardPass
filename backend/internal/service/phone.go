package service

import (
	"errors"
	"strings"
	"unicode"
)

// ErrInvalidResidentPhone is returned when a phone string cannot be normalized to a Russian mobile number.
var ErrInvalidResidentPhone = errors.New("Укажите корректный номер телефона (российский мобильный, 10 или 11 цифр).")

// NormalizeResidentPhone normalizes Russian mobile numbers to the form +7XXXXXXXXXX. Empty input yields nil.
func NormalizeResidentPhone(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil, nil
	}
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) == 11 && d[0] == '8' {
		d = "7" + d[1:]
	}
	if len(d) == 10 && d[0] == '9' {
		d = "7" + d
	}
	if len(d) != 11 || d[0] != '7' {
		return nil, ErrInvalidResidentPhone
	}
	out := "+" + d
	return &out, nil
}
