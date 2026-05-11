package service

import (
	"strings"
	"unicode"
)

var allowedPlateLatin = map[rune]struct{}{
	'A': {}, 'B': {}, 'C': {}, 'E': {}, 'H': {}, 'K': {}, 'M': {},
	'O': {}, 'P': {}, 'T': {}, 'X': {}, 'Y': {},
}

// russianPlateToLatin — визуальные аналоги кириллицы на номере → латиница с таблички.
var russianPlateToLatin = map[rune]rune{
	'А': 'A', 'В': 'B', 'С': 'C', 'Е': 'E', 'К': 'K',
	'М': 'M', 'Н': 'H', 'О': 'O', 'Р': 'P', 'Т': 'T',
	'У': 'Y', 'Х': 'X',
}

func isPlateLatin(r rune) bool {
	_, ok := allowedPlateLatin[r]
	return ok
}

// NormalizeCarPlate приводит номер к виду на табличке: без пробелов и разделителей,
// кириллические «номерные» буквы → латиница, только цифры и допустимые буквы.
// При любой недопустимой букве или символе возвращает пустую строку.
func NormalizeCarPlate(plate string) string {
	plate = strings.TrimSpace(plate)
	if plate == "" {
		return ""
	}

	var result strings.Builder
	for _, r := range plate {
		if r == ' ' || r == '-' || r == '.' || r == '\u00a0' || r == '_' {
			continue
		}

		u := unicode.ToUpper(r)
		if mapped, ok := russianPlateToLatin[u]; ok {
			result.WriteRune(mapped)
			continue
		}
		if u >= '0' && u <= '9' {
			result.WriteRune(u)
			continue
		}
		if u >= 'A' && u <= 'Z' {
			if !isPlateLatin(u) {
				return ""
			}
			result.WriteRune(u)
			continue
		}
		return ""
	}

	return result.String()
}
