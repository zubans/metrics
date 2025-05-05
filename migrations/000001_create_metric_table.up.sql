CREATE TABLE metrics
(
    id        SERIAL PRIMARY KEY,
    name      TEXT      NOT NULL,
    type      TEXT      NOT NULL CHECK (type IN ('gauge', 'counter')),
    value     DOUBLE PRECISION,
    delta     BIGINT,
    timestamp TIMESTAMP NOT NULL,
    CONSTRAINT metrics_name_type_unique UNIQUE (name, type)
);