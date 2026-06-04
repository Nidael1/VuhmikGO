package evidence

// GuardMutation verifica que un registro Evidence puede ser mutado.
// Bloquea UPDATE y DELETE sobre registros en estado issued o locked.
//
// Retorna ErrImmutable (ER-CORE-001) si la mutación está prohibida.
// Retorna nil si la mutación está permitida.
//
// Este guard es la única vía de validación de inmutabilidad del Core.
// No existe bypass por rutas alternativas de escritura.
func GuardMutation(e Evidence) error {
	if e.State == StateIssued || e.State == StateLocked {
		return ErrImmutable
	}
	return nil
}
