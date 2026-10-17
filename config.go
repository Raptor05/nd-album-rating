package main

import (
	"strconv"
	"strings"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

// Configuration keys (see manifest.json)
const (
	cfgThreshold = "threshold_percent"
	cfgRemove    = "remove_when_below"
	cfgInterval  = "interval_minutes"
)

// Defaults, used when a value is missing or invalid
const (
	defaultThreshold = 75
	defaultRemove    = true
	defaultInterval  = 5
)

type settings struct {
	RatingSettings
	IntervalMinutes int
}

func loadSettings() settings {
	return parseSettings(pdk.GetConfig)
}

// parseSettings reads the settings through the given getter (injectable for tests).
func parseSettings(get func(string) (string, bool)) settings {
	s := settings{
		RatingSettings:  RatingSettings{ThresholdPercent: defaultThreshold, RemoveWhenBelow: defaultRemove},
		IntervalMinutes: defaultInterval,
	}
	if v, ok := get(cfgThreshold); ok {
		if n, err := parseInt(v); err == nil && n >= 1 && n <= 100 {
			s.ThresholdPercent = n
		}
	}
	if v, ok := get(cfgRemove); ok {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			s.RemoveWhenBelow = b
		}
	}
	if v, ok := get(cfgInterval); ok {
		if n, err := parseInt(v); err == nil && n >= 1 && n <= 1440 {
			s.IntervalMinutes = n
		}
	}
	return s
}

// parseInt accepts "75" as well as "75.0" (JSON numbers may arrive as floats).
func parseInt(v string) (int, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	return int(f), err
}
