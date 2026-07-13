# go-modularity

`go-modularity` analyzes Go packages and reports type-level and package-level
modularity metrics. It can render a readable terminal table, a stable JSON
schema, flat CSV rows for scripts, or an interactive single-file HTML report.

## Install

Requires Go 1.26.5 or newer.

Install from the module path:

```sh
go install github.com/mostafakhairy0305-dot/go-modularity/cmd/go-modularity@latest
```

Install from a local checkout:

```sh
git clone https://github.com/mostafakhairy0305-dot/go-modularity.git
cd go-modularity
go install ./cmd/go-modularity
```

Build a local binary without installing it:

```sh
go build -o ./bin/go-modularity ./cmd/go-modularity
```

Verify the binary:

```sh
go-modularity -version
```

## Quick Start

Analyze the current module:

```sh
go-modularity ./...
```

Write JSON to a file:

```sh
go-modularity -format=json -output=report.json ./...
```

Generate CSV for spreadsheets or CI artifacts:

```sh
go-modularity -format=csv -output=report.csv ./...
```

Analyze only internal packages:

```sh
go-modularity ./internal/...
```

Logs are written to stderr. Reports are written to stdout unless `-output` is
set.

## Usage

```text
go-modularity [flags] [patterns...]
```

`patterns` are Go package patterns such as `./...`, `./internal/...`, or a
single package path. If no pattern is provided, the command uses `./...`.

Put flags before package patterns:

```sh
go-modularity -format=json ./...
```

## Options

