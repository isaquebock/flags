package snapshot

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestMutateCreateFromEmpty(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	store := NewRedisStore(client, 3)

	snap, err := store.Mutate(context.Background(), "test-client", func(s *Snapshot) error {
		return s.AddFlag("feature-1", true, "Test feature")
	})

	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	if snap == nil {
		t.Fatal("snap is nil")
	}

	flag, exists := snap.Flags["feature-1"]
	if !exists {
		t.Fatal("flag not found")
	}

	if !flag.Enabled {
		t.Fatal("flag should be enabled")
	}

	if flag.Description != "Test feature" {
		t.Fatalf("expected 'Test feature', got %q", flag.Description)
	}

	if snap.GeneratedAt == nil {
		t.Fatal("GeneratedAt should be set")
	}
}

func TestMutateDuplicate(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	store := NewRedisStore(client, 3)

	// Create first flag
	_, err := store.Mutate(context.Background(), "test-client", func(s *Snapshot) error {
		return s.AddFlag("feature-1", true, "Test feature")
	})
	if err != nil {
		t.Fatalf("First Mutate failed: %v", err)
	}

	// Try to create duplicate
	_, err = store.Mutate(context.Background(), "test-client", func(s *Snapshot) error {
		return s.AddFlag("feature-1", false, "Another feature")
	})

	if err != ErrFlagExists {
		t.Fatalf("expected ErrFlagExists, got %v", err)
	}
}

func TestMutatePreservesBetweenCalls(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	store := NewRedisStore(client, 3)

	// First mutation
	snap1, err := store.Mutate(context.Background(), "test-client", func(s *Snapshot) error {
		return s.AddFlag("feature-1", true, "First")
	})
	if err != nil {
		t.Fatalf("First Mutate failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure different timestamp

	// Second mutation should see the first flag
	snap2, err := store.Mutate(context.Background(), "test-client", func(s *Snapshot) error {
		if _, exists := s.Flags["feature-1"]; !exists {
			t.Fatal("feature-1 not found in second mutation")
		}
		return s.AddFlag("feature-2", false, "Second")
	})
	if err != nil {
		t.Fatalf("Second Mutate failed: %v", err)
	}

	if len(snap2.Flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(snap2.Flags))
	}
}
