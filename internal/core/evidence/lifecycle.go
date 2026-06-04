package evidence

// validTransitions es la matriz explícita y cerrada de transiciones
// permitidas entre estados Core.
//
// Esta tabla es la única fuente de verdad para validación de lifecycle.
// Toda transición no declarada aquí es inválida por definición.
//
// Transiciones válidas:
//
//	draft   → issued
//	issued  → locked
//	issued  → voided
//	locked  → voided
var validTransitions = map[State][]State{
	StateDraft:  {StateIssued},
	StateIssued: {StateLocked, StateVoided},
	StateLocked: {StateVoided},
	StateVoided: {},
}

// GuardTransition valida que la transición de current a next
// está declarada en la matriz de lifecycle.
//
// Retorna ErrInvalidTransition (ER-CORE-002) si la transición no existe.
// Retorna nil si la transición es válida.
//
// No existe transición implícita ni permisiva por defecto.
func GuardTransition(current, next State) error {
	allowed, ok := validTransitions[current]
	if !ok {
		return ErrInvalidTransition
	}
	for _, s := range allowed {
		if s == next {
			return nil
		}
	}
	return ErrInvalidTransition
}
