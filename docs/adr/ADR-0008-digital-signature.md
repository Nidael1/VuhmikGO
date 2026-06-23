# ADR-0008 — Mecanismo de integridad y firma digital

## Estado
Propuesto

## Fecha
2026-06-22

## Contexto

El export clínico (ADR-0007) incluye un hash SHA-256 para verificar
integridad. Sin embargo, el hash solo garantiza que el archivo no fue
modificado despues de generarse — no garantiza que fue generado por
el servidor de VUHMIK y no por un tercero.

Para que el export sea legalmente probatorio ante COFEPRIS y en
procedimientos judiciales, se requiere un mecanismo de firma digital
que permita verificar la autenticidad del documento.

## Decisión

### Fase 1 — Hash SHA-256 (v1, implementacion inmediata)

El export incluye un hash SHA-256 del contenido canonico.
El servidor puede recalcular el hash en cualquier momento para
verificar que el documento es identico al que genero.
Esta fase no requiere certificados ni infraestructura de PKI.

### Fase 2 — Firma HMAC-SHA256 con clave del servidor (v1.5)

El servidor firma el hash con su clave privada (JWT_SECRET o
clave dedicada). El receptor puede verificar la firma si tiene
acceso a la clave publica del servidor.
Esta fase permite verificacion sin acceso a la BD.

### Fase 3 — Firma digital con certificado (v2)

Firma con certificado digital reconocido por el SAT o por la
Secretaria de Salud. Nivel de prueba maxima en procedimientos
legales mexicanos.
Esta fase requiere contratar un proveedor de certificacion.

## Implementacion en v1

Solo se implementa Fase 1 (hash SHA-256).
Las fases 2 y 3 requieren ADR de implementacion y presupuesto.

## Algoritmo de hash canonico

  1. Extraer todos los campos del documento excepto "hash"
     y "exported_at".
  2. Serializar como JSON con claves en orden alfabetico,
     sin espacios ni saltos de linea.
  3. Calcular SHA-256 del string UTF-8 resultante.
  4. Incluir como "sha256:<hex_lowercase>".

## Consecuencias

  El servidor puede verificar cualquier export en cualquier momento.
  El medico puede detectar si alguien modifico su documento.
  COFEPRIS puede solicitar verificacion al servidor.
  No requiere infraestructura adicional en v1.
