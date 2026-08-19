# VUHMÍK — Registro de Decisiones de Arquitectura (ADR)

Documento índice consolidado de todas las decisiones de arquitectura del
proyecto VUHMÍK (repositorio VuhmikGO). Cada ADR vive como archivo propio en
`docs/adr/`; este registro resume su estado, decisión y dependencias en un
solo lugar.

Última actualización: 2026-08-19

---

## Estado general

  Total planeados:  30
  Redactados:       30  (ADR-0001 a ADR-0030, en main)
  Pendientes:        0

Los ADR-0001 a 0009 son fundacionales (MVP y base del producto).
Los ADR-0010 a 0015 derivan del estudio de mercado de 14 países y definen
el roadmap clínico sobre el estándar IPS.
Los ADR-0016 a 0019 reafirman el Core agnóstico e introducen la capa de
administración (capacidades, toggles y métricas).

---

## Índice

  ADR-0001  Stack canónico Go                         Aceptado
  ADR-0002  Shaders por país y modo genérico          Aceptado
  ADR-0003  Estructura Go idiomática                  Aceptado
  ADR-0004  Frontend Vue SPA                          Aceptado
  ADR-0005  Autenticación JWT                         Aceptado
  ADR-0006  UX fluida / versionado silencioso         Aceptado
  ADR-0007  Formato de export (XML + JSON)            Aceptado
  ADR-0008  Firma digital / hash SHA-256              Aceptado
  ADR-0009  Traspaso de paciente entre tenants        Propuesto
  ADR-0010  IPS sobre FHIR (intercambio y contenido)  Propuesto
  ADR-0011  Módulo de medicación / receta             Propuesto
  ADR-0012  Módulo de alergias e intolerancias        Propuesto
  ADR-0013  Módulo de diagnósticos / CIE-10           Propuesto
  ADR-0014  Módulo de inmunizaciones                  Propuesto
  ADR-0015  Módulo de resultados de laboratorio       Propuesto
  ADR-0016  Core agnóstico (contenido opaco)          Aceptado
  ADR-0017  Registro de capacidades por tenant        Aceptado
  ADR-0018  Panel de toggles (escribe)                Aceptado
  ADR-0019  Panel de métricas (solo lectura)          Aceptado
  ADR-0020  Super admin                               Aceptado
  ADR-0021  Perfil profesional                        Aceptado
  ADR-0022  CQRS proyecciones                         Aceptado
  ADR-0023  Panel de actividad                        Aceptado
  ADR-0024  Módulo de consulta médica                 Aceptado
  ADR-0025  Modelo de datos tenant                    Aceptado
  ADR-0026  Referencia de vendedor                    Aceptado
  ADR-0027  Audit package ZIP sync                    Aceptado
  ADR-0028  IPS FHIR external import                  Aceptado
  ADR-0029  Stack despliegue Docker Coolify           Aceptado
  ADR-0030  Validación demográfica en Shader med_basic Aceptado

---

# ADRs fundacionales (0001–0009)

## ADR-0001 — Stack canónico Go
Estado: Aceptado · Fecha: 2026-06-02

Adopta Go como stack canónico de implementación, manteniendo intactos los
principios no negociables: un solo producto en tres capas (Engine, Shaders,
Asteroids); Core agnóstico, determinista e inmutable; acceso al Core solo vía
Shaders; evidencia append-only; export legal sin persistencia; observabilidad
sin PHI/PII; multi-tenant fail-closed; WAR-A con Redis y workers obligatorios.

Stack aprobado: Go 1.22+, net/http + ServeMux, DTOs tipados, pgx/v5 con SQL
explícito, golang-migrate forward-only, workers Go con cola Redis, StoragePort
como interfaz, logs JSON a stdout, Docker, Coolify + Hetzner. Rechaza ORM en el
Core (queries e invariantes deben ser explícitas y auditables) y frameworks
HTTP pesados (economía de guerra). El cambio de stack no autoriza nuevos
features ni cambios de arquitectura, lifecycle, tenancy o WAR-A.

## ADR-0002 — Shaders por país y modo genérico
Estado: Aceptado

Los Shaders son la única frontera de acceso al Core y la capa donde viven los
contratos y las políticas. Se admite un modo genérico (agnóstico) y modos
específicos por país/jurisdicción, sin que el Core conozca ninguno. El Core
permanece agnóstico; toda regla de dominio, país o profesión se resuelve en
Shaders.

## ADR-0003 — Estructura Go idiomática
Estado: Aceptado

