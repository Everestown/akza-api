-- +goose Up
-- +goose StatementBegin

-- Default admin: admin@akza.ru / akza2024
-- bcrypt hash generated with cost 12
INSERT INTO admins (email, password_hash, name)
VALUES (
    'admin@akza.ru',
    '$2a$12$P8ulaorndXZAOQqRFXGF/uEzTk5XVbFQ.D3q3RvyGvNDRuFImdzL.',
    'AKZA Admin'
) ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM admins WHERE email = 'admin@akza.ru';
-- +goose StatementEnd
