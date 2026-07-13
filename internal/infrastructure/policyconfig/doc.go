// Package policyconfig loads a modularity policy from a .modularity.yml file.
//
// It is the infrastructure adapter that maps the YAML document onto the pure
// policy domain: reading files, decoding YAML, and rejecting unknown keys and
// unsupported schema versions. The parsed policy is validated by the domain
// before it is returned.
package policyconfig
