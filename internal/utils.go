package internal

import (
	"unicode"
)

// IsntLetterOrNumber check if the give rune is not a letter nor a number
func IsntLetterOrNumber(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

func IsRoleValid(role string) bool {
	return role == "admin" || role == "user" || role == "operator"
}

func StringInSlice(str string, sl []string) bool {
	for _, s := range sl {
		if str == s {
			return true
		}
	}
	return false
}
