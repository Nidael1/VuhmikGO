package api

import (
	"net/http"
	"strings"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// RegisterAPIRoutes registra las rutas de la API JSON /api/v1.
func RegisterAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/register", HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", HandleLogin)
	mux.HandleFunc("/api/v1/auth/me", JWTMiddleware(HandleMe))
	mux.HandleFunc("/api/v1/auth/refresh", HandleRefresh)
	mux.HandleFunc("/api/v1/auth/logout", HandleLogout)
	mux.HandleFunc("/api/v1/evidence", JWTMiddleware(HandleEvidenceList))
	mux.HandleFunc("/api/v1/evidence/draft", JWTMiddleware(HandleEvidenceDraft))

	// Dispatcher dinamico para rutas con ID variable:
	// /api/v1/evidence/:id
	// /api/v1/evidence/:id/emit
	// /api/v1/evidence/:id/void
	// /api/v1/evidence/:id/replace
	// /api/v1/evidence/:id/export
	// /api/v1/evidence/:id/edit
	mux.HandleFunc("/api/v1/evidence/", JWTMiddleware(evidenceDispatcher))
	mux.HandleFunc("/api/v1/patients/import", JWTMiddleware(HandlePatientImport))
	mux.HandleFunc("/api/v1/patients", JWTMiddleware(patientsBaseDispatcher))
	mux.HandleFunc("/api/v1/patients/", JWTMiddleware(patientDispatcher))
	mux.HandleFunc("/api/v1/allergies/", JWTMiddleware(allergyDispatcher))
	mux.HandleFunc("/api/v1/profile", JWTMiddleware(profileDispatcher))
	mux.HandleFunc("/api/v1/admin/vendors", JWTMiddleware(HandleVendorList))
	mux.HandleFunc("/api/v1/admin/tenants", JWTMiddleware(AdminMiddleware(HandleAdminTenants)))
	mux.HandleFunc("/api/v1/admin/capabilities", JWTMiddleware(AdminMiddleware(HandleAdminCapabilityToggle)))
	mux.HandleFunc("/api/v1/admin/users", JWTMiddleware(AdminMiddleware(HandleAdminCreateUser)))
	mux.HandleFunc("/api/v1/admin/suspend", JWTMiddleware(AdminMiddleware(HandleAdminSuspend)))
	mux.HandleFunc("/api/v1/consultations", JWTMiddleware(consultationBaseDispatcher))
	mux.HandleFunc("/api/v1/consultations/", JWTMiddleware(consultationDispatcher))
	mux.HandleFunc("/api/v1/prescriptions", JWTMiddleware(HandlePrescriptionListAll))
	mux.HandleFunc("/api/v1/prescriptions/", prescriptionAuthDispatcher)
}

// evidenceDispatcher enruta requests con ID dinamico en el path.
// ServeMux de Go no soporta path params — este dispatcher los resuelve.
func evidenceDispatcher(w http.ResponseWriter, r *http.Request) {
	// Elimina el prefijo base
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/evidence/")
	// path puede ser: "abc123", "abc123/emit", "abc123/export", etc.
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 1 {
		// GET /api/v1/evidence/:id
		HandleEvidenceDetail(w, r)
		return
	}

	action := parts[1]
	switch action {
	case "emit":
		HandleEvidenceEmit(w, r)
	case "void":
		HandleEvidenceVoid(w, r)
	case "replace":
		HandleEvidenceReplace(w, r)
	case "export":
		HandleEvidenceExport(w, r)
	case "edit":
		HandleEvidenceEdit(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
	}
}

// JWTMiddleware protege un handler exigiendo un JWT valido.
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "token requerido")
			return
		}
		claims, err := auth.ValidateToken(header[7:])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "token invalido o expirado")
			return
		}
		r = r.WithContext(ContextWithClaims(r.Context(), claims))
		next(w, r)
	}
}

