// Package internal provides shared utilities used across bmclib providers.
package internal

import (
	"unicode"
)

// IsntLetterOrNumber check if the give rune is not a letter nor a number
func IsntLetterOrNumber(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

// IsRoleValid reports whether role is one of the supported user roles.
func IsRoleValid(role string) bool {
	return role == "admin" || role == "user" || role == "operator"
}

// StringInSlice reports whether str is present in the slice sl.
func StringInSlice(str string, sl []string) bool {
	for _, s := range sl {
		if str == s {
			return true
		}
	}
	return false
}
