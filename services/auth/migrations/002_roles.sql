-- Registered accounts start as buyers. Seller status is granted only through
-- the authenticated, CSRF-protected become-seller endpoint.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'buyer'
    CHECK (role IN ('buyer', 'seller'));

DROP INDEX IF EXISTS idx_users_email;
