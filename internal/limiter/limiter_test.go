package limiter

import (
    "context"
    "os"
    "testing"
    "time"

    "rate-limiter/internal/config"
    "rate-limiter/internal/store/redisstore"
)

func TestIPRateLimitBlocksAfterThreshold(t *testing.T) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "redis:6379"
    }
    rs := redisstore.New(addr, "", 0)
    cfg := config.Config{
        RateIPEnabled:      true,
        RateIPRPS:          3,
        RateIPBlockSeconds: 2,
        RateTokenEnabled:   false,
    }
    l := New(cfg, rs)
    ctx := context.Background()

    ip := "1.2.3.4"
    for i := 0; i < 3; i++ {
        res, err := l.AllowByIP(ctx, ip)
        if err != nil || !res.Allowed {
            t.Fatalf("unexpected disallow at %d: %#v err=%v", i, res, err)
        }
    }
    res, err := l.AllowByIP(ctx, ip)
    if err != nil {
        t.Fatal(err)
    }
    if res.Allowed {
        t.Fatalf("expected block on 4th request")
    }
    if res.RetryAfter <= 0 {
        t.Fatalf("expected positive retry after")
    }
    time.Sleep(2100 * time.Millisecond)
    res, err = l.AllowByIP(ctx, ip)
    if err != nil || !res.Allowed {
        t.Fatalf("expected allow after block, got %#v err=%v", res, err)
    }
}

func TestTokenOverrideBeatsIPLimit(t *testing.T) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "redis:6379"
    }
    rs := redisstore.New(addr, "", 0)
    cfg := config.Config{
        RateIPEnabled:          true,
        RateIPRPS:              1,
        RateIPBlockSeconds:     2,
        RateTokenEnabled:       true,
        RateTokenDefaultRPS:           5,
        RateTokenDefaultBlockSeconds:  2,
    }
    l := New(cfg, rs)
    ctx := context.Background()

    token := "abc123"
    for i := 0; i < 3; i++ {
        res, err := l.AllowByToken(ctx, token)
        if err != nil || !res.Allowed {
            t.Fatalf("unexpected token disallow at %d: %#v err=%v", i, res, err)
        }
    }
}

func TestTokenBlocksAfterThreshold(t *testing.T) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "redis:6379"
    }
    rs := redisstore.New(addr, "", 0)
    cfg := config.Config{
        RateTokenEnabled:              true,
        RateTokenDefaultRPS:           2,
        RateTokenDefaultBlockSeconds:  2,
    }
    l := New(cfg, rs)
    ctx := context.Background()

    token := "token-block"
    for i := 0; i < 2; i++ {
        res, err := l.AllowByToken(ctx, token)
        if err != nil || !res.Allowed {
            t.Fatalf("unexpected disallow at %d: %#v err=%v", i, res, err)
        }
    }
    res, err := l.AllowByToken(ctx, token)
    if err != nil {
        t.Fatal(err)
    }
    if res.Allowed {
        t.Fatalf("expected block on 3rd token request")
    }
}


