package evidence

// ReasonCode es el tipo de código de razón para acciones sensibles del Core.
//
// Solo se permiten valores de este catálogo. No se aceptan strings libres.
type ReasonCode string

// Catálogo canónico de reason_code para operaciones de void.
// Obligatorio en toda operación de anulación.
const (
	// RCVoidErrorDetected — se detectó un error en el contenido del registro.
	RCVoidErrorDetected ReasonCode = "RC-VOID-001"

	// RCVoidUpdateRequired — la información requiere actualización.
	RCVoidUpdateRequired ReasonCode = "RC-VOID-002"

	// RCVoidRequested — anulación solicitada formalmente.
	RCVoidRequested ReasonCode = "RC-VOID-003"

	// RCVoidAdministrative — decisión administrativa documentada.
	RCVoidAdministrative ReasonCode = "RC-VOID-004"
)

// Catálogo canónico de reason_code para operaciones de replace.
// Obligatorio en toda operación de reemplazo.
const (
	// RCReplaceCorrection — reemplazo por corrección de error.
	RCReplaceCorrection ReasonCode = "RC-REPLACE-001"

	// RCReplaceUpdate — reemplazo por actualización de información.
	RCReplaceUpdate ReasonCode = "RC-REPLACE-002"
)

// ReasonCodeCatalog es el catálogo indexado de todos los reason_code válidos.
// Es la única fuente de verdad para validación de reason_code.
// La integración en operaciones se define en el issue de enforcement.
var ReasonCodeCatalog = map[ReasonCode]string{
	RCVoidErrorDetected:  "Se detectó un error en el contenido del registro",
	RCVoidUpdateRequired: "La información del registro requiere actualización",
	RCVoidRequested:      "Anulación solicitada formalmente",
	RCVoidAdministrative: "Decisión administrativa documentada",
	RCReplaceCorrection:  "Reemplazo por corrección de error",
	RCReplaceUpdate:      "Reemplazo por actualización de información",
}
