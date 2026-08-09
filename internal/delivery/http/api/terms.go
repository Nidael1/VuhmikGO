package api

// CurrentTermsVersion identifica la version vigente del documento de
// terminos y condiciones. El formato es <pais>-<rubro>-v<n>.
//
// DEUDA ARQUITECTONICA: esta constante es provisional. Las politicas de
// consentimiento pertenecen a la capa de Shaders (consent_policy_*,
// marcado FUTURO en 03_VUHMIK_SHADERS_reglas.md seccion 5.3). Cuando ese
// shader exista, la version debe resolverse por tenant_area y contexto
// legal, no por constante global. Requiere ADR.
const CurrentTermsVersion = "mx-medicine-v1"
