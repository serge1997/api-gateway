package retry

type backoffType string

var (
	Exponential backoffType = "exponential"
	Constant    backoffType = "constant"
	Linear      backoffType = "linear"
	Jitter      backoffType = "jitter"
)

type RetryConfig struct {
	Backoff  backoffType `yaml:"backoff"`
	Attempts uint        `yaml:"attempts"`
	MaxDelay uint        `yam:"max_delay"`
}

func (b backoffType) IsExponential() bool {
	return b == "exponential"
}

func (b backoffType) IsConstant() bool {
	return b == "constant"
}
func (b backoffType) IsLinear() bool {
	return b == "constant"
}

func (b backoffType) IsJitter() bool {
	return b == "jitter"
}

func (b backoffType) Is(bType string) bool {
	return b == backoffType(bType)
}
