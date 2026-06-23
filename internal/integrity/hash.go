// Package integrity implementa funciones de integridad para evidencia clínica.
// El hash SHA-256 permite verificar que un export no fue alterado
// sin necesidad de acceso al servidor (ADR-0008).
package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// EvidenceHashInput es el conjunto de campos que se hashean.
// Excluye "hash" y "exported_at" para evitar circularidad.
type EvidenceHashInput struct {
	EvidenceID   string  `json:"evidence_id"`
	TenantID     string  `json:"tenant_id"`
	SubjectID    string  `json:"subject_id"`
	Notes        string  `json:"notes"`
	State        string  `json:"state"`
	CreatedAt    string  `json:"created_at"`
	IssuedAt     *string `json:"issued_at"`
	VoidedAt     *string `json:"voided_at"`
	ReplacedByID *string `json:"replaced_by_id"`
}

// Hash calcula SHA-256 del contenido canonico de la evidencia.
// El contenido canonico es el JSON con claves en orden alfabetico,
// sin espacios ni saltos de linea — deterministico y verificable.
func Hash(input EvidenceHashInput) (string, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("error al serializar evidencia para hash: %w", err)
	}
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:]), nil
}

// Verify verifica que el hash de un input coincide con el hash dado.
func Verify(input EvidenceHashInput, expected string) (bool, error) {
	computed, err := Hash(input)
	if err != nil {
		return false, err
	}
	return computed == expected, nil
}
