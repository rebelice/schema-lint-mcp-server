# Advanced SQL Schema Lint Rules

## Rule: cascade-delete-restriction
- **Severity**: warning
- **Target**: tables.*.foreign_keys[on_delete=CASCADE]
- **Check**: has_comment()

Foreign keys with CASCADE DELETE must have a comment explaining the impact.

### Good Example
```sql
-- Deleting an order will automatically delete all associated order items
FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
```

### Bad Example
```sql
FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
```

## Rule: unique-constraint-naming
- **Severity**: warning
- **Target**: tables.*.constraints[type=unique]
- **Check**: name_pattern("^uq_[a-z]+_[a-z_]+$")

Unique constraints should follow the pattern: uq_<table>_<columns>.

## Rule: partition-key-index
- **Severity**: error
- **Target**: tables[partitioned=true]
- **Check**: partition_key_indexed()

Partitioned tables must have an index on the partition key.

## Rule: enum-check-constraint
- **Severity**: warning
- **Target**: tables.*.columns[has_check_constraint]
- **Check**: check_constraint_values_documented()

Columns with CHECK constraints for enum-like values should be documented.

## Rule: avoid-nullable-foreign-keys
- **Severity**: warning
- **Target**: tables.*.columns[is_foreign_key=true]
- **Check**: not_nullable()

Foreign key columns should generally not be nullable to maintain referential integrity.

## Rule: index-on-foreign-keys
- **Severity**: info
- **Target**: tables.*.foreign_keys.*
- **Check**: has_index()

Foreign key columns should have indexes for better query performance.

## Rule: boolean-column-naming
- **Severity**: warning
- **Target**: tables.*.columns[type=boolean]
- **Check**: name_pattern("^(is_|has_|can_|should_)[a-z_]+$")

Boolean columns should have descriptive names starting with is_, has_, can_, or should_.

## Rule: decimal-precision-specified
- **Severity**: error
- **Target**: tables.*.columns[type=decimal]
- **Check**: has_precision_and_scale()

DECIMAL columns must specify both precision and scale.

## Rule: avoid-float-for-money
- **Severity**: error
- **Target**: tables.*.columns[name_pattern=".*price.*|.*amount.*|.*cost.*"]
- **Check**: type_not_in(["float", "real", "double"])

Monetary values should use DECIMAL type, not floating-point types.

## Rule: composite-index-column-order
- **Severity**: info
- **Target**: tables.*.indexes[column_count>1]
- **Check**: optimal_column_order()

Multi-column indexes should have columns ordered by selectivity (most selective first).