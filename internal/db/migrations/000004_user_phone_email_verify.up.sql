ALTER TABLE users
    ADD COLUMN phone TEXT,
    ADD COLUMN phone_verified_at TIMESTAMPTZ;
