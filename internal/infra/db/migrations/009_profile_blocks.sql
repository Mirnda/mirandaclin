CREATE TABLE IF NOT EXISTS profile_blocks (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID        NOT NULL,
    profile_id   UUID        NOT NULL,
    clinic_id    UUID,
    blocked_date DATE        NOT NULL,
    start_time   TEXT,
    end_time     TEXT,
    reason       TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    canceled_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_profile_blocks_tenant_id  ON profile_blocks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_profile_blocks_profile_id ON profile_blocks(profile_id);
