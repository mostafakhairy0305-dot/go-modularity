package domain

import "github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"

// DocScope groups metrics-guide entries by the kind of entity they
// describe: type metrics, package metrics, or the structural columns that
// are counted rather than computed.
type DocScope string

const (
	// DocScopeType marks a type-level metric entry.
	DocScopeType DocScope = "type"
	// DocScopePackage marks a package-level metric entry.
	DocScopePackage DocScope = "package"
	// DocScopeStructural marks a counted column (Ca, Funcs, Fields, …).
	DocScopeStructural DocScope = "structural"
)

// Doc directions: whether smaller or larger values are better, or no
// universal direction exists. They mirror qualityByMetric.
const (
	DirectionLower   = "lower"
	DirectionHigher  = "higher"
	DirectionNeutral = "neutral"
)

// MetricDoc explains one reported metric or structural field to a human:
// what it means, how it is computed, and how to judge its values. It is
// the single source behind the standalone metrics guide (--help --web) and
// the report page's per-column explanations.
type MetricDoc struct {
	// Name is the metric or column key, e.g. "amc" or "ca".
	Name string
	// Label is the column heading, e.g. "AMC".
	Label string
	// FullName spells the metric out, e.g. "Average Method Complexity".
	FullName string
	// Scope groups the entry: type metric, package metric, or structural.
	Scope DocScope
	// Definition is the versioned formula id; empty for structural fields.
	Definition string
	// FormulaMathML holds display-mode <math> markup. MathML Core only, so
	// browsers typeset it natively and the page stays self-contained.
	// Empty for structural fields.
	FormulaMathML string
	// FormulaLaTeX is the LaTeX source of record behind FormulaMathML.
	FormulaLaTeX string
	// Summary is the one-sentence meaning.
	Summary string
	// HowCalculated spells out the inputs and mechanics.
	HowCalculated string
	// Interpretation explains when values are good or bad, and why.
	Interpretation string
	// NotApplicable states when the metric is n/a; empty means always
	// applicable.
	NotApplicable string
	// Direction is "lower", "higher", or "neutral", matching the quality
	// coloring of the renderers.
	Direction string
	// Bounded reports whether values live in [0, 1].
	Bounded bool
	// Example is a small worked numeric example.
	Example string
}

// LaTeX: \mathit{AMC} = \frac{\sum_{i=1}^{k} cc(m_i)}{k}
//
// LaTeX: cc(m) = 1 + n_{\mathrm{if}} + n_{\mathrm{for}} + n_{\mathrm{range}} + n_{\mathrm{case}} + n_{\mathrm{select}} + n_{\&\&,||}
const formulaAMC = `<math display="block" alttext="AMC = \frac{\sum_{i=1}^{k} cc(m_i)}{k}"><mrow><mi>AMC</mi><mo>=</mo><mfrac><mrow><munderover><mo>∑</mo><mrow><mi>i</mi><mo>=</mo><mn>1</mn></mrow><mi>k</mi></munderover><mi>cc</mi><mo stretchy="false">(</mo><msub><mi>m</mi><mi>i</mi></msub><mo stretchy="false">)</mo></mrow><mi>k</mi></mfrac></mrow></math>
<math display="block" alttext="cc(m) = 1 + n_{if} + n_{for} + n_{range} + n_{case} + n_{select} + n_{\&\&,||}"><mrow><mi>cc</mi><mo stretchy="false">(</mo><mi>m</mi><mo stretchy="false">)</mo><mo>=</mo><mn>1</mn><mo>+</mo><msub><mi>n</mi><mtext>if</mtext></msub><mo>+</mo><msub><mi>n</mi><mtext>for</mtext></msub><mo>+</mo><msub><mi>n</mi><mtext>range</mtext></msub><mo>+</mo><msub><mi>n</mi><mtext>case</mtext></msub><mo>+</mo><msub><mi>n</mi><mtext>select</mtext></msub><mo>+</mo><msub><mi>n</mi><mtext>&amp;&amp;,||</mtext></msub></mrow></math>`

// LaTeX: \mathit{LCOM1} = \max(P - Q,\ 0)
const formulaLCOM1 = `<math display="block" alttext="LCOM1 = \max(P - Q, 0)"><mrow><mi>LCOM1</mi><mo>=</mo><mi>max</mi><mo stretchy="false">(</mo><mi>P</mi><mo>−</mo><mi>Q</mi><mo>,</mo><mn>0</mn><mo stretchy="false">)</mo></mrow></math>`

