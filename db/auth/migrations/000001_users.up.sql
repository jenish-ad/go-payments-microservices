BEGIN;

-- citext compares text without regard to case, so Alice@Example.com and
-- alice@example.com are the same email and the unique constraint knows it.

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    id          uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
    email       citext          NOT NULL UNIQUE CHECK (length(email) <= 254),
    -- Argon2id in the PHC string format stores the parameters, salt, and hash in one string.
    -- Never the password itself.
    password_hash   text        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);

COMMIT;
