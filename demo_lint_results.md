# SQL Schema Lint Results

## Summary
- **Schema File**: ./examples/schemas/database.sql
- **Rules File**: ./examples/rules/sql-schema-rules.md
- **Dialect**: PostgreSQL
- **Total Issues**: 11
- **Errors**: 3
- **Warnings**: 6
- **Info**: 2

## Issues Found

### ❌ ERRORS (Must Fix)

#### 1. Missing Primary Key
- **Rule**: primary-key-required
- **Location**: Line 12, Table `products`
- **Issue**: Table `products` does not have a PRIMARY KEY constraint
- **Fix**: Add PRIMARY KEY constraint to the `id` column
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,  -- Add PRIMARY KEY here
    ...
);
```

#### 2. VARCHAR Without Length
- **Rule**: varchar-length-specified
- **Location**: Line 45, Column `categories.name`
- **Issue**: VARCHAR column without specified maximum length
- **Fix**: Specify a maximum length for the VARCHAR column
```sql
name VARCHAR(255) NOT NULL,  -- Specify length
```

#### 3. Inconsistent Naming Convention
- **Rule**: column-naming-convention
- **Location**: Line 8, Column `users.createdAt`
- **Issue**: Column name uses camelCase instead of snake_case
- **Fix**: Rename to `created_at` for consistency

### ⚠️ WARNINGS (Should Fix)

#### 4. Column Naming Convention
- **Rule**: column-naming-convention
- **Location**: Line 4, Column `users.firstName`
- **Issue**: Should be `first_name` (snake_case)

#### 5. Column Naming Convention
- **Rule**: column-naming-convention
- **Location**: Line 5, Column `users.lastName`
- **Issue**: Should be `last_name` (snake_case)

#### 6. Column Naming Convention
- **Rule**: column-naming-convention
- **Location**: Line 14, Column `products.Name`
- **Issue**: Should be `name` (lowercase)

#### 7. Foreign Key Naming
- **Rule**: foreign-key-naming
- **Location**: Line 40, Constraint `order_items_products_fk`
- **Issue**: Should follow pattern `fk_order_items_products`

#### 8. Foreign Key Naming
- **Rule**: foreign-key-naming
- **Location**: Line 53, Constraint `products_categories_fk`
- **Issue**: Should follow pattern `fk_products_categories`

#### 9. Index Naming Convention
- **Rule**: index-naming-convention
- **Location**: Line 50, Index `user_email_idx`
- **Issue**: Should be `idx_users_email`

#### 10. Index Naming Convention
- **Rule**: index-naming-convention
- **Location**: Line 51, Index `unique_product_name`
- **Issue**: Should be `idx_products_name` or `uidx_products_name`

### ℹ️ INFO (Consider)

#### 11. Timestamp Columns
- **Rule**: timestamp-columns
- **Location**: Multiple tables
- **Issue**: Inconsistent timestamp column naming (`createdAt` vs `created_at`)
- **Suggestion**: Standardize all timestamp columns to use snake_case

## Recommendations

1. **Fix all ERROR-level issues first** - These are critical for schema integrity
2. **Address WARNING-level issues** - These improve consistency and maintainability
3. **Consider INFO-level suggestions** - These enhance overall schema quality

## Next Steps

1. Add PRIMARY KEY to the `products` table
2. Specify length for all VARCHAR columns without length constraints
3. Rename columns to follow snake_case convention consistently
4. Update foreign key and index names to follow naming patterns