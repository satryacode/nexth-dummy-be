-- db/schema.sql
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(255),
    email      VARCHAR(255),
    password   VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO users (username, email, password)
VALUES ('admin', 'admin@dummy.com', 'admin123')
ON CONFLICT DO NOTHING;
