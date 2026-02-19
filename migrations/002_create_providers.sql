-- +goose Up
CREATE TABLE providers (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    access_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, type),
    UNIQUE (type, provider_user_id)
);

CREATE INDEX idx_providers_user_id ON providers(user_id);
CREATE INDEX idx_providers_type_provider_user_id ON providers(type, provider_user_id);

-- +goose Down
DROP TABLE providers;
