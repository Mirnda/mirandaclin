CREATE TABLE IF NOT EXISTS email_verifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    token      TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS udx_email_verif_token   ON email_verifications(token);
CREATE        INDEX IF NOT EXISTS idx_email_verif_user_id ON email_verifications(user_id);
