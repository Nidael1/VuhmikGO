package application

import (
	"testing"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/inmemory"
)

// TestAislamientoMultiTenant_FindByID verifica que un tenant no puede
// leer registros de otro tenant aunque conozca el ID exacto (Issue #56).
func TestAislamientoMultiTenant_FindByID(t *testing.T) {
	repo := inmemory.NewEvidenceRepository()
	svc := NewECEService(repo)

	if _, err := svc.CreateDraft("ev-001", "tenant-A"); err != nil {
		t.Fatalf("no esperaba error al crear: %v", err)
	}

	// tenant-A puede acceder a su propio registro
	if _, err := repo.FindByID("tenant-A", "ev-001"); err != nil {
		t.Fatalf("tenant-A debe poder leer su propio registro: %v", err)
	}

	// tenant-B NO debe poder acceder al registro de tenant-A
	if _, err := repo.FindByID("tenant-B", "ev-001"); err == nil {
		t.Fatal("tenant-B no debe poder leer un registro de tenant-A")
	}
}

// TestAislamientoMultiTenant_Issue verifica que Issue() falla si el
// tenantID no coincide con el propietario del registro (Issue #56).
func TestAislamientoMultiTenant_Issue(t *testing.T) {
	repo := inmemory.NewEvidenceRepository()
	svc := NewECEService(repo)

	if _, err := svc.CreateDraft("ev-002", "tenant-A"); err != nil {
		t.Fatalf("no esperaba error al crear: %v", err)
	}

	// tenant-B intenta emitir un registro de tenant-A — debe fallar
	if _, err := svc.Issue("tenant-B", "ev-002"); err == nil {
		t.Fatal("tenant-B no debe poder emitir un registro de tenant-A")
	}

	// tenant-A sí puede emitir su propio registro
	issued, err := svc.Issue("tenant-A", "ev-002")
	if err != nil {
		t.Fatalf("tenant-A debe poder emitir su propio registro: %v", err)
	}
	if issued.State != evidence.StateIssued {
		t.Fatalf("estado incorrecto: %s", issued.State)
	}
}
