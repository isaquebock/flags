package snapshot

import (
	"errors"
	"time"
)

var (
	ErrFlagNotFound     = errors.New("flag not found")
	ErrFlagExists       = errors.New("flag already exists")
	ErrTooManyRetries   = errors.New("too many retries")
	ErrInvalidSnapshot  = errors.New("invalid snapshot")
)

type Snapshot struct {
	SchemaVersion int                 `json:"schema_version"`
	ClientID      string              `json:"client_id"`
	GeneratedAt   *time.Time          `json:"generated_at"`
	Flags         map[string]FlagFull `json:"flags"`
}

type FlagFull struct {
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func emptySnapshot(clientID string) *Snapshot {
	return &Snapshot{
		SchemaVersion: 1,
		ClientID:      clientID,
		Flags:         make(map[string]FlagFull),
	}
}

func (s *Snapshot) AddFlag(key string, enabled bool, description string) error {
	if _, exists := s.Flags[key]; exists {
		return ErrFlagExists
	}

	now := time.Now().UTC()
	s.Flags[key] = FlagFull{
		Enabled:     enabled,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return nil
}

func (s *Snapshot) UpdateFlag(key string, enabled *bool, description *string) error {
	flag, exists := s.Flags[key]
	if !exists {
		return ErrFlagNotFound
	}

	if enabled != nil {
		flag.Enabled = *enabled
	}
	if description != nil {
		flag.Description = *description
	}
	flag.UpdatedAt = time.Now().UTC()

	s.Flags[key] = flag
	return nil
}

func (s *Snapshot) RemoveFlag(key string) error {
	if _, exists := s.Flags[key]; !exists {
		return ErrFlagNotFound
	}

	delete(s.Flags, key)
	return nil
}
