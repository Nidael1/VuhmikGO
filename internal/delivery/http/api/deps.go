package api

import (
	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
)

// Deps contiene las dependencias inyectadas en los handlers de la API.
type Deps struct {
	EvidenceRepo ports.EvidenceRepository
	UserRepo     *postgres.UserRepository
}

// deps es la instancia global de dependencias, inicializada desde main.go.
var deps Deps

// InitDeps inicializa las dependencias de la API.
// Debe llamarse desde main.go antes de registrar rutas.
func InitDeps(d Deps) {
	deps = d
}
