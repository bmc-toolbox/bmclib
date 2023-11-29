package internal

import (
	"fmt"
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

func ParseInt32(i32 string) (int32, error) {
	var i int32
	_, err := fmt.Sscanf(i32, "%d", &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}
