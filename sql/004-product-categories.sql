CREATE TABLE IF NOT EXISTS product_categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL,
    name VARCHAR(32)
);

ALTER TABLE products 
    ADD COLUMN category_id INTEGER,
    ADD CONSTRAINT fk_category
        FOREIGN KEY (category_id)
        REFERENCES product_categories(id)
        ON DELETE CASCADE;

CREATE INDEX idx_category ON products(category_id);