package config

import (
    "log"
    "os"
    "strconv"
)

type Config struct {
    ServerPort string

    RedisAddr     string
    RedisPassword string
    RedisDB       int

    RateIPEnabled          bool
    RateIPRPS              int
    RateIPBlockSeconds     int
    RateTokenEnabled              bool
    RateTokenHeader               string
    RateTokenDefaultRPS           int
    RateTokenDefaultBlockSeconds  int

    TokensSeedJSON string
}

func Load() Config {
    mustLookup := func(key string) string {
        v, ok := os.LookupEnv(key)
        if !ok {
            log.Fatalf("missing required env: %s", key)
        }
        return v
    }
    mustLookupInt := func(key string) int {
        v := mustLookup(key)
        n, err := strconv.Atoi(v)
        if err != nil {
            log.Fatalf("invalid int for %s: %q", key, v)
        }
        return n
    }
    mustLookupBool := func(key string) bool {
        v := mustLookup(key)
        b, err := strconv.ParseBool(v)
        if err != nil {
            log.Fatalf("invalid bool for %s: %q", key, v)
        }
        return b
    }

    return Config{
        ServerPort:                  mustLookup("PORT"),
        RedisAddr:                   mustLookup("REDIS_ADDR"),
        RedisPassword:               mustLookup("REDIS_PASSWORD"),
        RedisDB:                     mustLookupInt("REDIS_DB"),
        RateIPEnabled:               mustLookupBool("RATE_LIMIT_IP_ENABLED"),
        RateIPRPS:                   mustLookupInt("RATE_LIMIT_IP_RPS"),
        RateIPBlockSeconds:          mustLookupInt("RATE_LIMIT_IP_BLOCK_SECONDS"),
        RateTokenEnabled:            mustLookupBool("RATE_LIMIT_TOKEN_ENABLED"),
        RateTokenHeader:             mustLookup("RATE_LIMIT_HEADER"),
        RateTokenDefaultRPS:         mustLookupInt("RATE_LIMIT_TOKEN_DEFAULT_RPS"),
        RateTokenDefaultBlockSeconds: mustLookupInt("RATE_LIMIT_TOKEN_DEFAULT_BLOCK_SECONDS"),
        TokensSeedJSON:              mustLookup("RATE_LIMIT_TOKENS_JSON"),
    }
}


