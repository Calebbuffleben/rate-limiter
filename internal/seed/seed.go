package seed

import (
    "context"
    "encoding/json"
    "fmt"

    "rate-limiter/internal/config"
    "rate-limiter/internal/limiter"
    "rate-limiter/internal/store"
)

type tokenSeed struct {
    Token        string `json:"token"`
    RPS          int    `json:"rps"`
    BlockSeconds int    `json:"blockSeconds"`
}

func SeedTokens(ctx context.Context, st store.Store, cfg config.Config) error {
    if cfg.TokensSeedJSON == "" || cfg.TokensSeedJSON == "[]" {
        return nil
    }
    var seeds []tokenSeed
    if err := json.Unmarshal([]byte(cfg.TokensSeedJSON), &seeds); err != nil {
        return err
    }
    for _, s := range seeds {
        if s.Token == "" {
            continue
        }
        data, _ := json.Marshal(limiter.TokenCfg{RPS: s.RPS, BlockSeconds: s.BlockSeconds})
        key := fmt.Sprintf("rl:token:cfg:%s", s.Token)

        _ = st.Set(ctx, key, string(data), 0)
    }
    return nil
}


