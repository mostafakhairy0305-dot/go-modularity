package policyconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"go.yaml.in/yaml/v4"
)

// FileName is the conventional policy config file name.
const FileName = ".modularity.yml"

// documentDecoder is the narrow YAML-decoding behavior needed after a policy
// file has been opened. Keeping file access separate from document decoding
// makes the parser reusable with any compatible decoder.
type documentDecoder interface {
	// KnownFields enables rejection of unknown fields in typed YAML mappings.
	KnownFields(bool)
	// Decode reads the next YAML document into value.
	Decode(value any) error
}

var _ documentDecoder = (*yaml.Decoder)(nil)

// Discover reports the policy config path in dir, if a regular file exists.
// An empty dir means the current working directory.
func Discover(dir string) (string, bool) {
	if dir == "" {
		dir = "."
	}

	path := filepath.Join(dir, FileName)

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", false
	}

	return path, true
}

// Load reads path, decodes it, and returns the validated policy. Unknown keys,
// unsupported versions, and malformed YAML are errors, each prefixed with the
// file path.
func Load(path string) (domain.Policy, error) {
	dir, name := filepath.Split(path)
	if dir == "" {
		dir = "."
	}

	if name == "" {
		return domain.Policy{}, fmt.Errorf("%s: path is a directory", path)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return domain.Policy{}, err
	}
	defer func() { _ = root.Close() }()

	file, err := root.Open(name)
	if err != nil {
		return domain.Policy{}, err
	}
	defer func() { _ = file.Close() }()

	return decodePolicy(path, yaml.NewDecoder(file))
}

// decodePolicy decodes and validates one policy document from an already-open
// source. path is used only to qualify user-facing errors.
func decodePolicy(path string, decoder documentDecoder) (domain.Policy, error) {
	decoder.KnownFields(true)

	var doc fileDTO
	if err := decoder.Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return domain.Policy{}, fmt.Errorf("%s: config is empty", path)
		}

		return domain.Policy{}, fmt.Errorf("%s: %w", path, err)
	}

	policy, err := doc.toPolicy()
	if err != nil {
		return domain.Policy{}, fmt.Errorf("%s: %w", path, err)
	}

	if err := domain.Validate(policy); err != nil {
		return domain.Policy{}, fmt.Errorf("%s: %w", path, err)
	}

	return policy, nil
}

// fileDTO mirrors the .modularity.yml schema. Structural sections use explicit
// fields so KnownFields rejects typos; metric maps are open and validated by the
// domain against the real metric names and scopes.
type fileDTO struct {
	Version int                 `yaml:"version"` // schema version; must be 1
	Package packageDTO          `yaml:"package"` // per-package structural budgets
	Type    typeDTO             `yaml:"type"`    // per-type structural budgets
	Metrics map[string]limitDTO `yaml:"metrics"` // legacy/global metric bounds
}

type packageDTO struct {
	Types           *limitDTO           `yaml:"types"`            // named types per package
	ExportedFuncs   *limitDTO           `yaml:"exported_funcs"`   // exported functions and methods
	UnexportedFuncs *limitDTO           `yaml:"unexported_funcs"` // unexported functions and methods
	Afferent        *limitDTO           `yaml:"afferent"`         // incoming coupling (Ca)
	Efferent        *limitDTO           `yaml:"efferent"`         // outgoing coupling (Ce)
	Metrics         map[string]limitDTO `yaml:"metrics"`          // package metric bounds
}

type typeDTO struct {
	Fields  *limitDTO           `yaml:"fields"`  // struct fields per type
	Methods *limitDTO           `yaml:"methods"` // declared methods per type
	Metrics map[string]limitDTO `yaml:"metrics"` // type metric bounds
}

func (d fileDTO) toPolicy() (domain.Policy, error) {
	if d.Version != 1 {
		return domain.Policy{}, fmt.Errorf("unsupported version %d (want 1)", d.Version)
	}

	policy := domain.Policy{
		Metrics:        make(map[string]domain.Limit, len(d.Metrics)),
		PackageMetrics: make(map[string]domain.Limit, len(d.Package.Metrics)),
		TypeMetrics:    make(map[string]domain.Limit, len(d.Type.Metrics)),
	}

	policy.Package.Types = d.Package.Types.toLimit()
	policy.Package.ExportedFuncs = d.Package.ExportedFuncs.toLimit()
	policy.Package.UnexportedFuncs = d.Package.UnexportedFuncs.toLimit()
	policy.Package.Afferent = d.Package.Afferent.toLimit()
	policy.Package.Efferent = d.Package.Efferent.toLimit()
	policy.Type.Fields = d.Type.Fields.toLimit()
	policy.Type.Methods = d.Type.Methods.toLimit()

	copyMetricLimits(policy.Metrics, d.Metrics)
	copyMetricLimits(policy.PackageMetrics, d.Package.Metrics)
	copyMetricLimits(policy.TypeMetrics, d.Type.Metrics)

	return policy, nil
}

func copyMetricLimits(dst map[string]domain.Limit, src map[string]limitDTO) {
	for name, limit := range src {
		dst[name] = limit.toLimit()
	}
}

// limitDTO decodes either a bare scalar (a max bound) or a {min, max} mapping.
type limitDTO struct {
	max    float64
	hasMax bool
	min    float64
	hasMin bool
}

func (l *limitDTO) toLimit() domain.Limit {
	if l == nil {
		return domain.Limit{}
	}

	return domain.Limit{Max: l.max, HasMax: l.hasMax, Min: l.min, HasMin: l.hasMin}
}

// UnmarshalYAML accepts `key: 5` (a max) or `key: { min: .., max: .. }`.
// KnownFields does not propagate into a custom unmarshaler, so unknown mapping
// keys are rejected here by hand.
func (l *limitDTO) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var value float64
		if err := node.Decode(&value); err != nil {
			return fmt.Errorf("line %d: limit must be a number or a {min, max} mapping", node.Line)
		}

		l.max, l.hasMax = value, true

		return nil
	case yaml.MappingNode:
		return decodeMappingLimit(l, node)
	default:
		return fmt.Errorf("line %d: limit must be a number or a {min, max} mapping", node.Line)
	}
}

// decodeMappingLimit reads a {min, max} mapping node, rejecting unknown keys
// and an empty mapping by hand.
func decodeMappingLimit(l *limitDTO, node *yaml.Node) error {
	var aux struct {
		Max *float64 `yaml:"max"`
		Min *float64 `yaml:"min"`
	}
	if err := node.Decode(&aux); err != nil {
		return fmt.Errorf("line %d: %w", node.Line, err)
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if key != "max" && key != "min" {
			return fmt.Errorf(
				"line %d: unknown limit field %q (want max or min)",
				node.Content[i].Line,
				key,
			)
		}
	}

	if aux.Max != nil {
		l.max, l.hasMax = *aux.Max, true
	}

	if aux.Min != nil {
		l.min, l.hasMin = *aux.Min, true
	}

	if !l.hasMax && !l.hasMin {
		return fmt.Errorf("line %d: limit must set max and/or min", node.Line)
	}

	return nil
}
