ALTER TABLE public.flex_entries
  ADD CONSTRAINT flex_entries_user_id_date_key UNIQUE (user_id, date);
