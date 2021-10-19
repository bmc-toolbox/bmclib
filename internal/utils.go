package internal

import (
	"errors"
	"unicode"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// IsntLetterOrNumber check if the give rune is not a letter nor a number
func IsntLetterOrNumber(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

func ErrStringOrEmpty(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func IsRoleValid(role string) bool {
	return role == "admin" || role == "user"
}

func ValidateUserConfig(cfgUsers []*cfgresources.User) (err error) {
	for _, cfgUser := range cfgUsers {
		if cfgUser.Name == "" {
			msg := "User resource expects parameter: Name."
			return errors.New(msg)
		}

		if cfgUser.Password == "" {
			msg := "User resource expects parameter: Password."
			return errors.New(msg)
		}

		if !IsRoleValid(cfgUser.Role) {
			msg := "Parameter \"Role\" is one of ['admin', 'user']. You sent " + cfgUser.Role
			return errors.New(msg + "!")
		}
	}

	return nil
}

func StringInSlice(str string, sl []string) bool {
	for _, s := range sl {
		if str == s {
			return true
		}
	}
	return false
}
