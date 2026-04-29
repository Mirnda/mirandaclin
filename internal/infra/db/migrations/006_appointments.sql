CREATE TABLE IF NOT EXISTS appointments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    patient_id   UUID NOT NULL,
    dentist_id   UUID NOT NULL,
    clinic_id    UUID NOT NULL,
    secretary_id UUID,
    scheduled_at TIMESTAMPTZ NOT NULL,
    canceled_at  TIMESTAMPTZ,
    status       TEXT NOT NULL DEFAULT 'scheduled',
    notes        TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_appointments_tenant_id ON appointments(tenant_id);
