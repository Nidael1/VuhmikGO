package evidence

// Relations documents the structural domain relationships of Evidence.
//
// The Evidence entity has exactly two structural relationships in the Core.
// No other relationships exist at this layer.
//
// # Tenant association (mandatory)
//
// Every Evidence record belongs to exactly one tenant via TenantID.
// TenantID must never be empty. Cross-tenant access is forbidden.
//
// # Self-referential replacement (optional)
//
// ReplacedByID points to the record that supersedes this one.
//   - When nil: the record has not been replaced.
//   - When non-nil: the referenced record belongs to the same tenant.
//   - Replacement is one-directional (voided → replacement).
//   - The replacement record does not reference back.
//
// Lifecycle enforcement, persistence and integrity guards are defined
// in subsequent issues per the execution plan.
