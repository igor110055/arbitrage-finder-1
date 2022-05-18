CREATE OR REPLACE FUNCTION trigger_set_timestamp()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE account_statuses_enum AS ENUM ('created', 'active', 'banned');

CREATE TABLE IF NOT EXISTS accounts
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    phone      VARCHAR(64) NOT NULL UNIQUE,
    password   VARCHAR(150) NOT NULL,
    status     account_statuses_enum NOT NULL DEFAULT 'created'
);

CREATE INDEX accounts__phone_idx ON accounts (phone);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON accounts
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TABLE IF NOT EXISTS phone_confirmations_challenges
(
    id                 BIGSERIAL PRIMARY KEY,
    created_at         TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP NOT NULL DEFAULT NOW(),

    account_id         BIGINT NOT NULL,
    phone              VARCHAR(64) NOT NULL UNIQUE,
    code               VARCHAR(4) NOT NULL,
    remaining_attempts INTEGER NOT NULL ,
    used               BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_phone_confirmations_challenges__account_id
        FOREIGN KEY (account_id)
            REFERENCES accounts (id) ON DELETE CASCADE
);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON phone_confirmations_challenges
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TABLE refresh_tokens (
    id         BYTEA,
    account_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_refresh_tokens__account_id
        FOREIGN KEY (account_id)
            REFERENCES accounts (id) ON DELETE CASCADE
);