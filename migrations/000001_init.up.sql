CREATE TABLE IF NOT EXISTS users(
  id TEXT PRIMARY KEY,
  username TEXT,
  is_active BOOLEAN,
  team TEXT
);

CREATE TABLE IF NOT EXISTS pullrequests(
  id TEXT PRIMARY KEY,
  name TEXT,
  author_id TEXT,
  status TEXT DEFAULT 'OPEN',
  reviewers_id TEXT[],
  merged_at timestamp,

  CONSTRAINT fk_author_id
    FOREIGN KEY (author_id)
    REFERENCES users(id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION
);