Define la organización hexagonal del código: `internal/core` (Core inmutable),
`internal/shaders` (frontera contractual), `internal/application/ports`
(interfaces), `internal/infrastructure` (adaptadores: postgres, redis,
inmemory), `internal/delivery` (HTTP) y `internal/observability`. Separa
dominio de infraestructura mediante puertos e interfaces.

## ADR-0004 — Frontend Vue SPA
Estado: Aceptado

Adopta Vue 3 + Vite + TypeScript + Pinia + Vue Router como SPA para la capa de
Asteroides (UX). El frontend consume la API JSON `/api/v1` y no contiene
lógica de dominio: es presentación pura. Organización por capas
(domain/types, infrastructure/repositories, presentation/views).

## ADR-0005 — Autenticación JWT
Estado: Aceptado

Establece autenticación basada en JWT para la API. El token porta el contexto
de tenant y actor que el middleware inyecta en cada request. Base del
aislamiento multi-tenant en la capa de entrega.

## ADR-0006 — UX fluida / versionado silencioso
Estado: Aceptado

El médico percibe edición libre del expediente. Internamente, el sistema
ejecuta void + replace silencioso: cada consulta es una nota nueva, y "editar"
es corregir mediante anulación + reemplazo transparente. El médico NO ve los
estados internos (draft/issued/locked/voided); las notas anuladas son
invisibles para él pero permanecen en BD para trazabilidad COFEPRIS. El Core
no cambia: solo ve Create. Es el principio de experiencia que hace adoptable
un Core append-only.

## ADR-0007 — Formato de export (XML + JSON)
Estado: Aceptado

El export legal se genera en dos formatos: JSON (uso técnico) y XML (legal /
auditoría / COFEPRIS, basado en HL7 CDA simplificado, namespace
`urn:vuhmik:hl7:v1`). El formato se detecta por el header Accept. Ambos se
generan en memoria, sin persistencia del archivo, con Cache-Control: no-store.
Nota: el ADR-0010 reorienta este esquema CDA propio hacia el perfil IPS.

## ADR-0008 — Firma digital / hash SHA-256
Estado: Aceptado

Define la integridad probatoria de la evidencia mediante hash SHA-256 canónico.
Fase 1 (hash canónico del contenido) implementada; el export incluye el hash
verificable sin acceso al servidor. Fase 2 (HMAC) y Fase 3 (certificado SAT /
firma electrónica avanzada) quedan previstas para v1.5/v2.

## ADR-0009 — Traspaso de paciente entre tenants
Estado: Propuesto · Fecha: 2026-06-22

El traspaso NO comparte registros entre tenants; genera una copia en el tenant
destino, dejando los originales intactos e inmutables en el origen. Flujo:
el médico A exporta el expediente (XML con hash); el archivo se entrega por
canal externo; el médico B lo importa; el sistema verifica el hash y crea un
paciente nuevo con evidencias en estado issued, referenciando el origen
(import_source, import_ref). El CURP es el identificador de continuidad
asistencial: si ya existe en el destino, el sistema pregunta fusionar o crear.
El Core solo ve Create — no conoce el concepto de traspaso.

Dependencias: ADR-0007, ADR-0008, CURP en patients.

---

# ADRs del roadmap clínico — estándar IPS (0010–0015)

Derivados del estudio de mercado de 14 países (Japón SS-MIX2, Taiwán NHI
MediCloud, Corea MyHealthWay, Brasil RNDS, Turquía e-Nabız, Indonesia
SATUSEHAT, Singapur NEHR, Malasia, India ABDM, Tailandia, Vietnam, Filipinas,
Omán y el Servicio Universal de Salud de México). Conclusión central: el
patrón "exportador / receptor entre médicos" es el International Patient
Summary (IPS) sobre FHIR, y sus secciones definen el roadmap de funciones.

## ADR-0010 — IPS sobre FHIR (keystone)
Estado: Propuesto · Fecha: 2026-06-24

Adopta el International Patient Summary (IPS) sobre FHIR R4 como modelo canónico
del contenido clínico estructurado y del intercambio entre prestadores. El IPS
se serializa en JSON o XML; conserva la capacidad dual de ADR-0007 pero migra
del esquema CDA propio al perfil IPS estándar. Restricción arquitectónica: el
IPS vive en la frontera de Shaders y en la capa de export, NUNCA en el Core;
el Core sigue almacenando registros append-only genéricos, el Shader los
proyecta a documento IPS al exportar y los valida al recibir. El hash SHA-256
se aplica al documento IPS, volviéndolo un documento clínico verificable.
Define las secciones como roadmap: Problemas, Alergias y Medicación
(obligatorias); Inmunizaciones y Resultados (recomendadas).

