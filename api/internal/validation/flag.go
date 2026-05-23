package validation

import (
	"regexp"
)

var flagKeyRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{0,63}$`)

func ValidateFlagKey(key string) error {
	if !flagKeyRegex.MatchString(key) {
		return ErrInvalidFlagKey
	}
	return nil
}

func ValidateDescription(desc string) error {
	if len(desc) > 200 {
		return ErrDescriptionTooLong
	}
	return nil
}

func ValidateClientID(id string) error {
	if len(id) == 0 || len(id) > 128 {
		return ErrInvalidClientID
	}
	return nil
}
