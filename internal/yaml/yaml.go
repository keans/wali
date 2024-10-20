package yaml

import (
	"errors"
	"os"

	"github.com/keans/wali/internal/utils"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingWebjobsInYaml = errors.New("missing webjobs in YAML file")
)

// structure of the YAML file
type WaliYaml struct {
	Smtp    Smtp           `yaml:"smtp"`
	WebJobs map[string]Job `yaml:"webjobs"`
}

// read the YAML file of the WaliYaml format
func ReadYaml(filename string, wj *WaliYaml) error {
	// read the YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// unmarshall it
	if err := yaml.Unmarshal(data, wj); err != nil {
		return err
	}

	if len(wj.WebJobs) == 0 {
		// no webjobs found => return error
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
