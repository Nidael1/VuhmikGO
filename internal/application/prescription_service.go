package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// PrescriptionService orquesta el caso de uso de receta electronica.
//
// Flujo:
//  1. Handler recibe y delega.
//  2. PrescriptionService prepara la operacion.
//  3. Core consulta politica al Shader (CapabilityGuard + PrescriptionShader).
//  4. Shader valida campos minimos NOM-024 antes de permitir emision.
//  5. Core ejecuta + proyeccion se actualiza (ADR-0022).
//  6. Asteroide muestra resultado.
//
// La receta usa emit EXPLICITO (draft → issued) — no auto-emit (ADR-0011).
// Una vez emitida es inmutable. Correcciones via void + replace.
type PrescriptionService struct {
	repo  ports.EvidenceRepository
	proj  ports.PrescriptionProjectionRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NewPrescriptionService retorna un PrescriptionService con sus dependencias.
func NewPrescriptionService(
	repo ports.EvidenceRepository,
	proj ports.PrescriptionProjectionRepository,
	caps ports.CapabilityRepository,
) *PrescriptionService {
	return &PrescriptionService{repo: repo, proj: proj, caps: caps, rubro: "medico"}
}

// CreateDraft crea un borrador de receta. No tiene validez legal hasta Emit.
func (s *PrescriptionService) CreateDraft(tenantID, actorID, patientID string, content shaders.PrescriptionContent) (evidence.Evidence, error) {
	inner := shaders.NewPrescriptionShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "prescription", s.rubro)

	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		Operation: shaders.OperationCreate,
		ActorID:   actorID,
	}
	decision := guard.Evaluate(ctx)
	if decision.Result != shaders.DecisionAllow {
		return evidence.Evidence{}, fmt.Errorf("[%s] %s", decision.ErrorCode, decision.Reason)
	}

	blob, err := shaders.BuildPrescriptionBlob(content)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al construir blob: %w", err)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("rx-%s-%s", tenantID[:4], now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: patientID,
		Content:    blob,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear borrador de receta: %w", err)
	}

	// Proyección en draft (ADR-0022)
	_ = s.proj.Upsert(ports.PrescriptionProjection{
		EvidenceID:          e.ID,
		TenantID:            tenantID,
		PatientID:           patientID,
		MedicamentoGenerico: content.MedicamentoGenerico,
		Dosis:               content.Dosis,
		Diagnostico:         content.Diagnostico,
		Indicaciones:        content.Indicaciones,
		Seguimiento:         content.Seguimiento,
		State:               string(e.State),
		CreatedAt:           now,
	})

	return e, nil
}

// Emit emite una receta draft — adquiere validez legal (ADR-0011).
// Valida campos minimos NOM-024 antes de emitir.
func (s *PrescriptionService) Emit(tenantID, actorID, prescriptionID string, profile ports.Profile) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, prescriptionID)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("receta no encontrada: %w", err)
	}

	// Parsear blob para validar + enriquecer con datos del perfil
	var content shaders.PrescriptionContent
	if err := shaders.ParsePrescriptionBlob(e.Content, &content); err != nil {
		return evidence.Evidence{}, fmt.Errorf("blob invalido: %w", err)
	}

	// Enriquecer con cedula y especialidad del perfil del medico
	content.CedulaProfesional = profile.CedulaProfesional
	content.Especialidad = profile.Especialidad

	// Validar campos minimos NOM-024
	if err := shaders.ValidatePrescriptionContent(content); err != nil {
		return evidence.Evidence{}, fmt.Errorf("campos minimos NOM-024 incompletos: %w", err)
	}

	// Rebuild blob con cedula y especialidad
	blob, err := shaders.BuildPrescriptionBlob(content)
	if err != nil {
		return evidence.Evidence{}, err
	}
	e.Content = blob

	// Transición draft → issued
	if err := evidence.GuardTransition(e.State, evidence.StateIssued); err != nil {
		return evidence.Evidence{}, err
	}
	now := time.Now().UTC()
	e.State = evidence.StateIssued
	e.IssuedAt = &now

	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir receta: %w", err)
	}

	// Actualizar proyección a issued (ADR-0022)
	_ = s.proj.Upsert(ports.PrescriptionProjection{
		EvidenceID:          e.ID,
		TenantID:            tenantID,
		PatientID:           e.SubjectRef,
		MedicamentoGenerico: content.MedicamentoGenerico,
		Dosis:               content.Dosis,
		Diagnostico:         content.Diagnostico,
		Indicaciones:        content.Indicaciones,
		Seguimiento:         content.Seguimiento,
		State:               string(e.State),
		CreatedAt:           e.CreatedAt,
		IssuedAt:            e.IssuedAt,
	})

	return e, nil
}

// Void anula una receta emitida. Corrección via void + replace (ADR-0006).
func (s *PrescriptionService) Void(tenantID, actorID, prescriptionID string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, prescriptionID)
	if err != nil {
		return evidence.Evidence{}, err
	}
	voided, err := evidence.Void(e, evidence.RCVoidErrorDetected, time.Now().UTC())
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := s.repo.UpdateForVoid(tenantID, voided); err != nil {
		return evidence.Evidence{}, err
	}
	_ = s.proj.UpdateState(tenantID, prescriptionID, string(evidence.StateVoided))
	return voided, nil
}

// ListByPatient retorna recetas emitidas de un paciente.
// Lee de prescription_projections (ADR-0022).
func (s *PrescriptionService) ListByPatient(tenantID, patientID string) ([]ports.PrescriptionProjection, error) {
	return s.proj.ListByPatient(tenantID, patientID)
}

// ListAll retorna todas las recetas emitidas del tenant.
// Para la vista global de recetas en el sidebar.
func (s *PrescriptionService) ListAll(tenantID string) ([]ports.PrescriptionProjection, error) {
	return s.proj.ListAll(tenantID)
}
