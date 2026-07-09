package shared

import (
	"fmt"
	"os"
)

func LoadServiceYml() ([]byte, error) {
	data, err := os.ReadFile("./../../services.yml")
	if err != nil {
		return nil, fmt.Errorf("erro on open service yml file. detail: %v", err)
	}
	return data, nil
}
