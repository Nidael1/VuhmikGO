# ADR-0031 — Módulo de agendamiento de citas

## Estado
Aceptado

## Fecha
2026-08-19

## Contexto

El sistema tiene consultas (ADR-0024) como unidad clínica central, y un widget
de agenda en el sidebar que muestra las consultas del día (issues #163/#164).

Sin embargo no existe ningún mecanismo para agendar una cita con anticipación.
El médico no puede reservar un horario futuro para un paciente, y el widget
siempre muestra "Sin consultas registradas" porque no hay datos que alimentarlo
para el día actual a menos que la consulta ya haya sido iniciada.

El flujo real del consultorio tiene dos fases:
  1. Agendamiento — el médico (o su asistente) reserva fecha, hora y motivo.
  2. Atención — el día de la cita, el médico abre la consulta desde la cita.

Actualmente solo existe la fase 2 (consulta), sin la fase 1.

## Decisión

### Entidad Appointment (cita)

Una cita es una reserva de tiempo, no una consulta. Vive en su propia tabla
de proyección siguiendo el patrón CQRS (ADR-0022). No usa el Core de evidencia
porque una cita no es un evento clínico inmutable — puede cancelarse, reagendarse
o completarse sin dejar rastro clínico.

  appointments:
    id              TEXT PK (UUID)
    tenant_id       TEXT NOT NULL
    patient_id      TEXT NOT NULL FK → patients.id
    scheduled_at    TIMESTAMPTZ NOT NULL        -- fecha y hora de la cita
    duration_min    INTEGER NOT NULL DEFAULT 30 -- duración estimada en minutos
    reason          TEXT NOT NULL DEFAULT ''    -- motivo (texto libre)
    state           TEXT NOT NULL DEFAULT 'scheduled'
                    -- scheduled | completed | cancelled | no_show
    consultation_id TEXT                        -- FK → consultation_projections cuando se atiende
    notes           TEXT NOT NULL DEFAULT ''    -- notas internas del médico
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()

  Índices: (tenant_id, scheduled_at), (tenant_id, patient_id), (tenant_id, state)

### Estados de una cita

  scheduled  → completed  (médico inicia consulta desde la cita)
  scheduled  → cancelled   (médico cancela la cita)
  scheduled  → no_show     (paciente no se presentó)
  completed  → (inmutable)
  cancelled  → (inmutable)
  no_show    → (inmutable)

### Relación con Consulta (ADR-0024)

  Cuando el médico inicia la consulta desde una cita:
    - Se crea la consulta (draft) normalmente.
    - appointment.consultation_id = nueva_consulta.id
    - appointment.state = "completed"

  La consulta NO requiere una cita previa — el flujo directo sigue funcionando.

### API endpoints

  POST   /api/v1/appointments                    -- crear cita
  GET    /api/v1/appointments                    -- listar citas del tenant (filtros: date, state)
  GET    /api/v1/appointments/today              -- citas del día (alimenta el widget)
  GET    /api/v1/patients/:id/appointments       -- citas de un paciente específico
  PATCH  /api/v1/appointments/:id/state          -- cancelar / no_show
  POST   /api/v1/appointments/:id/start          -- inicia consulta desde la cita

### Widget de agenda (integración)

  El endpoint /api/v1/appointments/today reemplaza o complementa el endpoint
  actual de consultas del día — incluye tanto citas agendadas como consultas
  sin cita previa del mismo día.

### UX — dos puntos de entrada

  1. Sección "Citas" en el sidebar (igual que Recetas):
     - Lista de citas del tenant ordenadas por fecha/hora.
     - Filtro por estado (todas / hoy / próximas / canceladas).
     - Botón "Nueva cita" — selecciona paciente por nombre (búsqueda).
     - Cada fila: nombre del paciente, fecha/hora, motivo, estado, acciones.
     - Acción "Iniciar consulta" en citas scheduled del día.
     - Acción "Cancelar" / "No se presentó".

  2. Botón "Agendar cita" dentro del perfil del paciente (PatientDetailView):
     - Abre el mismo formulario de nueva cita con el paciente pre-seleccionado.
     - Las citas del paciente aparecen en su perfil (tab o sección).

### Formulario de nueva cita

  - Paciente (búsqueda por nombre — pre-llenado si viene del perfil)
  - Fecha y hora
  - Duración (15 / 30 / 45 / 60 min — selector)
  - Motivo (texto libre, opcional)
  - Notas internas (texto libre, opcional)

### Validaciones (en el handler, no en Shader — citas no son evidencia clínica)

  - scheduled_at debe ser en el futuro al crear.
  - patient_id debe pertenecer al tenant.
  - duration_min entre 5 y 480.

## Dependencias

  - ADR-0022: CQRS — appointments sigue el mismo patrón de proyección.
  - ADR-0024: consulta — la cita puede generar una consulta al iniciarse.
  - Issues #163/#164: el widget de agenda se alimentará de appointments/today.

## Consecuencias

  El médico puede gestionar su agenda antes de que lleguen los pacientes.
  El widget de hoy mostrará citas reales con estado y nombre del paciente.
  El flujo de consulta directa (sin cita previa) no cambia.
  Las citas no son evidencia clínica — no pasan por el Core ni el Shader.