// LaTeX: \mathit{LCOM96b} = 1 - \frac{a}{f \cdot k}
const formulaLCOM96b = `<math display="block" alttext="LCOM96b = 1 - \frac{a}{f \cdot k}"><mrow><mi>LCOM96b</mi><mo>=</mo><mn>1</mn><mo>−</mo><mfrac><mi>a</mi><mrow><mi>f</mi><mo>⋅</mo><mi>k</mi></mrow></mfrac></mrow></math>`

// LaTeX: \mathit{CAMC} = \frac{\mathrm{ones}(M)}{k \cdot p}
const formulaCAMC = `<math display="block" alttext="CAMC = \frac{ones(M)}{k \cdot p}"><mrow><mi>CAMC</mi><mo>=</mo><mfrac><mrow><mi>ones</mi><mo stretchy="false">(</mo><mi>M</mi><mo stretchy="false">)</mo></mrow><mrow><mi>k</mi><mo>⋅</mo><mi>p</mi></mrow></mfrac></mrow></math>`

// LaTeX: \mathit{TCC} = \frac{c}{k(k-1)/2}
const formulaTCC = `<math display="block" alttext="TCC = \frac{c}{k(k-1)/2}"><mrow><mi>TCC</mi><mo>=</mo><mfrac><mi>c</mi><mrow><mi>k</mi><mo stretchy="false">(</mo><mi>k</mi><mo>−</mo><mn>1</mn><mo stretchy="false">)</mo><mo>/</mo><mn>2</mn></mrow></mfrac></mrow></math>`

// LaTeX: \mathit{CBO}(t) = \lvert R(t) \rvert
const formulaCBO = `<math display="block" alttext="CBO(t) = |R(t)|"><mrow><mi>CBO</mi><mo stretchy="false">(</mo><mi>t</mi><mo stretchy="false">)</mo><mo>=</mo><mo stretchy="false">|</mo><mi>R</mi><mo stretchy="false">(</mo><mi>t</mi><mo stretchy="false">)</mo><mo stretchy="false">|</mo></mrow></math>`

// LaTeX: \mathit{RI} = w_c C + w_k (1 - K) + w_t T + w_d D
//
// LaTeX: C = 1 - \mathit{LCOM96b}
//
// LaTeX: K = \frac{\mathit{CBO}}{\mathit{CBO} + 1}
//
// LaTeX: T = \frac{1}{1 + \max(0,\ \mathit{AMC} - 1)}
//
// LaTeX: D = \frac{\text{documented exported members}}{\text{exported members}}
const formulaReusability = `<math display="block" alttext="RI = w_c C + w_k (1 - K) + w_t T + w_d D"><mrow><mi>RI</mi><mo>=</mo><msub><mi>w</mi><mi>c</mi></msub><mi>C</mi><mo>+</mo><msub><mi>w</mi><mi>k</mi></msub><mo stretchy="false">(</mo><mn>1</mn><mo>−</mo><mi>K</mi><mo stretchy="false">)</mo><mo>+</mo><msub><mi>w</mi><mi>t</mi></msub><mi>T</mi><mo>+</mo><msub><mi>w</mi><mi>d</mi></msub><mi>D</mi></mrow></math>
<math display="block" alttext="C = 1 - LCOM96b"><mrow><mi>C</mi><mo>=</mo><mn>1</mn><mo>−</mo><mi>LCOM96b</mi></mrow></math>
<math display="block" alttext="K = \frac{CBO}{CBO + 1}"><mrow><mi>K</mi><mo>=</mo><mfrac><mi>CBO</mi><mrow><mi>CBO</mi><mo>+</mo><mn>1</mn></mrow></mfrac></mrow></math>
<math display="block" alttext="T = \frac{1}{1 + \max(0, AMC - 1)}"><mrow><mi>T</mi><mo>=</mo><mfrac><mn>1</mn><mrow><mn>1</mn><mo>+</mo><mi>max</mi><mo stretchy="false">(</mo><mn>0</mn><mo>,</mo><mi>AMC</mi><mo>−</mo><mn>1</mn><mo stretchy="false">)</mo></mrow></mfrac></mrow></math>
<math display="block" alttext="D = \frac{documented exported members}{exported members}"><mrow><mi>D</mi><mo>=</mo><mfrac><mtext>documented exported members</mtext><mtext>exported members</mtext></mfrac></mrow></math>`

