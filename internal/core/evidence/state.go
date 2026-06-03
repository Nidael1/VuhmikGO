// Package evidence defines the minimal Core entity representing a base
// evidence record in VUHMÍK.
//
// SCOPE NOTE (Issue #1):
// This package contains only structure. It has no behavior, no methods,
// no lifecycle enforcement, no persistence and no business logic.
// Subsequent issues introduce states, lifecycle, transitions, immutability
// guards and persistence per the execution plan.
package evidence

// State is the strongly-typed value of the State field on Evidence.
//
// SCOPE NOTE (Issue #1):
// This type exists solely to type the State field on the Evidence struct,
// as authorized by ADR-0003 §5. Valid state values, lifecycle, transitions
// and behavior are defined in subsequent issues per the execution plan.
//
// This file intentionally contains:
//   - no state constants
//   - no transition matrix
//   - no methods
//   - no behavior
type State string
