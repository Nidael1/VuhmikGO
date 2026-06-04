package main

import (
	"log"
	"net/http"

	delivery "github.com/Nidael1/VuhmikGO/internal/delivery/http"
)

func main() {
	mux := http.NewServeMux()
	delivery.RegisterRoutes(mux)

	log.Println("servidor iniciado en :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("error al iniciar servidor: %v", err)
	}
}
