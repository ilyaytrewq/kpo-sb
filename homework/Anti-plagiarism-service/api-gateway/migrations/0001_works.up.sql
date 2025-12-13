CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS works (
  work_id     uuid PRIMARY KEY,
  name       text NOT NULL,
  description text,
  created_at  timestamptz NOT NULL DEFAULT now()
);
