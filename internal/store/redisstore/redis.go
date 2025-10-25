package redisstore

import (
    "context"
    "time"

    redis "github.com/redis/go-redis/v9"
)

type RedisStore struct {
    client *redis.Client
}

func New(addr string, password string, db int) *RedisStore {
    c := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    return &RedisStore{client: c}
}

func (r *RedisStore) Client() *redis.Client { return r.client }

var incrScript = redis.NewScript(`
local v = redis.call('INCR', KEYS[1])
if v == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return v
`)

var setnxScript = redis.NewScript(`
local set = redis.call('SETNX', KEYS[1], ARGV[1])
if set == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return set
`)

func (r *RedisStore) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
    v, err := incrScript.Run(ctx, r.client, []string{key}, int64(ttl/time.Millisecond)).Int64()
    if err != nil {
        return 0, err
    }
    return v, nil
}

func (r *RedisStore) Exists(ctx context.Context, key string) (bool, error) {
    n, err := r.client.Exists(ctx, key).Result()
    if err != nil {
        return false, err
    }
    return n == 1, nil
}

func (r *RedisStore) SetNXWithTTL(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
    v, err := setnxScript.Run(ctx, r.client, []string{key}, value, int64(ttl/time.Millisecond)).Int64()
    if err != nil {
        return false, err
    }
    return v == 1, nil
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return "", nil
    }
    return val, err
}

func (r *RedisStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
    return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
    d, err := r.client.TTL(ctx, key).Result()
    if err == redis.Nil {
        return 0, nil
    }
    return d, err
}


