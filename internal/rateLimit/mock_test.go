package ratelimit

func mockConfig(tpe string) *Config {
	return &Config{
		Type:     tpe,
		Limit:    6,
		Interval: "100ms",
		Key:      "posts_3893_by_user",
	}
}
