CREATE TABLE IF NOT EXISTS users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT        NOT NULL,
    email_verified_at TIMESTAMPTZ,
    salt              TEXT,
    password_hash     TEXT,
    last_tenant_id    UUID,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS udx_users_email  ON users(email) WHERE deleted_at IS NULL;
CREATE        INDEX IF NOT EXISTS idx_users_deleted ON users(deleted_at);
