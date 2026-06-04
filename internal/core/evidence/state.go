package evidence

// State is the strongly-typed value of the State field on Evidence.
//
// Valid state constants are defined below. Any value outside these
// constants is invalid and will be rejected by lifecycle enforcement
// introduced in subsequent issues.
type State string

// Valid states of an Evidence record.
const (
	StateDraft  State = "draft"
	StateIssued State = "issued"
	StateLocked State = "locked"
	StateVoided State = "voided"
)