Dependencias: ADR-0007, ADR-0008, ADR-0009.

## ADR-0011 — Módulo de medicación / receta electrónica
Estado: Propuesto · Fecha: 2026-06-24

Sección IPS obligatoria (Medication Summary) y mínimo funcional de la NOM-024.
Una receta es un registro append-only en el Core; el Shader valida los campos
mínimos de validez legal (cédula profesional, especialidad, paciente,
medicamento genérico, dosis) antes de emitir. Lifecycle draft → issued →
locked; correcciones solo por void + replace. Los medicamentos controlados
COFEPRIS (Grupo II/III) quedan EXCLUIDOS de v1: requieren validación
regulatoria explícita; el sistema muestra advertencia para usar recetario
especial en papel. Se modela como MedicationStatement / MedicationRequest del
IPS. Cada receta se vincula obligatoriamente a un paciente y se firma con hash
al emitir.

Dependencias: ADR-0006, ADR-0008, ADR-0009, ADR-0010.

## ADR-0012 — Módulo de alergias e intolerancias
Estado: Propuesto · Fecha: 2026-06-24

Segunda sección IPS obligatoria (AllergyIntolerance). Registro append-only en
el Core. Campos: agente y tipo de reacción (obligatorios), criticidad, certeza,
fecha (opcionales); el agente es texto libre en v1, codificación SNOMED/RxNorm
en v2. Las alergias activas se muestran de forma prominente en el detalle del
paciente, antes del historial de notas, y al componer una receta (informativo
en v1, bloqueante en v2). Correcciones por void + replace; el historial
completo permanece.

Dependencias: ADR-0006, ADR-0008, ADR-0009, ADR-0010, ADR-0011.

## ADR-0013 — Módulo de diagnósticos / lista de problemas (CIE-10)
Estado: Propuesto · Fecha: 2026-06-24

Tercera sección IPS obligatoria (Problem List). Convierte las notas de texto
libre en diagnósticos estructurados. Registro append-only en el Core. El código
CIE-10 es opcional en v1 (adopción rápida) y obligatorio en v2 (perfil IPS
completo + certificación CENETEC). Se modela como Condition del IPS, separando
problemas activos (Problem List) de resueltos (History of Past Illness). Un
diagnóstico NO reemplaza a una nota clínica: coexisten como registros
complementarios.

Dependencias: ADR-0006, ADR-0008, ADR-0009, ADR-0010, ADR-0011.

## ADR-0014 — Módulo de inmunizaciones / vacunación
Estado: Propuesto · Fecha: 2026-06-24

Primera sección IPS recomendada (Immunizations). Alto valor, bajo costo.
Registro append-only en el Core. Campos: vacuna y fecha (obligatorios), lote,
dosis, vía, aplicada_por (opcionales); vacuna en texto libre en v1, CVX/SNOMED
en v2. Distingue vacuna aplicada en consulta de la reportada por el paciente.
Sin integración con SINAVE en v1: el médico privado registra lo que aplica o lo
que el paciente reporta. Se modela como Immunization del IPS.

Dependencias: ADR-0006, ADR-0008, ADR-0009, ADR-0010.

## ADR-0015 — Módulo de resultados de laboratorio
Estado: Propuesto · Fecha: 2026-06-24

Segunda sección IPS recomendada (Results) y segundo mínimo funcional de la
NOM-024. Registro append-only en el Core. Cubre laboratorio clínico; radiología
e imágenes DICOM quedan fuera de alcance. Campos: estudio, fecha y resultado
(obligatorios); laboratorio, solicitante, referencia al diagnóstico, rangos,
interpretación (opcionales); estudio en texto libre en v1, LOINC en v2. v1 NO
almacena archivos PDF: solo una referencia externa (archivo_ref); el archivo
vive fuera de VUHMÍK. Se modela como Observation (Results: laboratory) del IPS.

Dependencias: ADR-0006, ADR-0008, ADR-0009, ADR-0010, ADR-0013.

---

# ADRs pendientes — Core agnóstico y capa de administración (0016–0019)

Cierran la coherencia del Core agnóstico e introducen la administración de la
plataforma (capacidades, toggles y métricas), separando estrictamente lo que
escribe de lo que solo lee.

## ADR-0016 — Core agnóstico: contenido opaco y type en Shader
Estado: Pendiente de redacción

