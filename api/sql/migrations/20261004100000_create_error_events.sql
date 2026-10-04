CREATE TABLE error_events (
    fingerprint text PRIMARY KEY CHECK (length(fingerprint) = 64),
    service varchar(64) NOT NULL,
    environment varchar(64) NOT NULL,
    status integer NOT NULL CHECK (status BETWEEN 500 AND 599),
    method varchar(16) NOT NULL,
    url text NOT NULL,
    path text NOT NULL,
    error_type varchar(128) NOT NULL DEFAULT '',
    message text NOT NULL DEFAULT '',
    stack text NOT NULL DEFAULT '',
    commit_sha varchar(64) NOT NULL DEFAULT '',
    cf_ray varchar(128) NOT NULL DEFAULT '',
    user_agent varchar(1024) NOT NULL DEFAULT '',
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    occurrences bigint NOT NULL DEFAULT 1 CHECK (occurrences > 0)
);

CREATE INDEX error_events_last_seen_at_idx ON error_events (last_seen_at DESC);
