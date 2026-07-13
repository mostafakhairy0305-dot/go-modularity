package main

import (
	"fmt"
	"strconv"
	"strings"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	policydomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/policyconfig"
)

// override is one CLI policy bound: a condition key and its numeric value.
// Metric keys may be legacy/global ("amc") or scoped ("type.amc",
// "package.distance").
type override struct {
	key   string
	value float64
}

// overrideList collects repeated -max / -min flags. It implements flag.Value.
type overrideList struct {
	items []override
}

func (o *overrideList) String() string {
	parts := make([]string, len(o.items))
	for i, ov := range o.items {
		parts[i] = ov.key + "=" + strconv.FormatFloat(ov.value, 'g', -1, 64)
	}

	return strings.Join(parts, ",")
}

func (o *overrideList) Set(value string) error {
	key, number, ok := strings.Cut(value, "=")
	if !ok {
		return fmt.Errorf("expected key=value, got %q", value)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("empty key in %q", value)
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(number), 64)
	if err != nil {
		return fmt.Errorf("invalid number in %q: %w", value, err)
	}

	o.items = append(o.items, override{key: key, value: parsed})

	return nil
}

// resolvePolicy builds the effective policy: a base from the explicit config
// path, else an auto-discovered .modularity.yml, else the recommended
// defaults; then the CLI overrides on top (defaults < file < flags). It
// returns a human-readable source label so the CLI can say which policy ran,
// and the validated policy.
func resolvePolicy(configPath string, maxima, minima overrideList) (policydomain.Policy, string, error) {
	var (
		policy policydomain.Policy
		source string
	)

	switch {
	case configPath != "":
		loaded, err := policyconfig.Load(configPath)
		if err != nil {
			return policydomain.Policy{}, "", err
		}

		policy, source = loaded, configPath

	default:
		if path, ok := policyconfig.Discover("."); ok {
			loaded, err := policyconfig.Load(path)
			if err != nil {
				return policydomain.Policy{}, "", err
			}

			policy, source = loaded, path
		} else {
			policy, source = policydomain.DefaultPolicy(), "recommended defaults"
		}
	}

	for _, ov := range maxima.items {
		if err := policydomain.ApplyOverride(&policy, ov.key, policydomain.ComparatorMax, ov.value); err != nil {
			return policydomain.Policy{}, "", err
		}
	}

	for _, ov := range minima.items {
		if err := policydomain.ApplyOverride(&policy, ov.key, policydomain.ComparatorMin, ov.value); err != nil {
			return policydomain.Policy{}, "", err
		}
	}

	if len(maxima.items) > 0 || len(minima.items) > 0 {
		source += " + flag overrides"
	}

	if err := policydomain.Validate(policy); err != nil {
		return policydomain.Policy{}, "", err
	}

	return policy, source, nil
}

// gatedMetrics unions the policy's constrained metrics into the display set so
// every gated metric is computed and rendered — a metric absent from the
// report cannot be checked. Base order is preserved; new names are appended.
func gatedMetrics(base []gomodularity.MetricName, policy policydomain.Policy) []gomodularity.MetricName {
	if len(base) == 0 {
		base = gomodularity.DefaultMetrics()
	}

	present := make(map[gomodularity.MetricName]bool, len(base))
	for _, name := range base {
		present[name] = true
	}

	out := append([]gomodularity.MetricName(nil), base...)
	for _, name := range policydomain.MetricNames(policy) {
		metric := gomodularity.MetricName(name)
		if !present[metric] {
			out = append(out, metric)
			present[metric] = true
		}
	}

	return out
}
