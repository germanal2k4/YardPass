package service

import (
	"fmt"
	"testing"
)

func TestNormalizeCarPlate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"  ", ""},
		{"A123BC77", "A123BC77"},
		{"a123bc77", "A123BC77"},
		{"А123ВС77", "A123BC77"},
		{"а 123 вс 77", "A123BC77"},
		{"A-123.BC-777", "A123BC777"},
		{"A123BC", "A123BC"},
		{"X999XX99", "X999XX99"},
		{"Х999ХХ99", "X999XX99"},
		{"У777УУ77", "Y777YY77"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%q", tc.in), func(t *testing.T) {
			t.Parallel()
			got := NormalizeCarPlate(tc.in)
			if got != tc.want {
				t.Fatalf("NormalizeCarPlate(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeCarPlate_InvalidLetters(t *testing.T) {
	t.Parallel()
	invalid := []string{
		"D123DD77",
		"А123ДС77",
		"G123GG77",
		"Q123QQ77",
		"A123BC7Z",
		"Щ123BC77",
	}
	for _, s := range invalid {
		t.Run(fmt.Sprintf("%q", s), func(t *testing.T) {
			t.Parallel()
			if got := NormalizeCarPlate(s); got != "" {
				t.Fatalf("expected empty, got %q", got)
			}
		})
	}
}
