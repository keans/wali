package models

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"
	"wali/internal/database"

	"golang.org/x/exp/rand"
)

var (
	ErrEmptyDurationString                 = errors.New("empty duration string")
	ErrInvalidCharactersInDurationString   = errors.New("invalid characters in duration string")
	ErrInvalidNumberInDurationString       = errors.New("invalid number in duration string")
	ErrMissingUnitMappingForDurationString = errors.New("missing unit mapping for duration string")
)

type WebJobs struct {
	WebJobs []WebJob `yaml:"webjobs"`
}

type WebJob struct {
	Name         string `yaml:"name"`
	Url          string `yaml:"url"`
	RawFrequency string `yaml:"frequency"`
}

func NewWebJob(name string, url string, rawFrequency string) *WebJob {
	return &WebJob{
		Name:         name,
		Url:          url,
		RawFrequency: rawFrequency,
	}
}

func (wj *WebJob) Execute(db *database.Database) bool {
	fmt.Printf("Loading URL: %v\n", wj.Url)

	// rand.Intn(3)*int(time.Millisecond)
	time.Sleep(time.Duration(rand.Intn(3000) * int(time.Millisecond)))

	fmt.Printf("DONE Loading URL: %v\n", wj.Url)

	return true
}

func (wj *WebJob) IsValid() bool {
	if wj.Name == "" {
		slog.Error("name cannot be empty", "name",
			wj.Name)

		return false
	}

	if wj.Url == "" {
		slog.Error("url cannot be empty", "url",
			wj.Url)

		return false
	}

	if _, err := wj.Frequency(); err != nil {
		slog.Error("could not parse raw frequency", "rawFrequeny",
			wj.RawFrequency, "error", err)

		return false
	}

	return true
}

func (wj *WebJob) Frequency() (time.Duration, error) {
	// define mapping of abbreviation to duration
	unitMapping := map[string]time.Duration{
		"s": time.Second,
		"m": time.Minute,
		"h": time.Hour,
		"d": 24 * time.Hour,
		"w": 7 * 24 * time.Hour,
	}

	if wj.RawFrequency == "" {
		// empty raw frequency string
		return 0, ErrEmptyDurationString
	}

	re := regexp.MustCompile(`^(\d+[smhdw])+$`)
	if !re.MatchString(wj.RawFrequency) {
		// invalid characters in raw frequency
		return 0, ErrInvalidCharactersInDurationString
	}

	// get regex matches
	re = regexp.MustCompile(`(\d+)([smhdw])`)
	matches := re.FindAllStringSubmatch(wj.RawFrequency, -1)

	var duration time.Duration
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, ErrInvalidNumberInDurationString
		}

		if unit, exists := unitMapping[match[2]]; exists {
			duration += time.Duration(value) * unit
		} else {
			return 0, ErrMissingUnitMappingForDurationString
		}
	}

	return duration, nil
}
