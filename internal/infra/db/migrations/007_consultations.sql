CREATE TABLE IF NOT EXISTS consultations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    appointment_id UUID NOT NULL,
    patient_id     UUID NOT NULL,
    dentist_id     UUID NOT NULL,
    diagnosis      TEXT,
    treatment      TEXT,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consultations_tenant_id  ON consultations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_consultations_patient_id ON consultations(patient_id);
CREATE INDEX IF NOT EXISTS idx_consultations_dentist_id ON consultations(dentist_id);
