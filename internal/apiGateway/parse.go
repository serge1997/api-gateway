package apigateway

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

var ErrFailToReadYamlFile = errors.New("error on load config yml service file")

func parse() (*apiGateway, error) {
	var gtw apiGateway
	ymlBytes, err := loadYmlFile()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("%s. reason: %s", ErrFailToReadYamlFile, err))
	}
	if err := yaml.Unmarshal(ymlBytes, &gtw); err != nil {
		return nil, fmt.Errorf("erro on unmarshall servivce byte. reason: %s", err)
	}
	return &gtw, nil
}
func loadYmlFile() ([]byte, error) {
	absDir, err := filepath.Abs("./")
	if err != nil {
		return nil, fmt.Errorf("erro on read asbolute path. detail: %v", err)
	}
	data, err := os.ReadFile(fmt.Sprintf("%s/api-gateway.yml", absDir))
	if err != nil {
		return nil, fmt.Errorf("erro on open service yml file. detail: %v", err)
	}
	return data, nil
}
