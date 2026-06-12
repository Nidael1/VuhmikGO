package main

import (
	"log"
	"net/http"

	delivery "github.com/Nidael1/VuhmikGO/internal/delivery/http"
	"github.com/Nidael1/VuhmikGO/internal/observability"
)

func main() {
	if err := observability.ValidateRuntimeSecrets(); err != nil {
		log.Fatalf("error de configuración: %v", err)
	}

	mux := http.NewServeMux()
	delivery.RegisterRoutes(mux)
	handler := delivery.Handler(mux)

	observability.Logger.Info("servidor iniciado", "addr", ":8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("error al iniciar servidor: %v", err)
	}
}
