CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    role TEXT NOT NULL,
    auth_provider TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash TEXT NOT NULL,
    secrets_version INTEGER NOT NULL,
    totp_secret TEXT NOT NULL DEFAULT '',
    -- refresh_token INTEGER REFERENCES refresh_tokens(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_name UNIQUE (name),
    CONSTRAINT unique_email UNIQUE (email)
);