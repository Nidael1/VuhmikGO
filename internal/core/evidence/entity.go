package evidence

import "time"

// Evidence is the minimal Core entity representing a base evidence record.
//
// SCOPE NOTE (Issue #1):
// This struct defines only structure (fields, types, nullability).
// It has no methods, no behavior, no lifecycle enforcement, no validators
// and no business logic. Lifecycle, transitions, immutability guards,
// persistence and other behaviors are introduced in subsequent issues
// per the execution plan.
//
// Field semantics:
//
//   - ID            : unique identifier of the evidence record.
//   - TenantID      : tenant that owns the record (multi-tenant isolation).
//   - State         : current state of the record, strongly typed via State.
//     Valid values are defined in a subsequent issue.
//   - CreatedAt     : creation timestamp (non-nullable).
//   - IssuedAt      : timestamp when the record was issued (nullable).
//   - VoidedAt      : timestamp when the record was voided (nullable).
//   - ReplacedByID  : identifier of the record that replaces this one
//     (nullable).
//
// Nullability is expressed via pointer types for nullable fields, matching
// idiomatic Go for optional values without committing to any specific
// persistence representation.
type Evidence struct {
	ID           string
	TenantID     string
	State        State
	CreatedAt    time.Time
	IssuedAt     *time.Time
	VoidedAt     *time.Time
	ReplacedByID *string
}
