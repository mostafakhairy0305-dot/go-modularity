package analyzer

import (
	gomodularity "github.com/mostafakhairy0305-dot/go-modularity/gomodularity"
	policydomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/policyconfig"
)

// resolvePolicy builds the effective policy: an explicit config path, else
// .modularity.yml under directory, else the recommended defaults.
func resolvePolicy(configPath, directory string) (policydomain.Policy, error) {
	if configPath != "" {
		return policyconfig.Load(configPath)
	}

	discoverDir := directory
	if discoverDir == "" {
		discoverDir = "."
	}

	if path, ok := policyconfig.Discover(discoverDir); ok {
		return policyconfig.Load(path)
	}

	return policydomain.DefaultPolicy(), nil
}

// gatedMetrics unions the policy's constrained metrics into the display set so
// every gated metric is computed — a metric absent from the report cannot be
// checked.
func gatedMetrics(policy policydomain.Policy) []gomodularity.MetricName {
	base := gomodularity.DefaultMetrics()
	present := make(map[gomodularity.MetricName]bool, len(base))
	out := append([]gomodularity.MetricName(nil), base...)

	for _, name := range base {
		present[name] = true
	}

	for _, name := range policydomain.MetricNames(policy) {
		metric := gomodularity.MetricName(name)
		if present[metric] {
			continue
		}

		out = append(out, metric)
		present[metric] = true
	}

	return out
}
