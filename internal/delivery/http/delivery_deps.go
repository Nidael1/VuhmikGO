package delivery

import "github.com/Nidael1/VuhmikGO/internal/application/ports"

// DeliveryDeps contiene las dependencias inyectadas en los handlers de entrega.
// Sigue el mismo patrón que api.Deps (ADR-0003 §3, monolito modular hexagonal).
type DeliveryDeps struct {
	TenantRepo ports.TenantRepository
}

var deliveryDeps DeliveryDeps

// InitDeliveryDeps inicializa las dependencias del paquete delivery.
// Llamar desde main.go antes de RegisterRoutes.
func InitDeliveryDeps(d DeliveryDeps) {
	deliveryDeps = d
}