func init() {
	// rutas de pacientes registradas en RegisterAPIRoutes
}

// patientDispatcher enruta requests de pacientes con ID dinamico.
func patientDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			HandlePatientDetail(w, r)
		case http.MethodPut:
			HandlePatientUpdate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
		return
	}
	// Subrutas: /patients/:id/allergies
	switch parts[1] {
	case "export":
		HandlePatientExport(w, r)
	case "export/zip":
		HandlePatientExportZIP(w, r)
	case "consultations":
		switch r.Method {
		case http.MethodGet:
			HandleConsultationListByPatient(w, r)
		case http.MethodPost:
			HandleConsultationCreate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
	case "prescriptions":
		switch r.Method {
		case http.MethodGet:
			HandlePrescriptionListByPatient(w, r)
		case http.MethodPost:
			HandlePrescriptionCreate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
	case "lab-results":
		switch r.Method {
		case http.MethodGet:
			HandleLabResultListByPatient(w, r)
		case http.MethodPost:
			HandleLabResultCreate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
	case "immunizations":
		switch r.Method {
		case http.MethodGet:
			HandleImmunizationListByPatient(w, r)
		case http.MethodPost:
			HandleImmunizationCreate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
	case "diagnoses":
		switch r.Method {
		case http.MethodGet:
			HandleDiagnosisListByPatient(w, r)
		case http.MethodPost:
			HandleDiagnosisCreate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
	case "allergies":
		switch r.Method {
		case http.MethodGet:
			HandleAllergyList(w, r)
		case http.MethodPost:
			HandleAllergyCreate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		}
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
	}
}

// patientsBaseDispatcher maneja GET (lista) y POST (crear) en /api/v1/patients.
func patientsBaseDispatcher(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandlePatientList(w, r)
	case http.MethodPost:
		HandlePatientCreate(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
	}
}

// profileDispatcher enruta GET y PUT en /api/v1/profile.
func profileDispatcher(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandleGetProfile(w, r)
	case http.MethodPut:
		HandleUpdateProfile(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
	}
}

// consultationBaseDispatcher maneja GET /api/v1/consultations
func consultationBaseDispatcher(w http.ResponseWriter, r *http.Request) {
	HandleConsultationListAll(w, r)
}

// consultationDispatcher enruta requests de consultas con ID dinamico.
func consultationDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/consultations/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		HandleConsultationDetail(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
}

// prescriptionAuthDispatcher decide el modo de autenticacion antes de
// delegar a prescriptionDispatcher. La sub-ruta /print no exige el header
// JWT estricto porque se abre en pestana nueva (window.open) y valida el
// token por su cuenta (header o query) dentro del propio handler.
func prescriptionAuthDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/prescriptions/")
	if strings.HasSuffix(path, "/print") {
		prescriptionDispatcher(w, r)
		return
	}
	JWTMiddleware(prescriptionDispatcher)(w, r)
}

// prescriptionDispatcher enruta requests de recetas con ID dinamico.
// Soporta: /api/v1/prescriptions/:id/emit, /api/v1/prescriptions/:id/void
// y /api/v1/prescriptions/:id/print
func prescriptionDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/prescriptions/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		HandlePrescriptionDetail(w, r)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "emit":
			HandlePrescriptionEmit(w, r)
		case "void":
			HandlePrescriptionVoid(w, r)
		case "print":
			HandlePrescriptionPrint(w, r)
		default:
			writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
		}
		return
	}
	writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
}

// AdminMiddleware protege rutas de admin exigiendo is_admin = true en el JWT.
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
		if !ok || claims == nil || !claims.IsAdmin {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "acceso restringido a administradores")
			return
		}
		next(w, r)
	}
}

// allergyDispatcher enruta requests de alergias con ID dinamico.
// Soporta: /api/v1/allergies/:id/void
func allergyDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/allergies/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 && parts[1] == "void" {
		HandleAllergyVoid(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
}
