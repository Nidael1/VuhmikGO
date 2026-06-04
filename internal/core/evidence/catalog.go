package evidence

// ErrorCatalog es el catálogo canónico de códigos de error del Core.
//
// Es la única fuente de verdad para códigos de error de evidencia.
// Los códigos son estables, inmutables y auditables.
// Ningún flujo del Core emite errores fuera de este catálogo.
//
// Semántica de cada código:
//
//	ER-CORE-001 — Inmutabilidad violada.
//	  Se emite cuando se intenta modificar o eliminar un registro
//	  en estado issued o locked. No existe excepción.
//
//	ER-CORE-002 — Transición de estado inválida.
//	  Se emite cuando la transición solicitada no está declarada
//	  en la matriz de lifecycle. No existe transición implícita.
//
//	ER-CORE-003 — reason_code ausente.
//	  Se emite cuando una operación que requiere reason_code
//	  (como void) se ejecuta sin proveerlo.
//
//	ER-CORE-004 — Reemplazo inválido.
//	  Se emite cuando el registro de reemplazo tiene un ID vacío
//	  o igual al del registro original.
var ErrorCatalog = map[string]*ErrCore{
	"ER-CORE-001": ErrImmutable,
	"ER-CORE-002": ErrInvalidTransition,
	"ER-CORE-003": ErrMissingReasonCode,
	"ER-CORE-004": ErrInvalidReplacement,
}
