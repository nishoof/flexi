ALTER TABLE app.terms
    ADD COLUMN start_date date,
    ADD COLUMN starting_amount double precision;

UPDATE app.terms
SET start_date = '2026-01-26'::date,
    starting_amount = 3010
WHERE start_date IS NULL
   OR starting_amount IS NULL;

ALTER TABLE app.terms
    ALTER COLUMN start_date SET NOT NULL,
    ALTER COLUMN starting_amount SET NOT NULL;

ALTER TABLE app.terms
    ADD CONSTRAINT terms_start_before_end CHECK (start_date <= end_date),
    ADD CONSTRAINT terms_starting_amount_nonnegative CHECK (starting_amount >= 0);