Reafirma que el Core es agnóstico de dominio dentro de la familia de sistemas
que requieren un registro probatorio e inmutable (no solo medicina: el mismo
Core puede ser ECE, ERP, CRM o sistema notarial). Consecuencia técnica: el
Core guarda un contenido opaco (blob) que nunca interpreta; el discriminador
de tipo (note, prescription, allergy, etc.) vive DENTRO del contenido y solo
el Shader lo lee. Resuelve la deuda del campo Notes (semántica médica filtrada
al Core, debe salir) y reinterpreta SubjectID como clave de correlación opaca.
El hash (ADR-0008) pasa a calcularse sobre el blob como bytes opacos. Declara
que las secciones de implementación de los ADR-0011 a 0015 se reinterpretan
sobre este modelo: no son tablas tipadas en el Core, sino tipos de contenido
que el Shader interpreta sobre el registro genérico.

## ADR-0017 — Registro de capacidades por tenant
Estado: Pendiente de redacción

Define el registro que enumera qué Shaders y Asteroides están activos por
tenant. Postura fail-closed: nada está activo salvo lo explícitamente
encendido. El Shader consulta este registro ANTES de tocar el Core; si un
módulo no está activo para el tenant, niega el acceso (compuerta real,
coherente con "todo acceso al Core pasa por Shaders"). El mismo registro es la
base de facturación: los módulos activos definen el plan y el costo. Es la
fuente de verdad única que leen tanto el toggle (escribe) como las métricas
(lee).

## ADR-0018 — Panel de toggles (escribe)
Estado: Pendiente de redacción

Panel de administración que activa y desactiva Shaders y Asteroides por cuenta,
operando sobre el registro de capacidades (ADR-0017). Es la superficie de
escritura más privilegiada del sistema: cambia qué accesos al Core existen.
Muestra el catálogo completo de Shaders y Asteroides, su estado (activo /
inactivo) por cuenta y el conteo de cuántos hay activos de cada tipo (modelo
tipo Odoo: el médico activa lo que necesita en vez de tener todo encendido).
Alcance estrictamente administrativo-comercial: gestiona módulos, planes,
costos y estado de cuenta; NUNCA accede a PHI ni al contenido del expediente.
Introduce la bandera is_admin. La suspensión por falta de pago bloquea el login
del médico pero nunca borra ni altera datos (el expediente permanece intacto,
append-only). El super-admin de rubro (define si la instancia es ECE, ERP,
CRM…) queda documentado como futuro y NO se construye en v1, porque la
instancia actual ya está fijada como ECE médico.

## ADR-0019 — Panel de métricas (solo lectura)
Estado: Pendiente de redacción

Panel de inteligencia de negocio, separado del de toggles porque solo lee y no
afecta nada. Muestra la lista de doctores con conteos (pacientes, notas,
recetas por periodo), estado de cuenta, plan y costo; y métricas agregadas de
negocio (cuentas activas/canceladas, MRR, churn, módulos más y menos activados,
distribución de uso). Frontera absoluta: muestra números y tendencias, NUNCA el
contenido detrás del número (quiénes son los pacientes, qué se recetó). Las
métricas se calculan como agregados (count, sum), coherente con la regla de
observabilidad sin PHI/PII, y se precalculan con un worker WAR-A. No modifica
nada del sistema.

---

# Notas de coherencia pendientes

  1. Deuda de Notes en el Core: el campo Notes (semántica médica) debe salir
     del Core al adoptar el ADR-0016. Migración forward-only que envuelve el
     contenido existente en el blob opaco.

  2. Reinterpretación de ADR-0011 a 0015: sus secciones de implementación
     mencionan tablas tipadas "en el Core"; con el ADR-0016 se reinterpretan
     como tipos de contenido opaco interpretados por el Shader. No requieren
     reescritura completa; el ADR-0016 lo declara explícitamente.

  3. ADR-0007 (export CDA propio) queda como legado a deprecar tras el
     ADR-0010 (IPS sobre FHIR). Migración progresiva, no eliminación inmediata.

  4. Firma legal de la receta: v1 usa cédula profesional + hash; la firma
     electrónica avanzada (e.firma/SAT) queda para v2 (ADR-0008 fase 3).

---

# Orden de ejecución sugerido

  1. ADR-0016  (base: Core agnóstico — del que dependen 0017, 0018, 0019)
  2. ADR-0017  (registro de capacidades — compuerta y base de facturación)
  3. ADR-0018  (panel de toggles — escribe)
  4. ADR-0019  (panel de métricas — solo lectura)
  5. Sprint 9  (primer módulo clínico sobre el Core agnóstico ya correcto)

Disciplina: un ADR = una rama = un commit = un PR = un merge.
