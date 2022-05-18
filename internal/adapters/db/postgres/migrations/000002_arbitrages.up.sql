CREATE TABLE IF NOT EXISTS arbitrages
(
    pair          VARCHAR(64) NOT NULL PRIMARY KEY,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    buy_exchange  VARCHAR(150) NOT NULL,
    sell_exchange VARCHAR(150) NOT NULL,
    buy_price     DECIMAL NOT NULL,
    sell_price    DECIMAL NOT NULL,
    profit        DECIMAL NOT NULL
);

CREATE INDEX arbitrages__pair_idx ON arbitrages (pair);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON arbitrages
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
