-- +goose Up
CREATE TABLE
  short_links (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMPTZ
  );

CREATE TABLE
  clicks (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) NOT NULL,
    ip INET,
    user_agent TEXT,
    referrer TEXT,
    clicked_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );

CREATE INDEX idx_link_clicked ON clicks (short_code, clicked_at);

-- +goose Down
DROP INDEX IF EXISTS idx_link_clicked;

DROP TABLE short_links;

DROP TABLE clicks;