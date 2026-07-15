ALTER TABLE app.entries
    ADD COLUMN term_id bigint;

UPDATE app.entries e
SET term_id = t.id
FROM app.terms t
WHERE t.user_id = e.user_id
  AND t.is_active = true;

ALTER TABLE app.entries
    ALTER COLUMN term_id SET NOT NULL,
    ADD CONSTRAINT entries_term_id_fkey
        FOREIGN KEY (term_id) REFERENCES app.terms (id),
    ADD CONSTRAINT entries_term_id_date_key
        UNIQUE (term_id, date),
    DROP CONSTRAINT entries_user_id_date_key,
    DROP CONSTRAINT entries_user_id_fkey,
    DROP COLUMN user_id;
