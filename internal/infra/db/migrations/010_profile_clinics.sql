CREATE TABLE IF NOT EXISTS profile_clinics (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID        NOT NULL,
    profile_id            UUID        NOT NULL,
    clinic_id             UUID        NOT NULL,
    working_days          JSONB,
    slot_duration_minutes INT         NOT NULL DEFAULT 30,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ,
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX        IF NOT EXISTS idx_profile_clinics_tenant_id ON profile_clinics(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS udx_profile_clinic            ON profile_clinics(profile_id, clinic_id);
