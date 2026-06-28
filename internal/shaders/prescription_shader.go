package shaders

import "encoding/json"

// PrescriptionShader es el Shader para el modulo de receta electronica.
//
// El Core conserva las invariantes (append-only, lifecycle, hash).
// El Shader valida los campos minimos de validez legal (NOM-024-SSA3-2012):
// cedula_profesional, especialidad, patient_id, medicamento_generico, dosis.
// Perfil: prescription — ADR-0011.
type PrescriptionShader struct{}

// PrescriptionContent es la estructura del blob para registros tipo prescription.
// El Core nunca la interpreta; solo el Shader y el Asteroide la conocen.
type PrescriptionContent struct {
	Type                string `json:"type"`
	CedulaProfesional   string `json:"cedula_profesional"`
	Especialidad        string `json:"especialidad"`
	MedicamentoGenerico string `json:"medicamento_generico"`
	Dosis               string `json:"dosis"`
	Diagnostico         string `json:"diagnostico,omitempty"`
	Indicaciones        string `json:"indicaciones,omitempty"`
	Seguimiento         string `json:"seguimiento,omitempty"`
	// Advertencia COFEPRIS: si el medicamento es controlado,
	// el medico debe usar recetario especial en papel (ADR-0011).
	EsControlado bool `json:"es_controlado,omitempty"`
}

// Evaluate evalua si la operacion esta permitida para el modulo prescription.
func (s *PrescriptionShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderContextInvalid,
			Reason:    err.Error(),
		}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport:
		return ShaderDecision{Result: DecisionAllow, Reason: "operacion permitida en modulo prescription"}
	default:
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "operacion no reconocida en modulo prescription",
		}
	}
}

// ValidatePrescriptionContent valida los campos minimos NOM-024.
// Obligatorios antes de emitir: cedula, especialidad, medicamento, dosis.
func ValidatePrescriptionContent(c PrescriptionContent) error {
	if c.CedulaProfesional == "" {
		return &missingFieldError{"cedula_profesional"}
	}
	if c.Especialidad == "" {
		return &missingFieldError{"especialidad"}
	}
	if c.MedicamentoGenerico == "" {
		return &missingFieldError{"medicamento_generico"}
	}
	if c.Dosis == "" {
		return &missingFieldError{"dosis"}
	}
	return nil
}

// BuildPrescriptionBlob construye el blob JSON opaco para el Core.
func BuildPrescriptionBlob(c PrescriptionContent) (string, error) {
	c.Type = "prescription"
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParsePrescriptionBlob parsea el blob JSON de una receta.
func ParsePrescriptionBlob(blob string, c *PrescriptionContent) error {
	return json.Unmarshal([]byte(blob), c)
}

// NewPrescriptionShader retorna una instancia del Shader de recetas.
func NewPrescriptionShader() Shader {
	return &PrescriptionShader{}
}