| Option | Meaning | Default | Example |
| --- | --- | --- | --- |
| `patterns...` | Package patterns passed to the Go package loader. Empty means all packages below the current directory. | `./...` | `go-modularity ./internal/...` |
| `-format` | Output format. Accepted values are `text`, `json`, `csv`, and `web`. | `text` | `go-modularity -format=json ./...` |
| `-web` | Shorthand for `-format=web`. Without `-output` the report is written to `modularity-report.html` and opened in the default browser (only when stdout is a terminal). Combining it with a different explicit `-format` is a usage error. Combined with `--help` it opens the metrics guide instead (see below). | `false` | `go-modularity -web ./...` |
| `-output` | Write the report to a file instead of stdout. | stdout | `go-modularity -output=modularity.txt ./...` |
| `-explain` | In text output, include notes for not-applicable metrics and dropped reusability components. | `false` | `go-modularity -explain ./...` |
| `-metrics` | Comma-separated display metric list. Dependencies needed to compute selected metrics are computed automatically but not shown unless selected. | all except `cbo` | `go-modularity -metrics=amc,lcom1,tcc ./...` |
| `-workers` | Number of concurrent package workers. `0` chooses `min(GOMAXPROCS, packageCount)`. | `0` | `go-modularity -workers=4 ./...` |
| `-field-usage` | Field-use resolution for cohesion metrics. `direct` counts only fields directly read by a method. `transitive` also propagates through calls to sibling methods of the same type. | `direct` | `go-modularity -field-usage=transitive ./...` |
| `-dependency-scope` | Import edge scope used by package coupling metrics. `project` counts only analyzed packages, `module` counts packages in the same module, and `all` counts standard-library and external imports too. | `module` | `go-modularity -dependency-scope=project ./...` |
| `-build-tags` | Comma-separated Go build tags used while loading packages. | none | `go-modularity -build-tags=integration,linux ./...` |
| `-tests` | Include test files and test packages in analysis. | `false` | `go-modularity -tests ./...` |
| `-generated` | Include generated files that contain the standard `Code generated ... DO NOT EDIT.` marker. | `false` | `go-modularity -generated ./...` |
| `-continue-on-error` | Continue past packages that fail to load or type-check. Skipped packages are not included in the report. | `false` | `go-modularity -continue-on-error ./...` |
| `-cpu-profile` | Write a CPU profile while the analysis runs. | off | `go-modularity -cpu-profile=cpu.prof ./...` |
| `-memory-profile` | Write a heap profile after analysis completes. | off | `go-modularity -memory-profile=heap.prof ./...` |
| `-verbose` | Enable debug logging to stderr. | `false` | `go-modularity -verbose ./...` |
| `-version` | Print the version and exit. | off | `go-modularity -version` |
| `-check` | Enforce a modularity policy (see [Policy Checks](#policy-checks)). Loads `.modularity.yml` if present, otherwise the recommended defaults. Any violation exits `3`. | `false` | `go-modularity -check ./...` |
| `-config` | Path to a policy config file. Implies `-check` and skips auto-discovery. | auto-discover `.modularity.yml` | `go-modularity -config=policy.yml ./...` |
| `-max` | Set an upper-bound condition `key=value`; repeatable. `key` is a structural field, metric name, or scoped metric key such as `type.amc` or `package.distance`. Implies `-check`. | none | `go-modularity -max=type.amc=5 ./...` |
| `-min` | Set a lower-bound condition `key=value`; repeatable. Implies `-check`. | none | `go-modularity -min=type.reusability=0.5 ./...` |

Terminal text output uses color only when stdout is a terminal. Set `NO_COLOR=1`
to disable color:

```sh
NO_COLOR=1 go-modularity ./...
```

## Output Formats

### Text

Text output is a tree table. Package paths are grouped by module-relative
directories, package rows show package-level metrics, and type rows show
type-level metrics. Directory and package rows also show means for type metrics
below them.

```sh
go-modularity -format=text -explain ./...
```

Important text conventions:

| Text field | Meaning |
| --- | --- |
| `PATH / TYPE` | A package, directory group, or type name. |
| Package or directory row | Shows package-level metrics, then means of type-level metrics under that row. Means ignore not-applicable values. |
| Type row | Shows metrics for one named type. Package-level columns are blank on type rows. |
| `Abst` | `abstractness`, a package-level metric. |
| `Inst` | `instability`, a package-level metric. |
| `Dist` | `distance`, a package-level metric. |
| `AMC` | `amc`, a type-level metric. |
| `LCOM1` | `lcom1`, a type-level metric. |
| `LCOM96b` | `lcom96b`, a type-level metric. |
| `CAMC` | `camc`, a type-level metric. |
| `TCC` | `tcc`, a type-level metric. |
| `CBO` | `cbo`, a type-level metric. Only shown when selected. |
| `Reuse` | `reusability`, a type-level metric. |
| dash | Not applicable. Use `-explain` to see why. |

Text colors are hints, not hard pass/fail rules. For bounded metrics with a
clear direction, green is generally better, yellow is mixed, and red deserves a
look. For unbounded type metrics such as `amc`, `lcom1`, and `cbo`, coloring is
relative to the range of values in the current package or subtree.

### JSON

JSON output is indented and versioned.

```sh
go-modularity -format=json ./... > report.json
```

Shape:

```json
{
  "schema_version": "2",
  "tool": {
    "name": "go-modularity",
    "version": "dev"
  },
  "packages": [
    {
      "path": "example.com/module/pkg",
      "afferent": 3,
      "efferent": 2,
      "funcs": 14,
      "metrics": {
        "abstractness": {
          "scope": "package",
          "value": 0,
          "applicable": true,
          "definition": "go-modularity/abstractness-v1"
        }
      },
      "types": [
        {
          "name": "Service",
          "fields": 4,
          "methods": 6,
          "metrics": {
            "amc": {
              "scope": "type",
              "value": 2,
              "applicable": true,
              "definition": "go-modularity/amc-v1"
            }
          }
        }
      ]
    }
  ]
}
```

JSON field meanings:

| Field | Meaning |
| --- | --- |
| `schema_version` | Output schema version. Current value is `2` (version 2 added the structural facts below). |
| `tool.name` | Tool name, always `go-modularity`. |
| `tool.version` | Build version embedded at link time. Local development builds usually report `dev`. |
| `packages` | Analyzed packages, sorted by import path so repeated runs are stable. |
| `packages[].path` | Full Go import path of the package. Use this as the package identifier. |
| `packages[].afferent` | Ca: how many analyzed packages import this one. Always measured within the analyzed set. |
| `packages[].efferent` | Ce: how many packages this one imports, counted under the configured `-dependency-scope`. |
| `packages[].funcs` | Number of declared functions and methods in the package's analyzed files. |
| `packages[].metrics` | Package-level metric object keyed by metric name. These metrics describe the package as a whole. |
| `packages[].types` | Named types in the package, sorted by type name. Each entry carries the selected type-level metrics for that type. |
| `packages[].types[].name` | Declared type name, such as `Service` or `Repository`. |
| `packages[].types[].fields` | Struct field count; an embedded field counts as one. `0` for non-struct types. |
| `packages[].types[].methods` | Declared method count (promoted methods excluded). |
| `packages[].types[].metrics` | Type-level metric object keyed by metric name. These metrics describe that type only. |
| `metrics.<name>.scope` | Entity kind for the metric: `package` for package metrics or `type` for type metrics. |
| `metrics.<name>.value` | Numeric score. Present only when `applicable` is `true`; absent when the metric cannot be calculated honestly. |
| `metrics.<name>.applicable` | Whether `value` can be read. `false` means the inputs do not make sense for that metric, for example a type with no methods for `amc`. |
| `metrics.<name>.reason` | Explanation for not-applicable metrics, isolated-package conventions, or dropped reusability components. Omitted when empty. |
| `metrics.<name>.definition` | Versioned formula identifier, such as `go-modularity/amc-v1`. Use it to detect formula changes across tool versions. |

Metric objects are keyed by metric name in the fixed render order. A missing
metric usually means it was not selected by `-metrics`; a present metric with
`"applicable": false` means it was selected but could not be calculated for
that entity.

### Web

Web output is one self-contained HTML file: the versioned JSON report is
embedded in the page and rendered by inline vanilla JavaScript — no external
requests, no server, and it works offline straight from `file://`.

```sh
go-modularity -web ./...
go-modularity -format=web -output=report.html ./...
```

Without `-output` the report is written to `modularity-report.html` and, when
stdout is a terminal, opened in the default browser. What the page offers:

| Feature | Meaning |
| --- | --- |
| Views | `Tree` (the same directory tree as the text report: nested directories with compressed single-child chains, collapsible at every level, directory and package rows carrying subtree means, plus one-click depth presets `1…N`/`All`), `Types` (flat table of every type with its field and method counts), and `Packages` (flat table with `Ca`/`Ce` coupling, function and type counts, and package metrics). |
| Sorting | In the `Types` and `Packages` views, click any column header to sort ascending/descending; not-applicable values always sort last. The tree view keeps the fixed path order of the text report. |
| Filtering | Live text search (press `/`), a package dropdown, a hide-n/a-rows toggle, and a column show/hide picker. |
| Values | Each cell shows the value with a mini-bar (bounded metrics absolute 0–1, unbounded relative to their column maximum). Values and bars are colored green (good), orange (mixed), or red (review) with the same quality rules as the terminal renderer — fixed thresholds for bounded metrics, column-range-relative for unbounded ones; `abstractness` and `instability` stay neutral. Hover a value or `–` for the full metric name, reason, and versioned definition. |
| Theme | Monochrome black/white chrome — green/orange/red are reserved for metric quality. The toggle switches black-on-white and white-on-black; the default follows the system color scheme. |
| Motion | Row and bar animations respect `prefers-reduced-motion`. |
| Explanations | Every documented column header carries a `?` button opening an info sheet with the metric's typeset formula, meaning, good/bad direction, and n/a conditions — the same content as the metrics guide below. |

The embedded payload wraps the same schema as `-format=json` (adding only the
module path), so anything you read from the JSON format applies to the web
report too.

#### Metrics guide (`--help --web`)

```sh
go-modularity --help --web
```

Combining `--help` and `--web` (either order) writes a self-contained
metrics guide to the OS temp directory (`go-modularity-help-*.html`), logs
the path, and — when stdout is a terminal — opens it in the default
browser. The guide explains every reported field: the formula as native
MathML (authored from LaTeX, no math library needed), how it is calculated,
what its values mean, when they are good or bad and why, the exact color
thresholds, n/a conventions, and a worked example each. Unlike plain
`--help` (a usage error, exit `2`), `--help --web` exits `0`.

### CSV

CSV output has one row per package metric and one row per type metric.

```sh
go-modularity -format=csv -output=report.csv ./...
```

CSV columns:

| Column | Meaning |
| --- | --- |
| `package` | Full Go import path of the package. This identifies the package for both package and type rows. |
| `type` | Type name for type-level metrics. Empty for package-level metrics. |
| `metric` | Metric name, such as `amc`, `reusability`, or `distance`. |
| `scope` | `type` or `package`; this tells you whether the row describes a type or a package. |
| `value` | Numeric score. Empty when `applicable` is `false`. Interpret it using the metric tables below. |
| `applicable` | `true` when `value` is valid, otherwise `false`. |
| `reason` | Explanation for not-applicable metrics, isolated-package conventions, or dropped components. Empty when there is no note. |
| `definition` | Versioned formula identifier. Use this when comparing reports across tool versions. |

## Metrics

Default output includes every metric except `cbo`. `cbo` is still computed when
needed by `reusability`, but it is displayed only when selected explicitly.

```sh
go-modularity -metrics=cbo,reusability ./...
```

### How to Read Values

No metric is a universal verdict by itself. Use them to find code worth
reviewing, then compare nearby packages and types that have similar
responsibilities.

General interpretation:

| Value pattern | Meaning |
| --- | --- |
| `applicable=false` | The metric should not be read for that entity. For example, cohesion over method pairs is not meaningful for a type with fewer than two methods. |
| Empty CSV `value` or missing JSON `value` | Same as `applicable=false`. The tool avoids fake zeros for metrics that cannot be calculated. |
| `0` on a lower-is-better metric | Usually best or least risky for that metric. |
| `1` on a 0..1 higher-is-better metric | Usually best for that metric. |
| `1` on a 0..1 lower-is-better metric | Usually worst for that metric. |
| Large unbounded values | Compare them against similar types in the same package; large `amc`, `lcom1`, or `cbo` values are review signals rather than automatic failures. |

The text renderer colors bounded metrics with fixed thresholds. Higher-is-better
metrics are green at about `0.66` and above, yellow from about `0.33`, and red
below that. Lower-is-better metrics invert that scale. Unbounded type metrics
are colored relative to the local column range.

Type-level metrics:

| Metric | How it is calculated | Value meaning | Good / bad signal |
| --- | --- | --- | --- |
| `amc` | `sum(method cyclomatic complexity) / methodCount`. Cyclomatic complexity starts at 1 and increases for branches. | Unbounded. `1` means methods are branch-free on average. Larger values mean methods contain more decision paths. Not applicable for types with no methods. | Lower is usually easier to test and understand. High values suggest splitting complex methods or simplifying control flow. |
| `lcom1` | `max(nonSharingMethodPairs - sharingMethodPairs, 0)`. Two methods share when they touch at least one common field. | Unbounded integer-like score. `0` means sharing pairs are at least as common as non-sharing pairs. Larger values mean more method pairs do not share fields. Not applicable with fewer than two methods or no fields. | Lower is better. High values suggest the type may contain unrelated responsibilities. |
| `lcom96b` | `1 - methodFieldAccesses / (fieldCount * methodCount)`. Each method-field pair counts once when the method uses the field. | Range `0..1`. `0` means every method uses every field. `1` means no recorded method-field usage. Not applicable with no fields or no methods. | Lower is more cohesive. High values suggest fields and methods may not belong together. |
| `camc` | `oneCells / (methodCount * distinctParameterTypeCount)` over the method by parameter-type matrix. | Range `(0..1]` when applicable. Higher values mean methods share more parameter-type vocabulary. Not applicable with no methods or no method parameters. | Higher is more cohesive. Low values suggest methods may operate on unrelated inputs. |
| `tcc` | `connectedMethodPairs / totalPossibleMethodPairs`. A pair is connected when methods share at least one field. | Range `0..1`. `0` means no method pair shares fields. `1` means every method pair shares at least one field. Not applicable with fewer than two methods. | Higher is more cohesive. Low values suggest weak internal relatedness. |
| `cbo` | Count of distinct other analyzed named types referenced by fields, method parameters, method returns, and embedded types. Self-references are excluded. | Unbounded count. `0` means the type does not reference other analyzed named types. Larger values mean broader type coupling. | Lower is less coupled. High values suggest the type depends on many collaborators or abstractions. |
| `reusability` | Weighted composite of cohesion, coupling, testability, and documentation. Defaults: cohesion `0.35`, coupling `0.25`, testability `0.25`, documentation `0.15`. Dropped components are removed and remaining weights are renormalized. | Range `0..1` when applicable. Higher values combine stronger cohesion, lower saturated coupling, simpler methods, and better exported-member documentation. | Higher is better. Low values are a review signal for coupling, complexity, cohesion, or documentation gaps. |

Reusability components:

| Component | Calculation | Meaning |
| --- | --- | --- |
| Cohesion | `1 - lcom96b` | Rewards types whose methods use their fields cohesively. Dropped when `lcom96b` is not applicable. |
| Coupling | `1 - (cbo / (cbo + 1))` | Rewards lower coupling while saturating large `cbo` values. Always applicable because `cbo` is always applicable. |
| Testability | `1 / (1 + max(0, amc - 1))` | Rewards low average method complexity. Dropped when `amc` is not applicable. |
| Documentation | `documentedExportedMembers / exportedMembers` | Rewards documented exported surface area. Dropped when a type has no exported members. |

Package-level metrics:

| Metric | How it is calculated | Value meaning | Good / bad signal |
| --- | --- | --- | --- |
| `abstractness` | `namedInterfaceTypes / totalRelevantNamedTypes`. Type aliases are excluded. | Range `0..1`. `0` means the package exposes only concrete named types. `1` means all relevant named types are interfaces. Not applicable when the package declares no relevant named types. | Neither high nor low is universally good. High abstractness is expected for contract/API packages; low abstractness is normal for implementation packages. |
| `instability` | `Ce / (Ca + Ce)`, where `Ce` is outgoing dependencies in the selected dependency scope and `Ca` is analyzed packages that import this package. | Range `0..1`. `0` means maximally stable or isolated. `1` means only outgoing dependencies were observed. Isolated packages are defined as `0` with a reason note. | Context dependent. Core packages usually want lower instability; adapter or application-edge packages can reasonably be higher. |
| `distance` | `abs(abstractness + instability - 1)`. | Range `0..1`. `0` is on the main sequence. Larger values are farther away. Not applicable if abstractness is not applicable. | Lower is better for package balance. High values often mean a package is concrete but unstable, or abstract but stable. |

Metric dependencies:

| Selected metric | Also computed |
| --- | --- |
| `reusability` | `lcom96b`, `amc`, `cbo` |
| `distance` | `abstractness`, `instability` |

### Scope and Conventions

A few metrics depend on which packages are in scope or fall back to a defined
convention. Keep these in mind when comparing runs:

- **`cbo` and `reusability` are scope-relative.** `cbo` counts only references
  to types that are *part of the current analysis*. Analyzing a single package
  (`./internal/foo`) yields lower `cbo` — and therefore a different
  `reusability` — than analyzing the whole module (`./...`), because fewer types
  are in scope. Compare `cbo`/`reusability` only across runs with the same
  patterns and `-dependency-scope`.
- **Isolated packages land in "distance = 1".** A package with no in-scope
  dependencies either way (no analyzed importers and no in-scope imports) is
  defined as `instability = 0` (maximally stable, with a reason note). For a
  concrete package (`abstractness = 0`) that makes `distance = |0 + 0 − 1| = 1`.
  A leaf `util` or a `main` analyzed on its own will therefore show the maximum
  distance; this is a convention, not necessarily a design problem.
- **Transitive field usage follows only direct sibling calls.**
  `-field-usage=transitive` propagates a method's field usage through direct
  `x.Method()` calls to sibling methods of the same type. Calls made through
  method expressions (`T.Method`), interface values, or stored function values
  are not followed, so transitive cohesion is a lower bound on true reachability.

## Policy Checks

`go-modularity` can **fail** a run when a package or type crosses a threshold,
so the same metrics that inform a review can gate a CI pipeline. A policy is a
set of **conditions** — budgets on structural facts and bounds on metrics —
evaluated against the report. Any violation prints a summary to stderr and
exits `3`; the report itself still goes to stdout.

Policy checks are **opt-in**. A committed `.modularity.yml` does nothing on its
own; a check runs only when you pass a policy flag:

```sh
go-modularity -check ./...                  # .modularity.yml if present, else defaults
go-modularity -config=policy.yml ./...      # an explicit file (implies -check)
go-modularity -check -max=type.amc=5 ./...   # defaults, with one bound overridden
```

The effective policy is layered, later winning: **recommended defaults → config
file → `-max`/`-min` flags**. Gated metrics are added to the display set
automatically, so a metric you gate on is always computed and shown. A
condition on a metric is skipped wherever that metric is not applicable (for
example `tcc` on a one-method type), so n/a cells never fail a build.

Every check logs its outcome and the policy it used to stderr, so a run is never
a silent no-op — for example `policy check passed source=.modularity.yml` or
`policy check failed source="recommended defaults" violations=15`. A discovered
`.modularity.yml` (like this repository's) is what `-check` uses unless you pass
`-config`, so the log line tells you whether the defaults or a file were
applied.

### Conditions

Every field can carry a `max`, a `min`, or both. Structural budgets are
per-package or per-type counts; metric bounds follow each metric's good/bad
direction. CLI overrides can use bare legacy metric keys (`amc`) or scoped keys
(`type.amc`, `package.distance`); scoped keys are preferred because they match
the report's package/type split.

| Key | Scope | Typical bound | Caps |
| --- | --- | --- | --- |
| `types` | package | `max` | named types per package |
| `exported_funcs` | package | `max` | exported functions and methods per package |
| `unexported_funcs` | package | `max` | unexported functions and methods per package |
| `afferent` | package | `max` | incoming coupling, `Ca` |
| `efferent` | package | `max` | outgoing coupling, `Ce` |
| `fields` | type | `max` | struct fields per type |
| `methods` | type | `max` | declared methods per type |
| `amc` | type | `max` | average method complexity |
| `lcom1` | type | `max` | non-cohesive method pairs |
| `lcom96b` | type | `max` | lack of cohesion `0..1` |
| `camc` | type | `min` | cohesion among parameter types |
| `tcc` | type | `min` | tight class cohesion |
| `cbo` | type | `max` | coupling between objects |
| `reusability` | type | `min` | composite reusability index |
| `abstractness` | package | either | interface ratio (no default) |
| `instability` | package | either | `Ce / (Ca + Ce)` (no default) |
| `distance` | package | `max` | distance from the main sequence |

### `.modularity.yml`

The file mirrors these keys. Each field takes a `max`, a `min`, or both; a bare
number is shorthand for `max`. Unknown keys and unsupported versions are errors,
so a typo fails fast rather than silently disabling a check.

```yaml
---
version: 1

package:
  types:
    max: 12
  exported_funcs:
    max: 12
  unexported_funcs:
    max: 18
  efferent:
    max: 10
  metrics:
    distance:
      max: 0.5

type:
  fields:
    max: 8
  methods:
    max: 10
  metrics:
    amc:
      max: 3
    lcom1:
      max: 10
    lcom96b:
      max: 0.5
    camc:
      min: 0.5
    tcc:
      min: 0.5
    cbo:
      max: 6
    reusability:
      min: 0.7
```

For backward compatibility, a top-level `metrics:` map is still accepted as a
legacy/global form and applies to whichever scope emits that metric. New config
files should prefer `package.metrics` and `type.metrics`, which also lets the
loader reject a type-only metric under the package section.

### Recommended defaults

`-check` with no config file applies the strict recommended baseline (the
block above). It gates the robust, low-false-positive fields and follows each
metric's direction: lower-is-better metrics take a `max`, higher-is-better
metrics take a `min`. `abstractness` and `instability` have no universal good
direction and `afferent` is context dependent, so they are left unconstrained —
opt into them explicitly. `distance` is included but note that isolated
packages (a lone `main` or leaf `util`) sit at `distance = 1` by convention;
loosen or drop it if your analysis scope makes that common.

The repository ships a `.modularity.yml` tuned to pass its own code as a
working, fully commented template.

### In CI

```sh
# Fails the job (exit 3) on any policy violation.
go-modularity -check ./...

# Save the machine-readable report and still gate.
go-modularity -check -format=json -output=modularity.json ./...
```

## Common Examples

Report only cohesion metrics:

```sh
go-modularity -metrics=lcom1,lcom96b,camc,tcc ./...
```

Compare package architecture only:

```sh
go-modularity -metrics=abstractness,instability,distance ./...
```

Analyze test code too:

```sh
go-modularity -tests ./...
```

Use transitive field usage for cohesion:

```sh
go-modularity -field-usage=transitive -metrics=lcom1,lcom96b,tcc ./...
```

Explore the report interactively in a browser:

```sh
go-modularity -web ./...
```

Run in CI and save machine-readable output:

```sh
go-modularity -format=json -output=modularity-report.json -continue-on-error ./...
```

Gate CI on a modularity policy (exit `3` on violations):

```sh
go-modularity -check ./...
```

Profile the analyzer itself:

```sh
go-modularity -cpu-profile=cpu.prof -memory-profile=heap.prof ./...
```

## Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | Success. With `-check`, the report also satisfied every policy condition. |
| `1` | Analysis, profiling, or report writing failed. |
| `2` | Command-line usage error, such as an invalid flag, output format, or policy configuration. Plain `--help` exits `2`; `--help --web` writes the metrics guide and exits `0`. |
| `3` | Policy violations were found (`-check`). The report is on stdout; the violation summary is on stderr. |
| `130` | The run was cancelled by a signal (`Ctrl-C` / `SIGTERM`). |
