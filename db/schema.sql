-- db/schema.sql
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(255),
    email      VARCHAR(255),
    password   VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    blocked    SMALLINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS fraud_verdicts (
    id                           SERIAL PRIMARY KEY,
    source_ip                    VARCHAR(45),
    user_identity                VARCHAR(255),
    method                       VARCHAR(10),
    path                         VARCHAR(500),
    confidence_score             DECIMAL(4,2),
    reason                       VARCHAR(500),
    original_log_entry_reference TEXT,
    detected_at                  TIMESTAMP DEFAULT NOW(),
    remediated                   SMALLINT DEFAULT 0
);

INSERT INTO users (username, email, password)
VALUES ('admin', 'admin@dummy.com', 'admin123')
ON CONFLICT DO NOTHING;
