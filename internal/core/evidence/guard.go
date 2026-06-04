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

// GuardContentEdit bloquea modificaciones de contenido en registros
// en estado issued o locked.
//
// Este guard opera a nivel de dominio, antes de cualquier operación
// de persistencia. Se distingue de GuardMutation en que aplica
// específicamente a intentos de edición de campos de contenido,
// no a operaciones de base de datos.
//
// Retorna ErrImmutable (ER-CORE-001) si el registro no puede editarse.
// Retorna nil si la edición está permitida.
//
// No existe "edición parcial": cualquier intento de cambiar contenido
// en issued o locked es rechazado sin excepción.
func GuardContentEdit(e Evidence) error {
	if e.State == StateIssued || e.State == StateLocked {
		return ErrImmutable
	}
	return nil
}
