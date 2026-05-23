package snapshot

import "context"

type Store interface {
	Get(ctx context.Context, clientID string) (*Snapshot, error)
	Mutate(ctx context.Context, clientID string, fn func(*Snapshot) error) (*Snapshot, error)
}
