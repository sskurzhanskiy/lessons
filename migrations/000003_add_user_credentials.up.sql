ALTER TABLE users
    ADD COLUMN email TEXT,
    ADD COLUMN password_hash TEXT,
    ADD CONSTRAINT users_email_unique UNIQUE (email);