-- +goose Up
-- +goose StatementBegin

-- Add display name to product_variants
ALTER TABLE product_variants
  ADD COLUMN IF NOT EXISTS name VARCHAR(20) NOT NULL DEFAULT '';

-- Add media_type to variant_images (IMAGE or VIDEO)
ALTER TABLE variant_images
  ADD COLUMN IF NOT EXISTS media_type VARCHAR(10) NOT NULL DEFAULT 'IMAGE';

-- Add check constraint to enforce valid values
ALTER TABLE variant_images
  ADD CONSTRAINT chk_variant_images_media_type
  CHECK (media_type IN ('IMAGE', 'VIDEO'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE product_variants DROP COLUMN IF EXISTS name;
ALTER TABLE variant_images   DROP COLUMN IF EXISTS media_type;
ALTER TABLE variant_images   DROP CONSTRAINT IF EXISTS chk_variant_images_media_type;
-- +goose StatementEnd
