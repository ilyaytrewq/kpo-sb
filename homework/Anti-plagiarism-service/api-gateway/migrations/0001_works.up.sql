CREATE TABLE IF NOT EXISTS works (
  work_id     text PRIMARY KEY,
  name       text NOT NULL,
  description text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now()
);
