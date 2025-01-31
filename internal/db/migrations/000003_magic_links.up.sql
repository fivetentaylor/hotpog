CREATE TABLE magic_links (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_at TIMESTAMPTZ -- null until used
);

CREATE INDEX idx_magic_links_user_id ON magic_links(user_id);
CREATE INDEX idx_magic_links_expires_at ON magic_links(expires_at);
