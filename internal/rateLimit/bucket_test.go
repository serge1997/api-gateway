package ratelimit

import (
	"errors"
	"testing"
)

func TestBucketMustReturnBucketType(t *testing.T) {
	cfg := mockConfig("bucket")
	var bucketType interface{} = NewRateLimiter(cfg)
	t.Run("must return RateLimiter type", func(t *testing.T) {
		b, ok := bucketType.(RateLimiter)
		if !ok {
			t.Errorf("type assert of RateLimiter must return true, got %v", b)
		}
	})

	t.Run("must return true config type is type bucket", func(t *testing.T) {
		if !IsBucketType(cfg) {
			t.Errorf("config type must bucket, got %v", cfg.Type)
		}
	})

	t.Run("must save bucket in cache", func(t *testing.T) {
		key, _ := extractServiceNameFromKey(cfg.Key)
		newOrGetBucket(cfg)
		_, ok := bucketCache[key]
		if ok == false {
			t.Errorf("a bucket must be added to the cache")
		}
	})

	t.Run("must retrieved from cache", func(t *testing.T) {
		b := newOrGetBucket(cfg)
		key, _ := extractServiceNameFromKey(cfg.Key)
		expected, _ := bucketCache[key]
		if expected != b {
			t.Error("must retrieved bucket from cache")
		}
	})
}

func TestBucketExtractServiceNameFromKey(t *testing.T) {
	cfg := mockConfig("bucket")

	t.Run("extract service name must fail", func(t *testing.T) {
		key := cfg.Key
		cfg.Key = ""
		t.Cleanup(func() {
			cfg.Key = key
		})
		_, err := extractServiceNameFromKey(cfg.Key)
		if err == nil {
			t.Errorf("extract service name must fail")
		}
	})
	t.Run("extract service name must succeed", func(t *testing.T) {
		_, err := extractServiceNameFromKey(cfg.Key)
		if err != nil {
			t.Errorf("service name extraction must success, but got an error: %v", err)
		}
	})
	t.Run("must extract a service name from key", func(t *testing.T) {
		expected := "posts"
		key, _ := extractServiceNameFromKey(cfg.Key)
		if key != expected {
			t.Errorf("extracted service name must be %s, got %s", expected, key)
		}
	})
}

func TestBucketAllow(t *testing.T) {
	t.Run("must fail with bucket error", func(t *testing.T) {
		cfg := &Config{
			Type:     "bucket",
			Rate:     1,
			Limit:    2,
			Burst:    2,
			Interval: "100ms",
			Key:      "posts_3893_by_user",
		}
		b := newOrGetBucket(cfg)
		t.Cleanup(func() {
			bucketCache = map[string]*bucket{}
		})
		var err error
		for range 3 {
			err = b.Allow()
			if err != nil {
				break
			}
		}
		if !errors.Is(ErrBucketLimitExceeded, err) {
			t.Errorf("allow fail must return bucket err, got %v", err)
		}
	})

	t.Run("must allow requests", func(t *testing.T) {
		cfg := &Config{
			Type:     "bucket",
			Limit:    10,
			Burst:    8,
			Rate:     10,
			Interval: "100ms",
			Key:      "posts_3893_by_user",
		}
		b := newOrGetBucket(cfg)
		var err error
		for range 5 {
			err = b.Allow()
			if err != nil {
				break
			}
		}
		if err != nil {
			t.Errorf("allow must succeed, but got: %v", err)
		}
	})
}
