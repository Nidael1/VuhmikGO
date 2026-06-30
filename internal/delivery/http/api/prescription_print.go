package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// HandlePrescriptionPrint genera HTML imprimible de la receta (NOM-024).
// No persiste el PDF: se regenera bajo demanda en cada solicitud (economia
// de guerra — sin costo de storage adicional en el VPS).
//
// GET /api/v1/prescriptions/:id/print
//
// Esta ruta acepta el token por header Authorization o por query param
// "token", porque se abre típicamente en una pestaña nueva del navegador
// (window.open) donde no es posible adjuntar headers personalizados.
func HandlePrescriptionPrint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenStr := ""
	header := r.Header.Get("Authorization")
	if len(header) >= 8 && header[:7] == "Bearer " {
		tokenStr = header[7:]
	} else if q := r.URL.Query().Get("token"); q != "" {
		tokenStr = q
	}
	if tokenStr == "" {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "token invalido o expirado", http.StatusUnauthorized)
		return
	}
	tenantID := claims.TenantID
	actorID := claims.ActorID
	if tenantID == "" {
		http.Error(w, "no autenticado", http.StatusUnauthorized)
		return
	}

	prescriptionID := strings.TrimPrefix(r.URL.Path, "/api/v1/prescriptions/")
	prescriptionID = strings.TrimSuffix(prescriptionID, "/print")

	// Obtener receta
	rx, err := deps.PrescriptionService.FindByID(tenantID, prescriptionID)
	if err != nil {
		http.Error(w, "receta no encontrada", http.StatusNotFound)
		return
	}

	// Perfil del médico
	profile, _ := deps.ProfileRepo.Get(actorID)

	// Paciente
	patient, _ := deps.PatientRepo.FindByID(tenantID, rx.PatientID)

	// Fecha
	fecha := rx.CreatedAt.Format("02/01/2006")
	if rx.IssuedAt != nil {
		fecha = rx.IssuedAt.Format("02/01/2006")
	}

	// Edad
	edad := ""
	if patient.FechaNacimiento != "" {
		if fn, err2 := time.Parse("2006-01-02", patient.FechaNacimiento[:10]); err2 == nil {
			years := int(time.Since(fn).Hours() / 8766)
			if years > 0 {
				edad = fmt.Sprintf("%d años", years)
			}
		}
	}

	// Signos vitales desde consulta vinculada (si existe)
	ta, fc, fr, temp, peso, talla, sao2 := "", "", "", "", "", "", ""
	if rx.ConsultationID != "" {
		con, err2 := deps.ConsultationProjectionRepo.FindByID(tenantID, rx.ConsultationID)
		if err2 == nil {
			ta, fc, fr, temp, peso, talla, sao2 = con.TA, con.FC, con.FR, con.Temp, con.Peso, con.Talla, con.SAO2
		}
	}

	html := buildPrescriptionHTML(
		profile.NombreCompleto, profile.Universidad, profile.CedulaProfesional,
		profile.Especialidad, profile.Direccion, profile.Telefono,
		patient.Nombre, edad, fecha,
		ta, fc, fr, temp, peso, talla, sao2,
		rx.MedicamentoGenerico, rx.Dosis, rx.Indicaciones, rx.Seguimiento, rx.Diagnostico,
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func vitalCell(label, value, unit string) string {
	display := "______"
	if value != "" {
		display = value + " " + unit
	}
	return fmt.Sprintf(`<tr><td class="vl">%s:</td><td class="vv">%s</td></tr>`, label, display)
}

func buildPrescriptionHTML(
	nombre, universidad, cedula, especialidad, direccion, telefono string,
	paciente, edad, fecha string,
	ta, fc, fr, temp, peso, talla, sao2 string,
	medicamento, dosis, indicaciones, seguimiento, diagnostico string,
) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<title>Receta — %s</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: Arial, sans-serif; font-size: 12px; color: #111; background: #fff; }
  .receta { border: 2px solid #111; border-radius: 10px; max-width: 820px; margin: 20px auto; padding: 0; overflow: hidden; }
  .header { text-align: center; padding: 14px 20px 10px; border-bottom: 1.5px solid #111; }
  .header h1 { font-size: 18px; font-weight: 700; margin-bottom: 3px; }
  .header p { font-size: 11px; color: #333; }
  .paciente-row { display: flex; justify-content: space-between; padding: 8px 20px; border-bottom: 1.5px solid #111; font-size: 11.5px; }
  .paciente-row span { flex: 1; }
  .paciente-row .label { font-weight: 700; margin-right: 4px; }
  .body { display: grid; grid-template-columns: 1fr 2fr 1fr; border-bottom: 1.5px solid #111; }
  .col { padding: 12px 14px; }
  .col + .col { border-left: 1.5px solid #111; }
  .col-title { font-weight: 700; font-size: 11px; text-transform: uppercase; letter-spacing: .05em; border-bottom: 1px solid #ccc; padding-bottom: 5px; margin-bottom: 8px; }
  .vl { font-weight: 700; padding-right: 6px; font-size: 11px; white-space: nowrap; padding-top: 3px; }
  .vv { font-size: 11px; padding-top: 3px; border-bottom: 1px solid #ccc; min-width: 80px; }
  .med { font-size: 13px; font-weight: 700; margin-bottom: 6px; }
  .sub-title { font-weight: 700; font-size: 11px; margin-top: 10px; margin-bottom: 3px; text-transform: uppercase; letter-spacing: .04em; }
  .dx-col { display: flex; flex-direction: column; justify-content: space-between; }
  .footer { padding: 10px 20px; display: flex; justify-content: space-between; align-items: flex-end; font-size: 11px; }
  .firma-line { border-top: 1px solid #111; width: 160px; text-align: center; padding-top: 3px; font-size: 10px; color: #444; }
  @media print {
    body { margin: 0; }
    .receta { border-radius: 0; margin: 0; max-width: 100%%; border: 2px solid #111; }
    @page { margin: 10mm; size: A5 landscape; }
  }
</style>
</head>
<body>
<div class="receta">
  <div class="header">
    <h1>%s</h1>
    <p>%s &nbsp;·&nbsp; CÉD. PROF. %s</p>
    <p>%s</p>
  </div>

  <div class="paciente-row">
    <span><span class="label">PACIENTE:</span>%s</span>
    <span><span class="label">EDAD:</span>%s</span>
    <span style="text-align:right"><span class="label">FECHA:</span>%s</span>
  </div>

  <div class="body">
    <!-- Signos vitales -->
    <div class="col">
      <div class="col-title">Signos</div>
      <table style="width:100%%">
        %s
        %s
        %s
        %s
        %s
        %s
        %s
      </table>
    </div>

    <!-- Medicamentos e indicaciones -->
    <div class="col">
      <div class="col-title">Medicamentos (RX)</div>
      <div class="med">%s</div>
      <div style="font-size:11px;color:#333">%s</div>
      <div class="sub-title">Indicaciones</div>
      <div style="font-size:11px">%s</div>
      <div class="sub-title">Seguimiento</div>
      <div style="font-size:11px">%s</div>
    </div>

    <!-- Diagnóstico -->
    <div class="col dx-col">
      <div>
        <div class="col-title">Diagnóstico</div>
        <div style="font-size:11px">%s</div>
      </div>
      <div>
        <div style="font-size:11px;margin-bottom:4px">Tel.: %s</div>
      </div>
    </div>
  </div>

  <div class="footer">
    <span></span>
    <div class="firma-line">Firma del médico</div>
  </div>
</div>
<script>window.onload = function() { window.print(); }</script>
</body>
</html>`,
		paciente,
		nombre, universidad, cedula, direccion,
		paciente, edad, fecha,
		vitalCell("T/A", ta, "mmHg"),
		vitalCell("FC", fc, "lpm"),
		vitalCell("FR", fr, "rpm"),
		vitalCell("TEMP", temp, "°C"),
		vitalCell("PESO", peso, "kg"),
		vitalCell("TALLA", talla, "m"),
		vitalCell("SAO2", sao2, "%"),
		medicamento, dosis, indicaciones, seguimiento,
		diagnostico, telefono,
	)
}
