package middleware

import (
    "context"
    "net/http"
    "strconv"
    "time"
    "strings"

    "rate-limiter/internal/config"
    "rate-limiter/internal/limiter"
    "rate-limiter/internal/util"
)

type RateLimiterMiddleware struct {
    cfg     config.Config
    limiter *limiter.Limiter
}

func NewRateLimiterMiddleware(cfg config.Config, l *limiter.Limiter) *RateLimiterMiddleware {
    return &RateLimiterMiddleware{cfg: cfg, limiter: l}
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        token := strings.TrimSpace(r.Header.Get(m.cfg.RateTokenHeader))

        var res limiter.Result
        var err error

        if token != "" && m.cfg.RateTokenEnabled {
            res, err = m.limiter.AllowByToken(ctx, token)
        } else if m.cfg.RateIPEnabled {
            ip := util.GetClientIP(r)
            res, err = m.limiter.AllowByIP(ctx, ip)
        } else {
            // No limiting enabled
            next.ServeHTTP(w, r)
            return
        }
        if err != nil {
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            return
        }
        if !res.Allowed {
            // Respond 429 with message and Retry-After seconds
            retrySecs := int(res.RetryAfter / time.Second)
            if retrySecs <= 0 {
                retrySecs = 1
            }
            w.Header().Set("Retry-After", strconv.Itoa(retrySecs))
            w.WriteHeader(http.StatusTooManyRequests)
            _, _ = w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
            return
        }
        next.ServeHTTP(w, r)
    })
}

func (m *RateLimiterMiddleware) AllowByToken(ctx context.Context, token string) (limiter.Result, error) {
    return m.limiter.AllowByToken(ctx, token)
}


