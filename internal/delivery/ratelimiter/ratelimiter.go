package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	Limiter      *rate.Limiter
	LastSeenTime time.Time
	mu           sync.Mutex
}

var (
	ipLimiter sync.Map
	once      sync.Once
)

func cleanIps() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			ipLimiter.Range(func(ip, c any) bool {
				user := c.(*client)
				if now.Sub(user.LastSeenTime) >= 15*time.Minute {
					ipLimiter.Delete(ip)
				}
				return true
			})
		}
	}()
}

func RateLimit(key string, rl, burst int) *rate.Limiter {
	once.Do(cleanIps)
	lim, _ := ipLimiter.LoadOrStore(key,
		&client{
			Limiter:      rate.NewLimiter(rate.Limit(rl), burst),
			LastSeenTime: time.Now(),
		},
	)
	user := lim.(*client)
	user.mu.Lock()
	user.LastSeenTime = time.Now()
	user.mu.Unlock()
	return user.Limiter
}
