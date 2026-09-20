package ratelimit

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/time/rate"
)

type bucket struct {
	rateLimit *rate.Limiter
	key       string
}

var ErrBucketLimitExceeded error = errors.New("bucket limit exceeded")
var ErrBucketKeyEmpty error = errors.New("bucket key cannot be empty")
var bucketCache map[string]*bucket = make(map[string]*bucket)

func newOrGetBucket(conf *Config) *bucket {
	mu.Lock()
	defer mu.Unlock()
	key, _ := extractServiceNameFromKey(conf.Key)
	b, ok := bucketCache[key]
	if !ok {
		b = &bucket{
			key:       key,
			rateLimit: rate.NewLimiter(rate.Limit(conf.Rate), conf.Burst),
		}
		bucketCache[key] = b
	}
	return b
}

func (b *bucket) Limit() int {
	return int(b.rateLimit.Limit())
}

func (b *bucket) Burst() int {
	return b.rateLimit.Burst()
}

func extractServiceNameFromKey(key string) (string, error) {
	splitedKey := strings.Split(key, "_")
	if len(splitedKey) < 1 {
		return "", ErrBucketKeyEmpty
	}
	return splitedKey[0], nil
}

func (b *bucket) Allow() error {
	if !b.rateLimit.Allow() {
		return ErrBucketLimitExceeded
	}
	return nil
}

func (b *bucket) clearCache() {
	mu.Lock()
	defer mu.Unlock()
	bucketCache = make(map[string]*bucket)
}

func (b *bucket) ClearCacheForTest(t *testing.T) {
	t.Helper()
	bucketCache = make(map[string]*bucket)
}