// LaTeX: A = \frac{N_{\text{interface}}}{N_{\text{named}}}
const formulaAbstractness = `<math display="block" alttext="A = \frac{N_{interface}}{N_{named}}"><mrow><mi>A</mi><mo>=</mo><mfrac><msub><mi>N</mi><mtext>interface</mtext></msub><msub><mi>N</mi><mtext>named</mtext></msub></mfrac></mrow></math>`

// LaTeX: I = \frac{C_e}{C_a + C_e}
const formulaInstability = `<math display="block" alttext="I = \frac{C_e}{C_a + C_e}"><mrow><mi>I</mi><mo>=</mo><mfrac><msub><mi>C</mi><mi>e</mi></msub><mrow><msub><mi>C</mi><mi>a</mi></msub><mo>+</mo><msub><mi>C</mi><mi>e</mi></msub></mrow></mfrac></mrow></math>`

// LaTeX: D = \lvert A + I - 1 \rvert
const formulaDistance = `<math display="block" alttext="D = |A + I - 1|"><mrow><mi>D</mi><mo>=</mo><mo stretchy="false">|</mo><mi>A</mi><mo>+</mo><mi>I</mi><mo>−</mo><mn>1</mn><mo stretchy="false">|</mo></mrow></math>`

// MetricDocs returns the guide entries for every reported metric and
// structural column: type metrics in their fixed order, then package
// metrics, then the counted columns. Direction and Bounded mirror
// qualityByMetric (enforced by test).
func MetricDocs() []MetricDoc {
	return []MetricDoc{
		{
			Name:           metrics.MetricAMC,
			Label:          abbrev(metrics.MetricAMC),
			FullName:       "Average Method Complexity",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionAMC,
			FormulaMathML:  formulaAMC,
			FormulaLaTeX:   `AMC = \frac{\sum_{i=1}^{k} cc(m_i)}{k}` + "\n" + `cc(m) = 1 + n_{if} + n_{for} + n_{range} + n_{case} + n_{select} + n_{\&\&,||}`,
			Summary:        "The mean cyclomatic complexity of the type's methods.",
			HowCalculated:  "Each declared method's cyclomatic complexity cc starts at 1 and grows by one for every if statement, for loop, range loop, non-default case clause of a switch or type switch, non-default select communication clause, and each && or || operator. AMC is the sum of those method complexities divided by the method count k.",
			Interpretation: "An AMC of 1 means the methods are branch-free on average; every extra point is one more decision path per method to understand and test. Because AMC is unbounded, the report colors it relative to the other types in the same column — a high value is a signal to split complex methods or simplify control flow, not an automatic failure.",
			NotApplicable:  "When the type has no methods.",
			Direction:      DirectionLower,
			Bounded:        false,
			Example:        "Methods of complexity 1, 3, and 5 give AMC = (1 + 3 + 5) / 3 = 3.00.",
		},
		{
			Name:           metrics.MetricLCOM1,
			Label:          abbrev(metrics.MetricLCOM1),
			FullName:       "Lack of Cohesion in Methods (LCOM1)",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionLCOM1,
			FormulaMathML:  formulaLCOM1,
			FormulaLaTeX:   `LCOM1 = \max(P - Q,\ 0)`,
			Summary:        "Method pairs that share no fields, less the pairs that do.",
			HowCalculated:  "Every unordered pair of the type's methods is classified: a pair shares when the two methods' field-usage sets intersect (with -field-usage=transitive, usage propagates through calls to sibling methods). P counts the non-sharing pairs, Q the sharing pairs.",
			Interpretation: "0 means sharing pairs at least balance the non-sharing ones. The larger the value, the more method pairs touch disjoint sets of fields — a classic sign that the type bundles unrelated responsibilities that could be split. Unbounded, so it is colored relative to its column.",
			NotApplicable:  "When the type has fewer than 2 methods, or has no fields.",
			Direction:      DirectionLower,
			Bounded:        false,
			Example:        "4 methods form 6 pairs; if 2 pairs share a field and 4 do not, LCOM1 = max(4 − 2, 0) = 2.",
		},
		{
			Name:           metrics.MetricLCOM96b,
			Label:          abbrev(metrics.MetricLCOM96b),
			FullName:       "Lack of Cohesion (LCOM96b)",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionLCOM96b,
			FormulaMathML:  formulaLCOM96b,
			FormulaLaTeX:   `LCOM96b = 1 - \frac{a}{f \cdot k}`,
			Summary:        "Lack of cohesion as the emptiness of the method–field usage matrix.",
			HowCalculated:  "Build the method × field matrix with a 1 wherever a method uses a field; each distinct method–field pair counts once, and a is the number of 1-cells. With f fields and k methods, LCOM96b is one minus the matrix density. This variant is used instead of Henderson-Sellers LCOM* because it stays defined for a single method.",
			Interpretation: "0 means every method uses every field; 1 means no method uses any field. Values near 1 suggest the fields and methods may not belong together. Bounded to 0–1, so it is colored with the fixed thresholds; it also feeds the reusability index's cohesion component.",
			NotApplicable:  "When the type has no fields, or has no methods.",
			Direction:      DirectionLower,
			Bounded:        true,
			Example:        "3 methods over 4 fields with 6 used method–field pairs: LCOM96b = 1 − 6 / (4 × 3) = 0.50.",
		},
		{
			Name:           metrics.MetricCAMC,
			Label:          abbrev(metrics.MetricCAMC),
			FullName:       "Cohesion Among Methods of a Class",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionCAMC,
			FormulaMathML:  formulaCAMC,
			FormulaLaTeX:   `CAMC = \frac{\mathrm{ones}(M)}{k \cdot p}`,
			Summary:        "How much parameter-type vocabulary the methods share.",
			HowCalculated:  "Build the method × parameter-type matrix M with M[i][j] = 1 iff method i takes at least one parameter of distinct type j. Receivers and return types are excluded, and a repeated parameter type counts once per method. CAMC is the count of 1-cells over k methods times p distinct parameter types.",
			Interpretation: "1 means every method uses every parameter type in the type's vocabulary — the methods speak the same language. Low values suggest the methods operate on unrelated inputs and the type may be a grab bag of loosely related operations.",
			NotApplicable:  "When the type has no methods, or no method has parameters.",
			Direction:      DirectionHigher,
			Bounded:        true,
			Example:        "2 methods over 3 distinct parameter types with 4 one-cells: CAMC = 4 / (2 × 3) ≈ 0.67.",
		},
		{
			Name:           metrics.MetricTCC,
			Label:          abbrev(metrics.MetricTCC),
			FullName:       "Tight Class Cohesion",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionTCC,
			FormulaMathML:  formulaTCC,
			FormulaLaTeX:   `TCC = \frac{c}{k(k-1)/2}`,
			Summary:        "The fraction of method pairs connected through a shared field.",
			HowCalculated:  "Uses the same sharing predicate as LCOM1: a pair of methods is connected when their field-usage sets intersect. c counts the connected pairs; the denominator k(k−1)/2 is every possible unordered pair of the k methods.",
			Interpretation: "1 means every pair of methods shares at least one field; 0 means none do. Low TCC signals weak internal relatedness — the methods barely work on the same state, so the type may hide several smaller ones.",
			NotApplicable:  "When the type has fewer than 2 methods.",
			Direction:      DirectionHigher,
			Bounded:        true,
			Example:        "4 methods form 6 pairs; 3 connected pairs give TCC = 3 / 6 = 0.50.",
		},
		{
			Name:           metrics.MetricCBO,
			Label:          abbrev(metrics.MetricCBO),
			FullName:       "Coupling Between Objects",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionCBO,
			FormulaMathML:  formulaCBO,
			FormulaLaTeX:   `CBO(t) = \lvert R(t) \rvert`,
			Summary:        "How many other analyzed named types this type depends on.",
			HowCalculated:  "R(t) is the set of distinct other analyzed named types the type references through its fields, method parameters, method returns, and embedded types; self-references are excluded. Only types inside the current analysis count, so the value is scope-relative: analyzing one package yields lower CBO than analyzing the whole module. CBO is hidden by default and displayed only when selected via -metrics, but it is always computed when reusability needs it.",
			Interpretation: "Each referenced type is a collaborator that can break this one, so lower means fewer reasons to change. Unbounded, so it is colored relative to its column; compare it only across runs with the same patterns and -dependency-scope.",
			NotApplicable:  "",
			Direction:      DirectionLower,
			Bounded:        false,
			Example:        "A type whose fields and method signatures mention 3 distinct analyzed types has CBO = 3.",
		},
		{
			Name:           metrics.MetricReusability,
			Label:          abbrev(metrics.MetricReusability),
			FullName:       "Experimental Reusability Index",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionReusability,
			FormulaMathML:  formulaReusability,
			FormulaLaTeX:   `RI = w_c C + w_k (1 - K) + w_t T + w_d D` + "\n" + `C = 1 - LCOM96b` + "\n" + `K = \frac{CBO}{CBO + 1}` + "\n" + `T = \frac{1}{1 + \max(0,\ AMC - 1)}` + "\n" + `D = \frac{\text{documented exported members}}{\text{exported members}}`,
			Summary:        "An experimental composite of cohesion, coupling, testability, and documentation.",
			HowCalculated:  "Four normalized 0–1 components combine with default weights w_c = 0.35 (cohesion C), w_k = 0.25 (coupling K, contributing 1 − K), w_t = 0.25 (testability T), and w_d = 0.15 (documentation D). A component whose input is not applicable is dropped and the remaining weights are renormalized to sum to 1, keeping the index in 0–1; dropped components are listed in the metric's reason.",
			Interpretation: "A high index combines cohesive methods, few collaborators, simple control flow, and a documented exported surface — the properties that make a type safe to lift out and reuse elsewhere. It is experimental: treat it as a triage hint and read its four inputs before acting on it.",
			NotApplicable:  "Only when every weighted component is dropped — for example a type with no methods, no fields, and no exported members.",
			Direction:      DirectionHigher,
			Bounded:        true,
			Example:        "C = 0.8, 1 − K = 0.75, T = 0.5, D = 1.0 with default weights: RI = 0.35·0.8 + 0.25·0.75 + 0.25·0.5 + 0.15·1.0 ≈ 0.74.",
		},
		{
			Name:           metrics.MetricAbstractness,
			Label:          abbrev(metrics.MetricAbstractness),
			FullName:       "Abstractness",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionAbstractness,
			FormulaMathML:  formulaAbstractness,
			FormulaLaTeX:   `A = \frac{N_{\text{interface}}}{N_{\text{named}}}`,
			Summary:        "The share of a package's named types that are interfaces.",
			HowCalculated:  "The package's named interface types divided by all of its relevant named types; type aliases are excluded from both counts.",
			Interpretation: "Neither end is universally good, so the report leaves it uncolored. High abstractness is expected for contract and API packages; low abstractness is normal for implementation packages. Its real job is serving as an input to distance.",
			NotApplicable:  "When the package declares no relevant named types.",
			Direction:      DirectionNeutral,
			Bounded:        true,
			Example:        "2 interfaces among 8 named types: A = 2 / 8 = 0.25.",
		},
		{
			Name:           metrics.MetricInstability,
			Label:          abbrev(metrics.MetricInstability),
			FullName:       "Instability",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionInstability,
			FormulaMathML:  formulaInstability,
			FormulaLaTeX:   `I = \frac{C_e}{C_a + C_e}`,
			Summary:        "How much a package depends outward versus being depended on.",
			HowCalculated:  "Ca counts analyzed packages that import this one; Ce counts this package's imports within the configured -dependency-scope. An isolated package (Ca + Ce = 0) is defined as maximally stable — instability 0 — with the convention noted in the metric's reason.",
			Interpretation: "Uncolored because neither end is universally good: 0 means only incoming dependents (stable, hard to change safely), 1 means only outgoing dependencies (easy to change, nothing relies on it). Core packages usually want low instability while adapters at the application edge can reasonably be high; the balance is judged by distance.",
			NotApplicable:  "",
			Direction:      DirectionNeutral,
			Bounded:        true,
			Example:        "3 analyzed importers and 1 in-scope import: I = 1 / (3 + 1) = 0.25.",
		},
		{
			Name:           metrics.MetricDistance,
			Label:          abbrev(metrics.MetricDistance),
			FullName:       "Distance from the Main Sequence",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionDistance,
			FormulaMathML:  formulaDistance,
			FormulaLaTeX:   `D = \lvert A + I - 1 \rvert`,
			Summary:        "How far a package sits from the ideal abstractness–instability balance.",
			HowCalculated:  "The absolute distance of the package's abstractness A and instability I from the 'main sequence' line A + I = 1, where abstraction and stability balance.",
			Interpretation: "0 is on the main sequence. High distance means the package is either concrete and stable (rigid — everything depends on its details) or abstract and unstable (abstractions nobody depends on). Mind the isolated-package convention: a concrete isolated package has A = 0 and I = 0, so D = 1 by definition, not necessarily by design fault.",
			NotApplicable:  "When abstractness or instability is not applicable — for example, the package declares no relevant named types.",
			Direction:      DirectionLower,
			Bounded:        true,
			Example:        "A = 0.25 and I = 0.5: D = |0.25 + 0.5 − 1| = 0.25.",
		},
		{
			Name:           "ca",
			Label:          "Ca",
			FullName:       "Afferent coupling",
			Scope:          DocScopeStructural,
			Summary:        "How many analyzed packages import this package.",
			HowCalculated:  "Counted within the analyzed set only — importers outside the analysis are not observable, so the value depends on the patterns you analyze.",
			Interpretation: "A neutral count with no good/bad color. High Ca marks load-bearing packages: many others break when this one changes, so it should be stable and well tested. It is the incoming half of instability.",
			Direction:      DirectionNeutral,
			Example:        "If 3 analyzed packages import example.com/m/util, its Ca is 3.",
		},
		{
			Name:           "ce",
			Label:          "Ce",
			FullName:       "Efferent coupling",
			Scope:          DocScopeStructural,
			Summary:        "How many packages this package imports, within the dependency scope.",
			HowCalculated:  "The package's imports that fall in the configured -dependency-scope: project counts only other analyzed packages, module counts packages of the main module, all counts every import. Duplicates and self-imports are ignored.",
			Interpretation: "A neutral count with no good/bad color. High Ce means the package has many reasons to change. It is the outgoing half of instability.",
			Direction:      DirectionNeutral,
			Example:        "A package importing 2 in-scope packages has Ce = 2 regardless of how often each is imported.",
		},
		{
			Name:           "funcs",
			Label:          "Funcs",
			FullName:       "Functions",
			Scope:          DocScopeStructural,
			Summary:        "Declared functions and methods in the package.",
			HowCalculated:  "Counted over the package's analyzed files — excluded files (tests or generated code, unless included by flag) do not contribute.",
			Interpretation: "A neutral size measure: use it to weigh the metrics — a package with 3 funcs and a package with 300 deserve different scrutiny at the same scores.",
			Direction:      DirectionNeutral,
			Example:        "A package with 4 functions and 6 methods across its types shows Funcs = 10.",
		},
		{
			Name:           "types",
			Label:          "Types",
			FullName:       "Named types",
			Scope:          DocScopeStructural,
			Summary:        "Analyzed named types declared in the package.",
			HowCalculated:  "Counts the package's named type declarations that enter the analysis; type aliases never enter the model.",
			Interpretation: "A neutral size measure, shown in the Packages view. Many types with poor cohesion scores is a stronger signal than one outlier.",
			Direction:      DirectionNeutral,
			Example:        "A package declaring Service, Config, and an Option interface shows Types = 3.",
		},
		{
			Name:           "fields",
			Label:          "Fields",
			FullName:       "Struct fields",
			Scope:          DocScopeStructural,
			Summary:        "The type's struct field count.",
			HowCalculated:  "An embedded field counts as one; members promoted through embedding are not counted. Non-struct types show 0.",
			Interpretation: "A neutral count that sizes the cohesion metrics: LCOM1, LCOM96b, and TCC all reason about how methods use these fields.",
			Direction:      DirectionNeutral,
			Example:        "struct { ID int; Name string; sync.Mutex } has Fields = 3 — the embedded mutex counts as one.",
		},
		{
			Name:           "methods",
			Label:          "Methods",
			FullName:       "Declared methods",
			Scope:          DocScopeStructural,
			Summary:        "The type's declared method count.",
			HowCalculated:  "Value- and pointer-receiver methods are counted alike; methods promoted from embedded types are excluded.",
			Interpretation: "A neutral count that sizes the cohesion and complexity metrics: most of them are n/a below 1 or 2 methods, by design rather than as a gap.",
			Direction:      DirectionNeutral,
			Example:        "A type with func (s *S) Open() and func (s S) Close() has Methods = 2.",
		},
	}
}
