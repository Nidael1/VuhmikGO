package evidence

import (
	"testing"
	"time"
)

// --- Inmutabilidad (ER-CORE-001) ---

func TestGuardMutation_BloqueaIssued(t *testing.T) {
	e := Evidence{State: StateIssued}
	err := GuardMutation(e)
	if err == nil {
		t.Fatal("esperaba error en estado issued, obtuvo nil")
	}
	if ExtractErrorCode(err) != "ER-CORE-001" {
		t.Fatalf("código incorrecto: %s", ExtractErrorCode(err))
	}
}

func TestGuardMutation_BloqueaLocked(t *testing.T) {
	e := Evidence{State: StateLocked}
	err := GuardMutation(e)
	if err == nil {
		t.Fatal("esperaba error en estado locked, obtuvo nil")
	}
	if ExtractErrorCode(err) != "ER-CORE-001" {
		t.Fatalf("código incorrecto: %s", ExtractErrorCode(err))
	}
}

func TestGuardMutation_PermiteDraft(t *testing.T) {
	e := Evidence{State: StateDraft}
	if err := GuardMutation(e); err != nil {
		t.Fatalf("no esperaba error en draft: %v", err)
	}
}

// --- Transiciones inválidas (ER-CORE-002) ---

func TestGuardTransition_RechazaTransicionInvalida(t *testing.T) {
	err := GuardTransition(StateVoided, StateIssued)
	if err == nil {
		t.Fatal("esperaba error en transición voided→issued")
	}
	if ExtractErrorCode(err) != "ER-CORE-002" {
		t.Fatalf("código incorrecto: %s", ExtractErrorCode(err))
	}
}

func TestGuardTransition_AceptaDraftAIssued(t *testing.T) {
	if err := GuardTransition(StateDraft, StateIssued); err != nil {
		t.Fatalf("no esperaba error en draft→issued: %v", err)
	}
}

// --- Void básico ---

func TestVoid_Basico(t *testing.T) {
	e := Evidence{
		ID:       "ev-001",
		TenantID: "tenant-001",
		State:    StateIssued,
	}
	now := time.Now()
	voided, err := Void(e, RCVoidErrorDetected, now)
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if voided.State != StateVoided {
		t.Fatalf("estado incorrecto: %s", voided.State)
	}
	if voided.VoidedAt == nil {
		t.Fatal("voided_at debe estar asignado")
	}
}

func TestVoid_RechazaReasonCodeInvalido(t *testing.T) {
	e := Evidence{State: StateIssued}
	_, err := Void(e, "INVALIDO", time.Now())
	if err == nil {
		t.Fatal("esperaba error por reason_code inválido")
	}
	if ExtractErrorCode(err) != "ER-CORE-003" {
		t.Fatalf("código incorrecto: %s", ExtractErrorCode(err))
	}
}

// --- Replace básico ---

func TestReplace_Basico(t *testing.T) {
	original := Evidence{ID: "ev-001", TenantID: "t-001", State: StateIssued}
	replacement := Evidence{ID: "ev-002", TenantID: "t-001", State: StateDraft}
	voided, issued, err := Replace(original, replacement, RCReplaceCorrection, time.Now())
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if voided.State != StateVoided {
		t.Fatalf("original debe ser voided, es: %s", voided.State)
	}
	if *voided.ReplacedByID != "ev-002" {
		t.Fatalf("replaced_by_id incorrecto: %s", *voided.ReplacedByID)
	}
	if issued.State != StateIssued {
		t.Fatalf("reemplazo debe ser issued, es: %s", issued.State)
	}
}

func TestReplace_RechazaIDVacio(t *testing.T) {
	original := Evidence{ID: "ev-001", State: StateIssued}
	replacement := Evidence{ID: ""}
	_, _, err := Replace(original, replacement, RCReplaceCorrection, time.Now())
	if err == nil {
		t.Fatal("esperaba error por ID vacío")
	}
	if ExtractErrorCode(err) != "ER-CORE-004" {
		t.Fatalf("código incorrecto: %s", ExtractErrorCode(err))
	}
}
