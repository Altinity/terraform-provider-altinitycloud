package common

import "testing"

func TestClickHouseNameRegex(t *testing.T) {
	tests := map[string]bool{
		"a":                false, // the API requires at least 2 characters
		"ab":               true,
		"ch-1":             true,
		"123456789012345":  true,  // 15, the maximum
		"1234567890123456": false, // 16
		"-ch":              false,
		"ch-":              false,
		"CH":               false,
		"ch_1":             false,
		"":                 false,
	}

	for name, want := range tests {
		if got := clickHouseNameRegex.MatchString(name); got != want {
			t.Errorf("%q: got %v, want %v", name, got, want)
		}
	}
}

func TestClickHouseDiskNameRegex(t *testing.T) {
	tests := map[string]bool{
		"disk":             true,
		"disk1":            true,
		"disk-cold":        true,
		"disk123456789012": true,  // 16, the maximum
		"default":          false, // reserved for the main volume
		"volume1":          false,
		"":                 false,
	}

	for name, want := range tests {
		if got := clickHouseDiskNameRegex.MatchString(name); got != want {
			t.Errorf("%q: got %v, want %v", name, got, want)
		}
	}
}
