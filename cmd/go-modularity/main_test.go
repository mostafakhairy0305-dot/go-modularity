package main

import (
	"reflect"
	"testing"
)

// White-box: comma-list parsing trims blanks and empties.
func TestSplitList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , ,b ", []string{"a", "b"}},
	}
	for _, tt := range tests {
		if got := splitList(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("splitList(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// White-box: metric parsing maps names to MetricName values.
func TestParseMetrics(t *testing.T) {
	t.Parallel()

	if got := parseMetrics(""); got != nil {
		t.Errorf("empty = %v, want nil", got)
	}

	got := parseMetrics("amc,tcc")
	if len(got) != 2 || string(got[0]) != "amc" || string(got[1]) != "tcc" {
		t.Errorf("parseMetrics = %v", got)
	}
}

// White-box: bad flags and formats exit with code 2 before any loading.
func TestRunUsageErrors(t *testing.T) {
	t.Parallel()

	if code := run([]string{"--format=xml"}); code != 2 {
		t.Errorf("invalid format exit = %d, want 2", code)
	}

	if code := run([]string{"--this-flag-does-not-exist"}); code != 2 {
		t.Errorf("unknown flag exit = %d, want 2", code)
	}
}
