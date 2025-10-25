package store

import (
    "context"
    "time"
)

type Store interface {
    IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
    Exists(ctx context.Context, key string) (bool, error)
    SetNXWithTTL(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
}


