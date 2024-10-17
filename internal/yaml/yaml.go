package yaml

import (
	"errors"
	"os"
	"wali/internal/utils"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingWebjobsInYaml = errors.New("missing webjobs in YAML file")
)

func ReadYaml(filename string, wj *Jobs) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, wj); err != nil {
		return err
	}

	if len(wj.WebJobs) == 0 {
		// no webjobs found
		return ErrMissingWebjobsInYaml

	} else {
		// convert raw frequency string to frequency in ms
		for k, v := range wj.WebJobs {
			if v.FrequencyMs, err = utils.ParseFrequency(v.RawFrequency); err != nil {
				return err
			}
			v.Key = k
			wj.WebJobs[k] = v
		}
	}

	return nil
}
