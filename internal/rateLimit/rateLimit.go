package ratelimit

func NewRateLimiter(conf *Config) RateLimiter {
	if IsWindowType(conf) {
		return newOrGetWindow(conf)
	} else if IsBucketType(conf) {
		return newOrGetBucket(conf)
	}
	return nil
}

func IsWindowType(conf *Config) bool {
	return conf.Type == "window"
}

func IsBucketType(conf *Config) bool {
	return conf.Type == "bucket"
}
