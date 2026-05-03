CREATE TABLE IF NOT EXISTS tenant_members (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    tenant_id  UUID        NOT NULL,
    role       TEXT        NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS udx_member         ON tenant_members(user_id, tenant_id);
CREATE        INDEX IF NOT EXISTS idx_member_user_id ON tenant_members(user_id);
CREATE        INDEX IF NOT EXISTS idx_member_tenant  ON tenant_members(tenant_id);
