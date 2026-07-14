package retry

import "time"

type backoffType string

var (
	Exponential backoffType = "exponential"
	Constant    backoffType = "constant"
	Linear      backoffType = "linear"
	Jitter      backoffType = "jitter"
)

var defaultAttempt uint = 1

type RetryBackoffConfig struct {
	Backoff  backoffType   `yaml:"backoff"`
	Attempts uint          `yaml:"attempts"`
	Delay    time.Duration `yam:"delay"`
}

func (b backoffType) IsExponential() bool {
	return b == "exponential"
}

func (b backoffType) IsConstant() bool {
	return b == "constant"
}
func (b backoffType) IsLinear() bool {
	return b == "linear"
}

func (b backoffType) IsJitter() bool {
	return b == "jitter"
}

func (b backoffType) Is(bType string) bool {
	return b == backoffType(bType)
}

func (c RetryBackoffConfig) Attempt() uint {
	if c.Attempts < 1 {
		return defaultAttempt
	}
	return c.Attempts
}
