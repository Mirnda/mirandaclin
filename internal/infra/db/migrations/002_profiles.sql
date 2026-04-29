CREATE TABLE IF NOT EXISTS profiles (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL,
    tenant_id            UUID NOT NULL,
    full_name            TEXT,
    document             TEXT,
    birth_date           DATE,
    address_postal_code  TEXT,
    address_street       TEXT,
    address_number       TEXT,
    address_complement   TEXT,
    address_neighborhood TEXT,
    address_city         TEXT,
    address_state        TEXT,
    address_country      TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_user_id   ON profiles(user_id);
CREATE INDEX        IF NOT EXISTS idx_profiles_tenant_id ON profiles(tenant_id);
