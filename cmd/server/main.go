package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/joho/godotenv"

    "rate-limiter/internal/config"
    "rate-limiter/internal/http/middleware"
    "rate-limiter/internal/limiter"
    "rate-limiter/internal/seed"
    "rate-limiter/internal/store/redisstore"
)

func main() {
    _ = godotenv.Load()

    cfg := config.Load()

    rs := redisstore.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

    if err := seed.SeedTokens(context.Background(), rs, cfg); err != nil {
        log.Printf("seed tokens error: %v", err)
    }

    l := limiter.New(cfg, rs)
    rl := middleware.NewRateLimiterMiddleware(cfg, l)

    mux := http.NewServeMux()
    mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("pong"))
    })
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    srv := &http.Server{
        Addr:         ":" + cfg.ServerPort,
        Handler:      rl.Handler(mux),
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        log.Printf("listening on :%s", cfg.ServerPort)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
    <-stop
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(ctx)
}


