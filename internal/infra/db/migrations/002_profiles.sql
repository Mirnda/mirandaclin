CREATE TABLE IF NOT EXISTS profiles (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID,
    tenant_id               UUID        NOT NULL,
    role                    TEXT        NOT NULL,
    full_name               TEXT        NOT NULL,
    document                TEXT,
    birth_date              DATE,
    phone                   TEXT,
    has_whatsapp            BOOLEAN     NOT NULL DEFAULT false,
    emergency_contact_name  TEXT,
    emergency_contact_phone TEXT,
    address_postal_code     TEXT,
    address_street          TEXT,
    address_number          TEXT,
    address_complement      TEXT,
    address_neighborhood    TEXT,
    address_city            TEXT,
    address_state           TEXT,
    address_country         TEXT,
    address_latitude        TEXT,
    address_longitude       TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ,
    deleted_at              TIMESTAMPTZ
);

-- Staff (user_id NOT NULL) não pode ter perfil duplicado no mesmo tenant
CREATE UNIQUE INDEX IF NOT EXISTS udx_profiles_user_tenant
    ON profiles(user_id, tenant_id)
    WHERE user_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_profiles_tenant_id ON profiles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_profiles_role      ON profiles(tenant_id, role);
CREATE INDEX IF NOT EXISTS idx_profiles_deleted   ON profiles(deleted_at);

