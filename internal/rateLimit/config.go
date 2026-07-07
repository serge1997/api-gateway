package ratelimit

type Config struct {
	Type     string `json:"type" yaml:"type"`
	Rate     int    `json:"rate" yaml:"rate"`
	Limit    int    `json:"limit" yaml:"limit"`
	Interval string `json:"interval" yaml:"interval"`
	Burst    int    `json:"burst" yaml:"burst"`
	Key      string `json:"key" yaml:"key"`
}
