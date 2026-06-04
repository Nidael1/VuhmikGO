package evidence

import "time"

// Replace anula la evidencia original y la reemplaza por una nueva,
// preservando el historial completo.
//
// Operaciones de dominio:
//  1. Anula el original con reasonCode validado contra el catálogo.
//  2. Emite el registro de reemplazo como issued.
//  3. Enlaza original → reemplazo mediante replaced_by_id.
//
// Reglas:
//   - reasonCode debe estar en el ReasonCodeCatalog (validado por Void).
//   - El reemplazo debe tener un ID no vacío y distinto al original.
//   - El original permanece preservado e inalterable.
//
// Retorna (original anulado, reemplazo emitido, error).
func Replace(
	original Evidence,
	replacement Evidence,
	reasonCode ReasonCode,
	at time.Time,
) (Evidence, Evidence, error) {
	if replacement.ID == "" || replacement.ID == original.ID {
		return Evidence{}, Evidence{}, ErrInvalidReplacement
	}
	voided, err := Void(original, reasonCode, at)
	if err != nil {
		return Evidence{}, Evidence{}, err
	}
	voided.ReplacedByID = &replacement.ID
	replacement.State = StateIssued
	replacement.IssuedAt = &at
	return voided, replacement, nil
}
