package model

import (
	"fmt"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/bitset"
)

// Position locates a declaration in source, with File relative to the
// analysis directory when possible so output is machine-independent.
type Position struct {
	// File is the source file path, relative when possible.
	File string
	// Line is the 1-based source line.
	Line int
	// Column is the 1-based source column.
	Column int
}

// FieldFacts describes one struct field slot.
type FieldFacts struct {
	// Name is the field name (the type name for embedded fields).
	Name string
	// Exported reports whether the field name is exported.
	Exported bool
	// Embedded marks an embedded (anonymous) field.
	Embedded bool
}

// MethodFacts describes one explicitly declared method.
type MethodFacts struct {
	// Name is the method name.
	Name string
	// Exported reports whether the method name is exported.
	Exported bool
	// Pos locates the method declaration in source.
	Pos Position
	// FieldsUsed marks the receiver type's fields this method's body
	// accesses directly (resolved through type-checked selections, never
	// selector names). Indices refer to TypeFacts.Fields.
	FieldsUsed bitset.FieldSet
	// ParamTypeKeys are the canonical keys of the method's distinct
	// parameter types (receiver and returns excluded, duplicates collapsed),
	// sorted. Generic type parameters keep their identity.
	ParamTypeKeys []string
	// Branches carries the branch counts feeding cyclomatic complexity.
	Branches BranchStats
	// CalledSiblings are indices (into TypeFacts.Methods) of methods of the
	// same type this method calls; input for transitive field usage.
	CalledSiblings []int
}

// String summarizes the method facts for debugging.
func (m *MethodFacts) String() string {
	return fmt.Sprintf("method %q (exported %v) at %v: uses %d fields, %d param types, branches %+v, calls %v",
		m.Name, m.Exported, m.Pos, bitset.Count(m.FieldsUsed),
		len(m.ParamTypeKeys), m.Branches, m.CalledSiblings)
}

// BranchStats counts the syntax constructs that increment cyclomatic
// complexity. The formula itself lives in the complexity feature.
type BranchStats struct {
	Ifs         int // if statements
	Fors        int // for loops (all, including conditionless "for {}")
	Ranges      int // range loops
	Cases       int // non-default switch and type-switch cases
	SelectComms int // select communication clauses (default excluded)
	LogicalOps  int // && and ||
}
