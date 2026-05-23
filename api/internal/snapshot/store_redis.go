package snapshot

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	client     *redis.Client
	maxRetries int
}

func NewRedisStore(client *redis.Client, maxRetries int) Store {
	return &redisStore{
		client:     client,
		maxRetries: maxRetries,
	}
}

func (s *redisStore) Get(ctx context.Context, clientID string) (*Snapshot, error) {
	key := "snapshot:" + clientID

	raw, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return emptySnapshot(clientID), nil
	}
	if err != nil {
		return nil, err
	}

	snap := emptySnapshot(clientID)
	if err := json.Unmarshal([]byte(raw), snap); err != nil {
		return nil, err
	}

	return snap, nil
}

func (s *redisStore) Mutate(
	ctx context.Context,
	clientID string,
	fn func(*Snapshot) error,
) (*Snapshot, error) {
	key := "snapshot:" + clientID
	var result *Snapshot

	for attempt := 0; attempt < s.maxRetries; attempt++ {
		err := s.client.Watch(ctx, func(tx *redis.Tx) error {
			// Read current snapshot
			raw, err := tx.Get(ctx, key).Result()
			snap := emptySnapshot(clientID)

			if err == nil {
				if err := json.Unmarshal([]byte(raw), snap); err != nil {
					return err
				}
			} else if !errors.Is(err, redis.Nil) {
				return err
			}

			// Apply mutation
			if err := fn(snap); err != nil {
				return err
			}

			// Set generated_at
			now := time.Now().UTC()
			snap.GeneratedAt = &now

			// Serialize and persist
			payload, err := json.Marshal(snap)
			if err != nil {
				return err
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				return pipe.Set(ctx, key, payload, 0).Err()
			})

			result = snap
			return err
		}, key)

		if err == nil {
			return result, nil
		}

		if errors.Is(err, redis.TxFailedErr) {
			continue // Retry
		}

		return nil, err
	}

	return nil, ErrTooManyRetries
}
