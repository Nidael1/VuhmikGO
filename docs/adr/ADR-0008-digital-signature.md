# ADR-0008 — Mecanismo de integridad y firma digital

## Estado
Diferido — decision consciente, no gap

## Fecha
2026-06-22

## Actualizacion
2026-07-09

## Contexto

El export clínico (ADR-0007) incluye un hash SHA-256 para verificar
integridad. Sin embargo, el hash solo garantiza que el archivo no fue
modificado despues de generarse — no garantiza que fue generado por
el servidor de VUHMIK y no por un tercero.

Para que el export sea legalmente probatorio ante COFEPRIS y en
procedimientos judiciales, se requiere un mecanismo de firma digital
que permita verificar la autenticidad del documento.

## Decision (actualizada 2026-07-09)

La firma digital criptografica con PKI NO se implementa en v1.

Razon: ninguna ley o norma mexicana vigente exige firma digital
criptografica para el expediente clinico de medicos independientes.
La NOM-004-SSA3-2010 y la NOM-024-SSA3-2010 exigen integridad,
autoria y no alteracion — garantias que ya provee la arquitectura
append-only de VUHMIK (registros inmutables, void+replace, chain
de replaced_by_id, timestamps, backup diario, hash en traspaso).

La firma con PKI requiere gestion de claves que introduce complejidad
operativa sin beneficio proporcional para medicos independientes en v1.
Implementarla mal (clave privada en servidor) es peor que no tenerla.

## Implementacion en v1

No se implementa. Estado cerrado como Diferido — decision explicita,
no deuda tecnica pendiente.

## Cuando reconsiderar

  - Una norma exige firma digital para medicos independientes.
  - VUHMIK requiere interoperabilidad con IMSS/ISSSTE/SSA.
  - Un caso juridico especifico requiere no-repudio criptografico.

En ese momento: Fase 2 (HMAC-SHA256 servidor) antes de Fase 3
(certificado SAT/Secretaria de Salud).

## Consecuencias

  La arquitectura append-only provee garantias suficientes para v1.
  No se agrega complejidad de gestion de claves.
  Esta decision no afecta el lanzamiento de v1.
