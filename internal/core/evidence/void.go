package evidence

import "time"

// Void anula un registro de evidencia marcándolo como voided.
//
// Reglas:
//   - reasonCode debe estar en el ReasonCodeCatalog; si no, retorna ErrMissingReasonCode.
//   - La transición debe estar permitida por la matriz de lifecycle.
//   - El contenido original no se modifica; solo cambia el estado y voided_at.
//   - Un registro voided no es editable ni reanulable.
//
// Retorna el registro actualizado con estado voided y voided_at asignado.
func Void(e Evidence, reasonCode ReasonCode, voidedAt time.Time) (Evidence, error) {
	if _, ok := ReasonCodeCatalog[reasonCode]; !ok {
		return Evidence{}, ErrMissingReasonCode
	}
	if err := GuardTransition(e.State, StateVoided); err != nil {
		return Evidence{}, err
	}
	e.State = StateVoided
	e.VoidedAt = &voidedAt
	return e, nil
}
