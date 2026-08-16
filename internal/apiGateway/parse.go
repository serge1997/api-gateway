package apigateway

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/serge1997/apigateway/shared"
)

var ErrFailToReadYamlFile = errors.New("error on load config yml service file")

func parse() (*apiGateway, error) {
	var gtw apiGateway
	ymlBytes, err := shared.LoadServiceYml()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("%s. reason: %s", ErrFailToReadYamlFile, err))
	}
	if err := yaml.Unmarshal(ymlBytes, &gtw); err != nil {
		return nil, fmt.Errorf("erro on unmarshall servivce byte. reason: %s", err)
	}
	return &gtw, nil
}
