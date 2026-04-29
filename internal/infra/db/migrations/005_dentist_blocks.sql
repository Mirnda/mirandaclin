CREATE TABLE IF NOT EXISTS dentist_blocks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    dentist_id   UUID NOT NULL,
    clinic_id    UUID,
    blocked_date DATE NOT NULL,
    start_time   TEXT,
    end_time     TEXT,
    reason       TEXT,
    created_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_dentist_blocks_tenant_id  ON dentist_blocks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dentist_blocks_dentist_id ON dentist_blocks(dentist_id);
