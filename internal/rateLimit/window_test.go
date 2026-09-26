package ratelimit

import (
	"errors"
	"testing"
)

func TestNewOrGetWindow(t *testing.T) {
	cfg := mockConfig("window")
	t.Run("must add in cache", func(t *testing.T) {
		t.Cleanup(func() {
			cache = map[string]*window{}
		})
		newOrGetWindow(cfg)
		_, ok := cache[cfg.Key]
		if !ok {
			t.Errorf("must be added in cache")
		}
	})
	t.Run("must retrive from cache", func(t *testing.T) {
		t.Cleanup(func() {
			cache = map[string]*window{}
		})
		w := newOrGetWindow(cfg)
		expected, _ := cache[cfg.Key]
		if w != expected {
			t.Errorf("must be retrieved from cache in cache")
		}
	})
}

func TestWindowAllow(t *testing.T) {

	t.Run("window allow must fail (block request)", func(t *testing.T) {
		t.Cleanup(func() {
			cache = map[string]*window{}
		})
		cfg := mockConfig("window")
		cfg.Limit = 2
		cfg.Interval = "100ms"
		w := newOrGetWindow(cfg)
		var err error
		for range 4 {
			err = w.Allow()
			if err != nil {
				break
			}
		}
		if !errors.Is(err, ErrWindowLimitExceeded) {
			t.Errorf("allow must fail with window error: %s, got %s", ErrWindowLimitExceeded, err)
		}
	})

	t.Run("window rate limit must allow request", func(t *testing.T) {
		cfg := mockConfig("window")
		cfg.Limit = 60
		cfg.Interval = "60s"
		w := newOrGetWindow(cfg)
		var err error
		for range 10 {
			err = w.Allow()
			if err != nil {
				break
			}
		}
		if err != nil {
			t.Errorf("window allow must succeed allowing request, got %s", err)
		}
	})
}
