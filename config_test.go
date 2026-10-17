package main

import "testing"

func getter(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestParseSettingsDefaults(t *testing.T) {
	s := parseSettings(getter(nil))
	if s.ThresholdPercent != 75 || !s.RemoveWhenBelow || s.IntervalMinutes != 5 {
		t.Fatalf("unexpected defaults: %+v", s)
	}
}

func TestParseSettingsValues(t *testing.T) {
	s := parseSettings(getter(map[string]string{
		cfgThreshold: "80",
		cfgRemove:    "false",
		cfgInterval:  "15.0",
	}))
	if s.ThresholdPercent != 80 || s.RemoveWhenBelow || s.IntervalMinutes != 15 {
		t.Fatalf("unexpected settings: %+v", s)
	}
}

func TestParseSettingsInvalidFallsBack(t *testing.T) {
	s := parseSettings(getter(map[string]string{
		cfgThreshold: "0",
		cfgRemove:    "maybe",
		cfgInterval:  "abc",
	}))
	if s.ThresholdPercent != 75 || !s.RemoveWhenBelow || s.IntervalMinutes != 5 {
		t.Fatalf("invalid values should fall back to defaults: %+v", s)
	}
}
