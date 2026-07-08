package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// NoteService orquesta el caso de uso de notas clínicas genéricas.
//
// A diferencia de allergy/prescription/diagnosis/immunization/lab_result,
// "note" no tiene Shader tipado propio ni Content struct — el contenido
// es texto libre (campo Text) mas signos vitales opcionales. Usa
// MedicalBasicShader (perfil med_basic) como Shader base, envuelto en
// CapabilityGuard igual que los demas modulos (ADR-0017).
//
// Auto-emit (ADR-0006): a diferencia del resto, el borrador se emite
// automaticamente en el mismo Create — el medico no ve el estado draft
// como paso manual. El frontend nunca debe llamar a un endpoint de
// emit despues de crear una nota.
type NoteService struct {
	repo  ports.EvidenceRepository
	proj  ports.NoteProjectionRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NoteContent es el contenido de una nota clínica: texto libre mas
// signos vitales opcionales. No vive en shaders/ porque "note" no tiene
// Shader tipado — el contenido se pasa directo desde el handler.
type NoteContent struct {
	SubjectRef string
	Text       string
	TA         string
	FC         string
	FR         string
	Temp       string
	Peso       string
	Talla      string
	SAO2       string
}

// NewNoteService retorna un NoteService con sus dependencias.
func NewNoteService(repo ports.EvidenceRepository, proj ports.NoteProjectionRepository, caps ports.CapabilityRepository) *NoteService {
	return &NoteService{repo: repo, proj: proj, caps: caps, rubro: "medico"}
}

// Edit actualiza una nota. Si esta en draft, re-emite el mismo registro.
// Si ya esta emitida, ejecuta void + replace silencioso (ADR-0006) — el
// medico ve esto como "editar", pero el Core nunca muta un registro
// emitido; crea un reemplazo y anula el original.
//
// rawContent es el blob JSON completo del nuevo contenido; noteText es
// el texto ya extraido de ese blob, para la proyeccion (ADR-0022).
func (s *NoteService) Edit(tenantID, actorID, evidenceID, rawContent, noteText string) (evidence.Evidence, error) {
	inner := shaders.NewMedicalBasicShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "note", s.rubro)

	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		Operation: shaders.OperationReplace,
		ActorID:   actorID,
	}
	decision := guard.Evaluate(ctx)
	if decision.Result != shaders.DecisionAllow {
		return evidence.Evidence{}, fmt.Errorf("[%s] %s", decision.ErrorCode, decision.Reason)
	}

	orig, err := s.repo.FindByID(tenantID, evidenceID)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("nota no encontrada: %w", err)
	}

	now := time.Now().UTC()

	if orig.State == evidence.StateDraft {
		// Solo re-emite el draft existente
		if err := evidence.GuardTransition(orig.State, evidence.StateIssued); err != nil {
			return evidence.Evidence{}, err
		}
		orig.State = evidence.StateIssued
		orig.IssuedAt = &now
		if err := s.repo.Update(tenantID, orig); err != nil {
			return evidence.Evidence{}, fmt.Errorf("error al emitir nota: %w", err)
		}
		return orig, nil
	}

	// Void + replace silencioso (ADR-0006)
	newID := evidenceID + "-v" + now.Format("20060102150405")
	repl := evidence.Evidence{
		ID:         newID,
		TenantID:   tenantID,
		SubjectRef: orig.SubjectRef,
		Content:    rawContent,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}

	voided, issued, err := evidence.Replace(orig, repl, evidence.RCReplaceUpdate, now)
	if err != nil {
		return evidence.Evidence{}, err
	}

	// Orden FK crítico: primero Create(nuevo), luego UpdateForVoid(original)
	// para no violar evidence_replaced_by_fk (ADR-0006)
	if err := s.repo.Create(issued); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear version: %w", err)
	}
	if err := s.repo.UpdateForVoid(tenantID, voided); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al anular version anterior: %w", err)
	}

	// Actualizar proyección (ADR-0022)
	_ = s.proj.UpdateState(tenantID, orig.ID, string(evidence.StateVoided))
	_ = s.proj.Upsert(ports.NoteProjection{
		EvidenceID: issued.ID,
		TenantID:   tenantID,
		PatientID:  issued.SubjectRef,
		Text:       noteText,
		State:      string(issued.State),
		CreatedAt:  issued.CreatedAt,
		IssuedAt:   issued.IssuedAt,
	})

	return issued, nil
}
//
// rawContent es el blob JSON completo tal como lo envía el frontend
// (ej. {"type":"note","text":"..."}) — se guarda tal cual en el Core,
// que lo trata como blob opaco (ADR-0016). content.Text es el texto ya
// extraído de ese blob, usado solo para la proyección de lectura
// (ADR-0022), que no debe re-parsear JSON en cada lectura.
func (s *NoteService) Create(tenantID, actorID, rawContent string, content NoteContent) (evidence.Evidence, error) {
	inner := shaders.NewMedicalBasicShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "note", s.rubro)

	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		Operation: shaders.OperationCreate,
		ActorID:   actorID,
	}
	decision := guard.Evaluate(ctx)
	if decision.Result != shaders.DecisionAllow {
		return evidence.Evidence{}, fmt.Errorf("[%s] %s", decision.ErrorCode, decision.Reason)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("ev-%s", now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: content.SubjectRef,
		Content:    rawContent,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear nota: %w", err)
	}

	// Emite automáticamente (ADR-0006 — el médico no ve estados internos)
	if err := evidence.GuardTransition(e.State, evidence.StateIssued); err != nil {
		return evidence.Evidence{}, err
	}
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir nota: %w", err)
	}

	// Escribir proyección de lectura (ADR-0022)
	proj := ports.NoteProjection{
		EvidenceID: e.ID,
		TenantID:   tenantID,
		PatientID:  content.SubjectRef,
		Text:       content.Text,
		State:      string(e.State),
		CreatedAt:  now,
		IssuedAt:   e.IssuedAt,
		TA:         content.TA,
		FC:         content.FC,
		FR:         content.FR,
		Temp:       content.Temp,
		Peso:       content.Peso,
		Talla:      content.Talla,
		SAO2:       content.SAO2,
	}
	if err := s.proj.Upsert(proj); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al proyectar nota: %w", err)
	}
	return e, nil
}
