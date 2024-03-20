package limiter

import (
	"github.com/y2a-labs/evaluate/models"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ProviderRateLimiter struct {
	limiter     *rate.Limiter
	lastUpdated time.Time
}

type RateLimiterManager struct {
	limiters sync.Map // Key: Provider ID, Value: *ProviderRateLimiter
}

func NewRateLimiterManager() *RateLimiterManager {
	return &RateLimiterManager{}
}

func (m *RateLimiterManager) GetLimiter(provider *models.Provider) *rate.Limiter {
	val, ok := m.limiters.Load(provider.ID)
	if ok {
		prl := val.(*ProviderRateLimiter)
		// Check if the configuration has changed since the last update
		if prl.lastUpdated.Before(provider.UpdatedAt) {
			m.UpdateLimiter(provider)
		}
		return prl.limiter
	}
	return m.UpdateLimiter(provider)
}

func (m *RateLimiterManager) UpdateLimiter(provider *models.Provider) *rate.Limiter {
	var interval time.Duration
	switch provider.Unit {
	case "second(s)":
		interval = time.Second * time.Duration(provider.Interval)
	case "minute(s)":
		interval = time.Minute * time.Duration(provider.Interval)
	default:
		interval = time.Second // Default to seconds if unspecified
	}

	// Create a new rate limiter for this provider
	limiter := rate.NewLimiter(rate.Every(interval/time.Duration(provider.Requests)), provider.Requests)
	m.limiters.Store(provider.ID, &ProviderRateLimiter{limiter: limiter, lastUpdated: time.Now()})
	val, _ := m.limiters.Load(provider.ID)
	return val.(*ProviderRateLimiter).limiter
}
