package core

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// Rate-limit paste creation

var (
	pasteTimestamps    sync.Map
	rateLimitPerMinute = 5
	windowSize         = time.Minute.Nanoseconds()
)

func CheckAndRecordRateLimit(identifier string) bool {
	now := time.Now().UnixNano()

	existing, _ := pasteTimestamps.Load(identifier)
	timestamps, _ := existing.([]int64)

	var valid []int64
	for _, ts := range timestamps {
		if now-ts <= windowSize {
			valid = append(valid, ts)
		}
	}

	if len(valid) >= rateLimitPerMinute {
		return false
	}

	pasteTimestamps.Store(identifier, append(valid, now))
	return true
}

func GetIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return strings.Split(ip, ":")[0]
}
