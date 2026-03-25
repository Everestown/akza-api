-- +goose Up
-- +goose StatementBegin

-- Add DICTIONARY to page_section enum
ALTER TYPE page_section ADD VALUE IF NOT EXISTS 'DICTIONARY';

-- +goose StatementEnd
-- +goose Down