package metrics

// TypeMetricOrder is the fixed rendering order of type-level metrics.
func TypeMetricOrder() []string {
	return []string{
		MetricAMC, MetricLCOM1, MetricLCOM96b, MetricCAMC, MetricTCC,
		MetricCBO, MetricReusability,
	}
}

// PackageMetricOrder is the fixed rendering order of package-level metrics.
func PackageMetricOrder() []string {
	return []string{MetricAbstractness, MetricInstability, MetricDistance}
}

// dependencies is the metric-level dependency graph. Selecting a metric
// pulls its dependencies into the compute set; internal inputs such as the
// method-field matrix are handled inside features and need no entry.
var dependencies = map[string][]string{
	MetricReusability: {MetricLCOM96b, MetricAMC, MetricCBO},
	MetricDistance:    {MetricAbstractness, MetricInstability},
}

// Closure expands a selected display set to the full compute set: the
// transitive closure over metric dependencies, deduplicated, in a
// deterministic order. A metric computed only to satisfy a dependency is
// not rendered unless also selected.
func Closure(selected []string) []string {
	seen := make(map[string]bool, len(selected))

	var visit func(name string)

	visit = func(name string) {
		if seen[name] {
			return
		}

		seen[name] = true
		for _, dep := range dependencies[name] {
			visit(dep)
		}
	}
	for _, name := range selected {
		visit(name)
	}

	closure := make([]string, 0, len(seen))
	for _, name := range TypeMetricOrder() {
		if seen[name] {
			closure = append(closure, name)
		}
	}

	for _, name := range PackageMetricOrder() {
		if seen[name] {
			closure = append(closure, name)
		}
	}

	return closure
}
