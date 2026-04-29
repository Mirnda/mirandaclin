CREATE TABLE IF NOT EXISTS users (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID NOT NULL,
    email                   TEXT NOT NULL,
    password_hash           TEXT,
    salt                    TEXT,
    role                    TEXT NOT NULL,
    phone                   TEXT,
    has_whatsapp            BOOLEAN NOT NULL DEFAULT false,
    emergency_contact_name  TEXT,
    emergency_contact_phone TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ,
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX        IF NOT EXISTS idx_users_tenant_id  ON users(tenant_id);
CREATE INDEX        IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS udx_tenant_email     ON users(tenant_id, email);
