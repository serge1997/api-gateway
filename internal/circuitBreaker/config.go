package circuitbreaker

type Config struct {
	FailureThreshold int      `yaml:"failure_threshold"` //limit of failures before opening the circuit
	RetryTimeout     string   `yaml:"retry_timeout"`     //time in seconds to wait before retrying after the circuit is opened
	Before           []string `yaml:"before"`
}
