CREATE TABLE IF NOT EXISTS tasks (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  done boolean
);
