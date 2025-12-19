CREATE TABLE IF NOT EXISTS submissions (
  submission_id text PRIMARY KEY,
  work_id text NOT NULL REFERENCES works(work_id) ON DELETE CASCADE,
  file_id uuid NOT NULL,
  uploaded_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS submissions_work_id_idx ON submissions (work_id);
