package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// ConsultationService orquesta el caso de uso de consulta médica (ADR-0024).
type ConsultationService struct {
	repo  ports.EvidenceRepository
	proj  ports.ConsultationProjectionRepository
	rubro string
}

func NewConsultationService(
	repo ports.EvidenceRepository,
	proj ports.ConsultationProjectionRepository,
) *ConsultationService {
	return &ConsultationService{repo: repo, proj: proj, rubro: "medico"}
}

// Create crea y emite una consulta en un solo paso.
func (s *ConsultationService) Create(tenantID, actorID, patientID string, content shaders.ConsultationContent) (evidence.Evidence, error) {
	blob, err := shaders.BuildConsultationBlob(content)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al construir blob: %w", err)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("con-%s-%s", tenantID[:4], now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: patientID,
		Content:    blob,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear consulta: %w", err)
	}

	// Emitir inmediatamente
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir consulta: %w", err)
	}

	// Proyección (ADR-0022)
	_ = s.proj.Upsert(ports.ConsultationProjection{
		EvidenceID: e.ID,
		TenantID:   tenantID,
		PatientID:  patientID,
		TA:         content.TA,
		FC:         content.FC,
		FR:         content.FR,
		Temp:       content.Temp,
		Peso:       content.Peso,
		Talla:      content.Talla,
		SAO2:       content.SAO2,
		State:      string(e.State),
		CreatedAt:  now,
		IssuedAt:   e.IssuedAt,
	})

	return e, nil
}

func (s *ConsultationService) ListByPatient(tenantID, patientID string) ([]ports.ConsultationProjection, error) {
	return s.proj.ListByPatient(tenantID, patientID)
}

func (s *ConsultationService) ListAll(tenantID string) ([]ports.ConsultationProjection, error) {
	return s.proj.ListAll(tenantID)
}

func (s *ConsultationService) FindByID(tenantID, id string) (ports.ConsultationProjection, error) {
	return s.proj.FindByID(tenantID, id)
}
