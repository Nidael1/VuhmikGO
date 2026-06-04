package shaders

import (
	"testing"
	"time"
)

// --- Validación de ShaderContext ---

func TestShaderContext_RechazaSinTenantID(t *testing.T) {
	ctx := ShaderContext{Operation: OperationCreate, ActorID: "actor-1"}
	if err := ctx.Validate(); err == nil {
		t.Fatal("esperaba error por tenant_id ausente")
	}
}

func TestShaderContext_RechazaSinOperation(t *testing.T) {
	ctx := ShaderContext{TenantID: "tenant-1", ActorID: "actor-1"}
	if err := ctx.Validate(); err == nil {
		t.Fatal("esperaba error por operation ausente")
	}
}

func TestShaderContext_RechazaSinActorID(t *testing.T) {
	ctx := ShaderContext{TenantID: "tenant-1", Operation: OperationCreate}
	if err := ctx.Validate(); err == nil {
		t.Fatal("esperaba error por actor_id ausente")
	}
}

func TestShaderContext_AceptaContextoValido(t *testing.T) {
	ctx := ShaderContext{TenantID: "t-1", Operation: OperationCreate, ActorID: "a-1"}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
}

// --- MedicalBasicShader ---

func TestMedicalBasicShader_PermiteOperacionesValidas(t *testing.T) {
	s := NewMedicalBasicShader()
	ops := []Operation{OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport}
	for _, op := range ops {
		ctx := ShaderContext{TenantID: "t-1", Operation: op, ActorID: "a-1"}
		d := s.Evaluate(ctx)
		if d.Result != DecisionAllow {
			t.Errorf("operación %s debe ser permitida, fue: %s", op, d.Result)
		}
	}
}

func TestMedicalBasicShader_DeniegaContextoInvalido(t *testing.T) {
	s := NewMedicalBasicShader()
	d := s.Evaluate(ShaderContext{})
	if d.Result != DecisionDeny {
		t.Fatal("esperaba deny por contexto vacío")
	}
	if d.ErrorCode != errShaderContextInvalid {
		t.Fatalf("código incorrecto: %s", d.ErrorCode)
	}
}

// --- LegalExportShader ---

func TestLegalExportShader_PermiteExport(t *testing.T) {
	s := NewLegalExportShader()
	ctx := ShaderContext{TenantID: "t-1", Operation: OperationExport, ActorID: "a-1"}
	d := s.Evaluate(ctx)
	if d.Result != DecisionAllow {
		t.Fatalf("esperaba allow en export: %s", d.Reason)
	}
}

func TestLegalExportShader_DeniegaOperacionNoExport(t *testing.T) {
	s := NewLegalExportShader()
	ctx := ShaderContext{TenantID: "t-1", Operation: OperationCreate, ActorID: "a-1"}
	d := s.Evaluate(ctx)
	if d.Result != DecisionDeny {
		t.Fatal("export shader solo debe permitir export")
	}
}

func TestLegalExportShader_GeneraExportEnMemoria(t *testing.T) {
	s := NewLegalExportShader()
	ctx := ShaderContext{TenantID: "t-1", Operation: OperationExport, ActorID: "a-1"}
	data := ExportData{
		EvidenceID: "ev-001",
		TenantID:   "t-1",
		State:      "issued",
		CreatedAt:  time.Now(),
	}
	b, err := s.GenerateExport(ctx, data)
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("export no debe estar vacío")
	}
}

func TestLegalExportShader_RechazaExportSinContexto(t *testing.T) {
	s := NewLegalExportShader()
	_, err := s.GenerateExport(ShaderContext{}, ExportData{})
	if err == nil {
		t.Fatal("esperaba error con contexto inválido")
	}
}
