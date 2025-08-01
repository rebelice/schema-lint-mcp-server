-- SQL schema with multiple violations of best practices

-- Table with poor naming convention
CREATE TABLE tbl_usr_data (
    ID INT,  -- No PRIMARY KEY defined
    FName VARCHAR,  -- No length specified, inconsistent naming
    LName VARCHAR,  -- No length specified
    Email TEXT,  -- Using TEXT instead of VARCHAR for email
    Age VARCHAR(10),  -- Age stored as VARCHAR instead of INT
    CreatedDate DATE,  -- Using DATE instead of TIMESTAMP
    ModifiedDate DATE
);

-- Table with no primary key at all
CREATE TABLE user_sessions (
    session_id VARCHAR(100),
    user_id INT,
    created_at TIMESTAMP,
    expires_at TIMESTAMP
);

-- Table with poor data types and no constraints
CREATE TABLE product_inventory (
    product_code VARCHAR(50),
    quantity TEXT,  -- Quantity as TEXT instead of INTEGER
    price VARCHAR(20),  -- Price as VARCHAR instead of DECIMAL
    discount FLOAT,  -- Using FLOAT for monetary values
    is_active VARCHAR(5),  -- Boolean as VARCHAR
    last_updated DATE
);

-- Foreign keys without indexes
CREATE TABLE customer_orders (
    order_id INT PRIMARY KEY,
    customer_id INT,
    product_id INT,
    order_status TEXT,  -- No CHECK constraint
    FOREIGN KEY (customer_id) REFERENCES tbl_usr_data(ID),  -- References table with no PK
    FOREIGN KEY (product_id) REFERENCES product_inventory(product_code)  -- Type mismatch
);

-- Table with reserved keywords as column names
CREATE TABLE transactions (
    select INT PRIMARY KEY,  -- Reserved keyword
    from VARCHAR(100),  -- Reserved keyword
    to VARCHAR(100),  -- Reserved keyword
    where TEXT,  -- Reserved keyword
    order INT  -- Reserved keyword
);

-- Duplicate indexes
CREATE INDEX idx_user_email ON tbl_usr_data(Email);
CREATE INDEX idx_user_email_2 ON tbl_usr_data(Email);  -- Duplicate
CREATE INDEX idx_user_email_backup ON tbl_usr_data(Email);  -- Another duplicate

-- No indexes on foreign keys
-- customer_orders.customer_id has no index
-- customer_orders.product_id has no index

-- Table with nullable primary key components (composite key)
CREATE TABLE user_preferences (
    user_id INT,
    preference_key VARCHAR(50),
    preference_value TEXT,
    PRIMARY KEY (user_id, preference_key)  -- Components should be NOT NULL
);

-- Circular dependency
CREATE TABLE departments (
    dept_id INT PRIMARY KEY,
    dept_name VARCHAR(100),
    manager_id INT
);

CREATE TABLE employees (
    emp_id INT PRIMARY KEY,
    emp_name VARCHAR(100),
    dept_id INT,
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
);

-- Add circular reference
ALTER TABLE departments ADD FOREIGN KEY (manager_id) REFERENCES employees(emp_id);

-- Table with too many columns (anti-pattern)
CREATE TABLE kitchen_sink (
    col1 INT, col2 INT, col3 INT, col4 INT, col5 INT,
    col6 VARCHAR(10), col7 VARCHAR(10), col8 VARCHAR(10), col9 VARCHAR(10), col10 VARCHAR(10),
    col11 TEXT, col12 TEXT, col13 TEXT, col14 TEXT, col15 TEXT,
    col16 DATE, col17 DATE, col18 DATE, col19 DATE, col20 DATE,
    col21 FLOAT, col22 FLOAT, col23 FLOAT, col24 FLOAT, col25 FLOAT,
    col26 BOOLEAN, col27 BOOLEAN, col28 BOOLEAN, col29 BOOLEAN, col30 BOOLEAN
);

-- View with SELECT *
CREATE VIEW all_user_data AS
SELECT * FROM tbl_usr_data;  -- Should specify columns explicitly

-- Missing cascade options on foreign keys that might need them
CREATE TABLE user_profiles (
    profile_id INT PRIMARY KEY,
    user_id INT NOT NULL,
    bio TEXT,
    FOREIGN KEY (user_id) REFERENCES tbl_usr_data(ID)  -- No ON DELETE action
);