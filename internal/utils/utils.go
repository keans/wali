package utils

import (
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	ErrEmptyDurationString                 = errors.New("empty duration string")
	ErrInvalidCharactersInDurationString   = errors.New("invalid characters in duration string")
	ErrInvalidNumberInDurationString       = errors.New("invalid number in duration string")
	ErrMissingUnitMappingForDurationString = errors.New("missing unit mapping for duration string")
)

func ParseFrequency(freq string) (int64, error) {
	// define mapping of abbreviation to duration
	unitMapping := map[string]time.Duration{
		"s": time.Second,
		"m": time.Minute,
		"h": time.Hour,
		"d": 24 * time.Hour,
		"w": 7 * 24 * time.Hour,
	}

	if freq == "" {
		// empty raw frequency string
		return 0, ErrEmptyDurationString
	}

	re := regexp.MustCompile(`^(\d+[smhdw])+$`)
	if !re.MatchString(freq) {
		// invalid characters in raw frequency
		return 0, ErrInvalidCharactersInDurationString
	}

	// get regex matches
	re = regexp.MustCompile(`(\d+)([smhdw])`)
	matches := re.FindAllStringSubmatch(freq, -1)

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

	return duration.Milliseconds(), nil
}

func Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		panic(res.Body)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return body, nil
}
