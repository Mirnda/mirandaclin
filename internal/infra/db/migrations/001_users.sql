CREATE TABLE IF NOT EXISTS users (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email                   TEXT        NOT NULL,
    password_hash           TEXT,
    salt                    TEXT,
    full_name               TEXT,
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
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ,
    deleted_at              TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS udx_users_email   ON users(email) WHERE deleted_at IS NULL;
CREATE        INDEX IF NOT EXISTS idx_users_deleted ON users(deleted_at);
