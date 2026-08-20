CREATE TABLE appointments (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    patient_id      TEXT NOT NULL REFERENCES patients(id),
    scheduled_at    TIMESTAMPTZ NOT NULL,
    duration_min    INTEGER NOT NULL DEFAULT 30,
    reason          TEXT NOT NULL DEFAULT '',
    state           TEXT NOT NULL DEFAULT 'scheduled',
    consultation_id TEXT,
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_appointments_tenant_date   ON appointments (tenant_id, scheduled_at);
CREATE INDEX idx_appointments_tenant_patient ON appointments (tenant_id, patient_id);
CREATE INDEX idx_appointments_tenant_state  ON appointments (tenant_id, state);
