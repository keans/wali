package yaml

import (
	"log/slog"
)

type Job struct {
	Key          string // key from WebJobs map
	Url          string `yaml:"url"`
	Xpath        string `yaml:"xpath"`
	RawFrequency string `yaml:"frequency"`
	FrequencyMs  int64
}

type Jobs struct {
	WebJobs map[string]Job `yaml:"webjobs"`
}

func (wj *Job) IsValid() bool {
	if wj.Key == "" {
		slog.Error("key cannot be empty", "key", wj.Key)

		return false
	}

	if wj.Url == "" {
		slog.Error("url cannot be empty", "url", wj.Url)

		return false
	}

	if wj.FrequencyMs == 0 {
		slog.Error("frequency cannot be 0", "frequency", wj.FrequencyMs)

		return false
	}

	return true
}
