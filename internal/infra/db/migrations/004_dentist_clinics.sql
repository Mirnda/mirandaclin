CREATE TABLE IF NOT EXISTS dentist_clinics (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID NOT NULL,
    dentist_id            UUID NOT NULL,
    clinic_id             UUID NOT NULL,
    working_days          TEXT[],
    start_time            TEXT,
    end_time              TEXT,
    slot_duration_minutes INT NOT NULL DEFAULT 30,
    active                BOOLEAN NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ
);

CREATE INDEX        IF NOT EXISTS idx_dentist_clinics_tenant_id ON dentist_clinics(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS udx_dentist_clinic            ON dentist_clinics(dentist_id, clinic_id);
