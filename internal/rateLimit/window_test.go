package ratelimit

import "testing"

func TestWewOrGetWindow(t *testing.T) {
	cfg := mockConfig("window")
	t.Run("must add in cache", func(t *testing.T) {
		newOrGetWindow(cfg)
		_, ok := cache[cfg.Key]
		if !ok {
			t.Errorf("must be added in cache")
		}
	})
	t.Run("must retrive from cache", func(t *testing.T) {
		w := newOrGetWindow(cfg)
		expected, _ := cache[cfg.Key]
		if w != expected {
			t.Errorf("must be retrieved from cache in cache")
		}
	})
}
