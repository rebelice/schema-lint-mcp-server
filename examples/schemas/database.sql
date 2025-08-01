-- Example SQL schema for an e-commerce database with intentional issues
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    firstName VARCHAR(100) NOT NULL,
    lastName VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Missing index on frequently queried columns
    phone VARCHAR(20),
    address TEXT
);

CREATE TABLE products (
    id SERIAL, -- Missing PRIMARY KEY
    Name VARCHAR(255) NOT NULL, -- Inconsistent naming (PascalCase)
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL, -- Status without CHECK constraint
    total_amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
    -- Missing index on user_id foreign key
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT order_items_products_fk FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL, -- VARCHAR without length specification
    parent_id INTEGER,
    FOREIGN KEY (parent_id) REFERENCES categories(id)
);

CREATE INDEX user_email_idx ON users(email);
CREATE UNIQUE INDEX unique_product_name ON products(Name);

ALTER TABLE products ADD CONSTRAINT products_categories_fk 
    FOREIGN KEY (category_id) REFERENCES categories(id);

CREATE VIEW active_products AS
SELECT p.*, c.name as category_name
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.stock_quantity > 0;

-- Additional tables with more violations

-- Table with poor indexing strategy
CREATE TABLE user_activity_log (
    log_id SERIAL PRIMARY KEY,
    user_id INTEGER,  -- No foreign key constraint
    action_type VARCHAR,  -- No length specified
    action_timestamp TIMESTAMP,
    ip_address VARCHAR(15),  -- Too short for IPv6
    user_agent TEXT,
    response_time FLOAT  -- Using FLOAT for time measurements
);

-- Table with denormalized data
CREATE TABLE order_summary (
    summary_id SERIAL PRIMARY KEY,
    order_id INTEGER,
    user_name VARCHAR(200),  -- Denormalized from users table
    user_email VARCHAR(255),  -- Denormalized from users table
    product_names TEXT,  -- Storing multiple values in single column
    total_items TEXT,  -- Should be INTEGER
    order_date DATE,  -- Should be TIMESTAMP
    delivery_status VARCHAR(50)
);

-- Table with missing constraints
CREATE TABLE inventory_movements (
    movement_id SERIAL,  -- PRIMARY KEY missing
    product_id INTEGER,
    quantity INTEGER,  -- No CHECK constraint for positive values
    movement_type VARCHAR(20),  -- No CHECK constraint for valid types
    movement_date TIMESTAMP,
    notes TEXT
);

-- Table with composite key issues
CREATE TABLE user_settings (
    user_id INTEGER,
    setting_name VARCHAR(100),
    setting_value TEXT,
    created_at TIMESTAMP,
    -- Missing PRIMARY KEY definition for composite key
    updated_at TIMESTAMP
);