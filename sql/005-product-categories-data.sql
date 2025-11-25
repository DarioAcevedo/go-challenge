-- Generate product categories --
INSERT INTO product_categories (code, name) VALUES
('CAT01', 'Clothing'),
('CAT02', 'Shoes'),
('CAT03', 'Accessories');


-- Link products to product categories
UPDATE products SET category_id = (SELECT id FROM product_categories WHERE code = 'CAT01') WHERE code IN ('PROD001', 'PROD004', 'PROD007');
UPDATE products SET category_id = (SELECT id FROM product_categories WHERE code = 'CAT02') WHERE code IN ('PROD002', 'PROD006');
UPDATE products SET category_id = (SELECT id FROM product_categories WHERE code = 'CAT03') WHERE code IN ('PROD003', 'PROD005', 'PROD008');