-- Sellers choose the contact details exposed on their published listing.
-- The listings service owns these fields so it does not need auth-db access.
ALTER TABLE listings
    ADD COLUMN seller_name text,
    ADD COLUMN contact_phone text,
    ADD COLUMN contact_email text;

ALTER TABLE listings
    ADD CONSTRAINT listings_seller_name_length
        CHECK (seller_name IS NULL OR char_length(seller_name) BETWEEN 2 AND 80),
    ADD CONSTRAINT listings_contact_phone_length
        CHECK (contact_phone IS NULL OR char_length(contact_phone) <= 40),
    ADD CONSTRAINT listings_contact_email_length
        CHECK (contact_email IS NULL OR char_length(contact_email) <= 254);
