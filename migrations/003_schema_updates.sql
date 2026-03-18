-- +goose Up
-- +goose StatementBegin

-- admins: expand name from 50 to 100
ALTER TABLE admins ALTER COLUMN name TYPE VARCHAR(100);

-- collections: tighten description to VARCHAR(150) (was TEXT)
ALTER TABLE collections ALTER COLUMN description TYPE VARCHAR(150);

-- products: add price_hidden flag
ALTER TABLE products ADD COLUMN IF NOT EXISTS price_hidden BOOLEAN NOT NULL DEFAULT FALSE;

-- orders: expand customer_name from 50 to 100
ALTER TABLE orders ALTER COLUMN customer_name TYPE VARCHAR(100);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE admins ALTER COLUMN name TYPE VARCHAR(50);
ALTER TABLE collections ALTER COLUMN description TYPE TEXT;
ALTER TABLE products DROP COLUMN IF EXISTS price_hidden;
ALTER TABLE orders ALTER COLUMN customer_name TYPE VARCHAR(50);
-- +goose StatementEnd
