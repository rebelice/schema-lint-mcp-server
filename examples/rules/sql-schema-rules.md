# SQL Schema Lint Rules

## Rule: table-naming-convention
- **Severity**: error
- **Target**: tables.*
- **Check**: name_pattern("^[a-z][a-z0-9_]*$")

Table names must be lowercase with underscores (snake_case).

### Good Example
```sql
CREATE TABLE user_accounts (
    id SERIAL PRIMARY KEY
);
```

### Bad Example
```sql
CREATE TABLE UserAccounts (
    id SERIAL PRIMARY KEY
);
```

## Rule: primary-key-required
- **Severity**: error
- **Target**: tables.*
- **Check**: has_primary_key()

Every table must have a primary key defined.

## Rule: column-naming-convention
- **Severity**: warning
- **Target**: tables.*.columns.*
- **Check**: name_pattern("^[a-z][a-z0-9_]*$")

Column names should use snake_case for consistency.

## Rule: foreign-key-naming
- **Severity**: warning
- **Target**: tables.*.foreign_keys.*
- **Check**: name_pattern("^fk_[a-z]+_[a-z]+$")

Foreign key constraints should follow the pattern: fk_<table>_<referenced_table>.

## Rule: index-naming-convention
- **Severity**: warning
- **Target**: tables.*.indexes.*
- **Check**: name_pattern("^idx_[a-z]+_[a-z_]+$")

Indexes should follow the pattern: idx_<table>_<columns>.

## Rule: no-reserved-words
- **Severity**: error
- **Target**: tables.*, tables.*.columns.*
- **Check**: not_reserved_word()

Table and column names must not use SQL reserved words.

## Rule: timestamp-columns
- **Severity**: info
- **Target**: tables.*
- **Check**: has_columns(["created_at", "updated_at"])

Tables should have created_at and updated_at timestamp columns.

## Rule: varchar-length-specified
- **Severity**: error
- **Target**: tables.*.columns[type=varchar]
- **Check**: has_length_constraint()

VARCHAR columns must specify a maximum length.

## Rule: consistent-timestamp-naming
- **Severity**: warning
- **Target**: tables.*.columns[type=timestamp]
- **Check**: name_pattern("^(created_at|updated_at|deleted_at|.*_at)$")

Timestamp columns should follow consistent naming with _at suffix.