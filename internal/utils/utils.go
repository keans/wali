package utils

import (
	"errors"
	"os"
	"wali/internal/models"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingWebjobsInYaml = errors.New("missing webjobs in YAML file")
)

func ReadYaml(filename string, wj *models.WebJobs) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, wj); err != nil {
		return err
	}

	if len(wj.WebJobs) == 0 {
		return ErrMissingWebjobsInYaml
	}

	return nil
}
