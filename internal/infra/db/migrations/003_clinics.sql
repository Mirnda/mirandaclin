CREATE TABLE IF NOT EXISTS clinics (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL,
    name                 TEXT NOT NULL,
    phone                TEXT,
    address_postal_code  TEXT,
    address_street       TEXT,
    address_number       TEXT,
    address_complement   TEXT,
    address_neighborhood TEXT,
    address_city         TEXT,
    address_state        TEXT,
    address_country      TEXT,
    address_latitude     TEXT,
    address_longitude    TEXT,
    operating_days       TEXT[],
    open_time            TEXT,
    close_time           TEXT,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ,
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_clinics_tenant_id  ON clinics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_clinics_deleted_at ON clinics(deleted_at);
