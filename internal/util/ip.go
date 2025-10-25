package util

import (
    "net"
    "net/http"
    "strings"
)

func GetClientIP(r *http.Request) string {
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        parts := strings.Split(xff, ",")
        ip := strings.TrimSpace(parts[0])
        if ip != "" {
            return ip
        }
    }
    if xr := r.Header.Get("X-Real-IP"); xr != "" {
        return xr
    }
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return host
}


