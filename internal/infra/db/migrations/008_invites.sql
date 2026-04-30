CREATE TABLE IF NOT EXISTS invites (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID        NOT NULL,
    token         TEXT        NOT NULL,
    email         TEXT        NOT NULL DEFAULT '',
    role          TEXT        NOT NULL DEFAULT '',
    password_hash TEXT        NOT NULL DEFAULT '',
    salt          TEXT        NOT NULL DEFAULT '',
    event_id      TEXT,
    used_at       TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS udx_invites_token  ON invites(token);
CREATE        INDEX IF NOT EXISTS idx_invites_tenant ON invites(tenant_id);
