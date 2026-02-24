ALTER TABLE users
    ADD COLUMN IF NOT EXISTS name TEXT;

UPDATE users
SET name = split_part(email, '@', 1)
WHERE name IS NULL OR btrim(name) = '';

ALTER TABLE users
    ALTER COLUMN name SET NOT NULL;
