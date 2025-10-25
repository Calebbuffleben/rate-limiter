package limiter

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"

    "rate-limiter/internal/config"
    "rate-limiter/internal/store"
)

type KeyType string

const (
    KeyTypeIP    KeyType = "ip"
    KeyTypeToken KeyType = "token"
)

type TokenCfg struct {
    RPS          int `json:"rps"`
    BlockSeconds int `json:"blockSeconds"`
}

type Result struct {
    Allowed    bool
    RetryAfter time.Duration
}

type Limiter struct {
    cfg   config.Config
    store store.Store
}

func New(cfg config.Config, st store.Store) *Limiter {
    return &Limiter{cfg: cfg, store: st}
}

func (l *Limiter) AllowByIP(ctx context.Context, ip string) (Result, error) {
    if !l.cfg.RateIPEnabled {
        return Result{Allowed: true}, nil
    }
    return l.allow(ctx, KeyTypeIP, ip, l.cfg.RateIPRPS, l.cfg.RateIPBlockSeconds)
}


func (l *Limiter) AllowByToken(ctx context.Context, token string) (Result, error) {
    if !l.cfg.RateTokenEnabled {
        return Result{Allowed: true}, nil
    }
    rps := l.cfg.RateTokenDefaultRPS
    block := l.cfg.RateTokenDefaultBlockSeconds
    if token != "" {
        if tc, ok, err := l.getTokenCfg(ctx, token); err == nil && ok {
            if tc.RPS > 0 {
                rps = tc.RPS
            }
            if tc.BlockSeconds > 0 {
                block = tc.BlockSeconds
            }
        }
    }
    return l.allow(ctx, KeyTypeToken, token, rps, block)
}

func (l *Limiter) getTokenCfg(ctx context.Context, token string) (TokenCfg, bool, error) {
    key := fmt.Sprintf("rl:token:cfg:%s", token)
    v, err := l.store.Get(ctx, key)
    if err != nil {
        return TokenCfg{}, false, err
    }
    if v == "" {
        return TokenCfg{}, false, nil
    }
    var tc TokenCfg
    if err := json.Unmarshal([]byte(v), &tc); err != nil {
        return TokenCfg{}, false, err
    }
    return tc, true, nil
}

func (l *Limiter) allow(ctx context.Context, kt KeyType, id string, rps int, blockSeconds int) (Result, error) {
    if id == "" {
        return Result{}, errors.New("empty id")
    }
    if rps <= 0 {
        bKey := blockKey(kt, id)
        _, _ = l.store.SetNXWithTTL(ctx, bKey, "1", time.Duration(blockSeconds)*time.Second)
        ttl := l.ttlOrDefault(ctx, bKey, time.Duration(blockSeconds)*time.Second)
        return Result{Allowed: false, RetryAfter: ttl}, nil
    }

    bKey := blockKey(kt, id)
    if exists, _ := l.store.Exists(ctx, bKey); exists {
        ttl := l.ttlOrDefault(ctx, bKey, time.Duration(blockSeconds)*time.Second)
        return Result{Allowed: false, RetryAfter: ttl}, nil
    }

    now := time.Now().Unix()
    cKey := counterKey(kt, id, now)
    count, err := l.store.IncrWithTTL(ctx, cKey, 1500*time.Millisecond) // 1.5s TTL guards clock skews
    if err != nil {
        return Result{}, err
    }
    if int(count) > rps {
        ok, _ := l.store.SetNXWithTTL(ctx, bKey, "1", time.Duration(blockSeconds)*time.Second)
        var ttl time.Duration
        if ok {
            ttl = time.Duration(blockSeconds) * time.Second
        } else {
            ttl = l.ttlOrDefault(ctx, bKey, time.Duration(blockSeconds)*time.Second)
        }
        return Result{Allowed: false, RetryAfter: ttl}, nil
    }
    return Result{Allowed: true}, nil
}

func (l *Limiter) ttlOrDefault(ctx context.Context, key string, def time.Duration) time.Duration {
    type ttlGetter interface{ TTL(context.Context, string) (time.Duration, error) }
    if tg, ok := l.store.(ttlGetter); ok {
        if d, err := tg.TTL(ctx, key); err == nil && d > 0 {
            return d
        }
    }
    return def
}

func blockKey(kt KeyType, id string) string {
    return fmt.Sprintf("rl:block:%s:%s", string(kt), id)
}

func counterKey(kt KeyType, id string, sec int64) string {
    return fmt.Sprintf("rl:cnt:%s:%s:%d", string(kt), id, sec)
}


