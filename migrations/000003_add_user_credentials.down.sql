ALTER TABLE users
    DROP CONSTRAINT users_email_unique,
    DROP COLUMN password_hash,
    DROP COLUMN email;