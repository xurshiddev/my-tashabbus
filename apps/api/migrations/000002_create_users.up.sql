CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL,
    phone TEXT NULL,
    telegram_id BIGINT UNIQUE NULL,
    telegram_username TEXT NULL,
    role TEXT NOT NULL,
    mfy_id UUID NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_role_check CHECK (
        role IN (
            'SUPER_ADMIN',
            'MFY_CHAIRMAN',
            'STREET_LEADER',
            'RESPONSIBLE_PERSON'
        )
    )
);

CREATE INDEX users_telegram_id_idx ON users (telegram_id);
CREATE INDEX users_role_idx ON users (role);
CREATE INDEX users_mfy_id_idx ON users (mfy_id);
CREATE INDEX users_is_active_idx ON users (is_active);
