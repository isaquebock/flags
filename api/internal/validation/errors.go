package validation

import "errors"

var (
	ErrInvalidFlagKey      = errors.New("invalid flag key: must start with lowercase letter or digit, contain only lowercase letters, digits, hyphens, underscores, and be at most 64 characters")
	ErrDescriptionTooLong  = errors.New("description must be at most 200 characters")
	ErrInvalidClientID     = errors.New("invalid client ID: must be between 1 and 128 characters")
)
