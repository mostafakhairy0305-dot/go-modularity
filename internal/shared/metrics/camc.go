package metrics

// MetricCAMC is the metric name for Cohesion Among Methods of a Class.
const MetricCAMC = "camc"

// DefinitionCAMC versions the CAMC formula implemented by CAMC.
const DefinitionCAMC = "go-modularity/camc-v1"

// CAMC computes Cohesion Among Methods of a Class from the method ×
// parameter-type occurrence matrix M, where M[i][j] = 1 iff method i has at
// least one parameter of distinct type j:
//
//	CAMC = (number of 1-cells in M) / (k × p)
//
// k is the method count (matrix height) and p the number of distinct
// parameter types across all methods (matrix width). Receivers and return
// types are excluded; a duplicated parameter type counts once per method.
// The result is in (0, 1]. Not applicable when k == 0 or p == 0.
func CAMC(oneCells, methodCount, distinctParamTypes int) MetricResult {
	if methodCount == 0 {
		return notApplicable(MetricCAMC, ScopeType, DefinitionCAMC, "type has no methods")
	}

	if distinctParamTypes == 0 {
		return notApplicable(MetricCAMC, ScopeType, DefinitionCAMC,
			"no method has parameters")
	}

	return applicable(MetricCAMC, ScopeType, DefinitionCAMC,
		float64(oneCells)/(float64(methodCount)*float64(distinctParamTypes)))
}
