-- +goose Up
-- +goose StatementBegin

-- Extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Enums
CREATE TYPE collection_status AS ENUM ('DRAFT','SCHEDULED','PUBLISHED','ARCHIVED');
CREATE TYPE order_status      AS ENUM ('NEW','CONTACTED','CONFIRMED','CANCELLED','COMPLETED');
CREATE TYPE media_type        AS ENUM ('IMAGE','VIDEO');
CREATE TYPE page_section      AS ENUM ('HERO','ABOUT','CONTACTS','FOOTER');

-- updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- admins
CREATE TABLE admins (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(50)  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- collections
CREATE TABLE collections (
    id           UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         VARCHAR(120)       NOT NULL,
    title        VARCHAR(100)       NOT NULL,
    description  TEXT,
    cover_url    TEXT,
    cover_s3_key VARCHAR(500),
    status       collection_status  NOT NULL DEFAULT 'DRAFT',
    scheduled_at TIMESTAMPTZ,
    sort_order   INTEGER            NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_collections_slug
    ON collections (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_status_sched
    ON collections (status, scheduled_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_title_trgm
    ON collections USING GIN (title gin_trgm_ops);

CREATE TRIGGER trg_collections_updated_at
    BEFORE UPDATE ON collections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- products
CREATE TABLE products (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id  UUID         NOT NULL REFERENCES collections(id) ON DELETE RESTRICT,
    slug           VARCHAR(120) NOT NULL UNIQUE,
    title          VARCHAR(100) NOT NULL,
    description    TEXT,
    characteristics JSONB       NOT NULL DEFAULT '{}',
    price          NUMERIC(12,2) NOT NULL,
    cover_url      TEXT,
    cover_s3_key   VARCHAR(500),
    sort_order     INTEGER      NOT NULL DEFAULT 0,
    is_published   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_products_collection_id
    ON products (collection_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_title_trgm
    ON products USING GIN (title gin_trgm_ops);

CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- product_variants
CREATE TABLE product_variants (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id   UUID         NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    slug         VARCHAR(150) NOT NULL UNIQUE,
    attributes   JSONB        NOT NULL DEFAULT '{}',
    is_published BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_variants_product_id
    ON product_variants (product_id, sort_order) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_variants_updated_at
    BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- variant_images
CREATE TABLE variant_images (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id UUID         NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    url        TEXT         NOT NULL,
    s3_key     VARCHAR(500) NOT NULL,
    sort_order INTEGER      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_variant_images_variant_id
    ON variant_images (variant_id, sort_order);

-- orders
CREATE TABLE orders (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id         UUID         NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    customer_name      VARCHAR(50)  NOT NULL,
    telegram_username  VARCHAR(100) NOT NULL,
    phone              VARCHAR(20),
    comment            VARCHAR(150),
    status             order_status NOT NULL DEFAULT 'NEW',
    tg_notified_at     TIMESTAMPTZ,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_status      ON orders (status, created_at DESC);
CREATE INDEX idx_orders_variant_id  ON orders (variant_id);

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- pg_notify for new orders
CREATE OR REPLACE FUNCTION notify_new_order()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_order', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_notify_new_order
    AFTER INSERT ON orders
    FOR EACH ROW EXECUTE FUNCTION notify_new_order();

-- media_assets
CREATE TABLE media_assets (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    s3_key        VARCHAR(500) NOT NULL UNIQUE,
    url           TEXT         NOT NULL,
    type          media_type   NOT NULL,
    original_name VARCHAR(255),
    size_bytes    BIGINT,
    mime_type     VARCHAR(100),
    uploaded_by   UUID         REFERENCES admins(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_assets_type ON media_assets (type, created_at DESC);

-- site_pages
CREATE TABLE site_pages (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    section    page_section NOT NULL UNIQUE,
    content    JSONB        NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by UUID         REFERENCES admins(id) ON DELETE SET NULL
);

CREATE TRIGGER trg_site_pages_updated_at
    BEFORE UPDATE ON site_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Seed default site pages
INSERT INTO site_pages (section, content) VALUES
    ('HERO',     '{"title":"AKZA","subtitle":"Видеть искусство в каждом стежке","video_url":null}'),
    ('ABOUT',    '{"text":"Fashion","instagram":"@the.akza"}'),
    ('CONTACTS', '{"telegram":"t.me/theakza","instagram":"@the.akza","address":"Махачкала, ул. Толстого 5/1"}'),
    ('FOOTER',   '{"copyright":"© AKZA 2026","links":[]}');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS site_pages;
DROP TABLE IF EXISTS media_assets;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS variant_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS admins;
DROP FUNCTION IF EXISTS notify_new_order();
DROP FUNCTION IF EXISTS update_updated_at();
DROP TYPE IF EXISTS page_section;
DROP TYPE IF EXISTS media_type;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS collection_status;
DROP EXTENSION IF EXISTS btree_gist;
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS pgcrypto;
-- +goose StatementEnd
