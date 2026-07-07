package ratelimit

import (
	"errors"
	"sync"
	"time"
)

var defaultInterval = "1m"
var ErrWindowNotFound error = errors.New("window not found")
var ErrWindowLimitExceeded error = errors.New("window limit exceeded")
var ErrEmptyCacheKey error = errors.New("erro on build window rate limit cache key")

var cache map[string]*window = make(map[string]*window)
var mu sync.Mutex

type window struct {
	limit    int
	interval time.Duration
	count    int
	time     time.Time
	key      string
}

func newWindow(conf *Config) *window {
	parseIntervalToTimeDuration, err := time.ParseDuration(conf.Interval)
	if err != nil {
		parseIntervalToTimeDuration, _ = time.ParseDuration(conf.Interval)
	}
	return &window{
		limit:    conf.Limit,
		interval: parseIntervalToTimeDuration,
		count:    0,
		time:     time.Now(),
		key:      conf.Key,
	}
}

func newOrGetWindow(conf *Config) *window {
	mu.Lock()
	defer mu.Unlock()
	win, exists := cache[conf.Key]
	if !exists {
		win = newWindow(conf)
		cache[conf.Key] = win
	}
	return win
}

func (w *window) Interval() time.Time {
	return w.time
}
func (w *window) Limit() int {
	return w.limit
}
func (w *window) Count() int {
	return w.count
}

func (w *window) Time() time.Time {
	return w.time
}

func (w *window) Allow() error {
	mu.Lock()
	defer mu.Unlock()
	if w.key == "" {
		return ErrEmptyCacheKey
	}
	windowIsInvalid := w.time.Add(w.interval).Before(time.Now())
	if windowIsInvalid {
		// reset the window and allow the request
		w.count = 1
		w.time = time.Now()
		return nil
	} else {
		if w.count <= w.limit {
			// increment the count and allow the request
			w.count++
			return nil
		}
	}
	//too many requests, reject the request
	return ErrWindowLimitExceeded
}
