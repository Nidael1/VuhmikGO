package evidence

import "time"

// Void anula un registro de evidencia marcándolo como voided.
//
// Reglas:
//   - reason_code es obligatorio; sin él retorna ErrMissingReasonCode (ER-CORE-003).
//   - La transición debe estar permitida por la matriz de lifecycle.
//   - El contenido original no se modifica; solo cambia el estado y voided_at.
//   - Un registro voided no es editable ni reanulable.
//
// Retorna el registro actualizado con estado voided y voided_at asignado.
// El registro original (pasado por valor) permanece sin cambios.
func Void(e Evidence, reasonCode string, voidedAt time.Time) (Evidence, error) {
	if reasonCode == "" {
		return Evidence{}, ErrMissingReasonCode
	}
	if err := GuardTransition(e.State, StateVoided); err != nil {
		return Evidence{}, err
	}
	e.State = StateVoided
	e.VoidedAt = &voidedAt
	return e, nil
}
