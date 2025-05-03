CREATE TABLE User_ (
  username VARCHAR(100) PRIMARY KEY,
  password_params VARCHAR(80)
);

CREATE TABLE Goal (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  start_datetime TIMESTAMP NOT NULL,
  end_date DATE,
  completed_datetime TIMESTAMP,
  notes VARCHAR(1000),
  username VARCHAR(100) REFERENCES User_(username) NOT NULL
);

CREATE TABLE SessionId (
  username VARCHAR(100) REFERENCES User_(username) NOT NULL PRIMARY KEY REFERENCES User_(username),
  session_id_sha256 BYTEA
);

CREATE INDEX idx_session_id_sha256 ON SessionId (session_id_sha256);

