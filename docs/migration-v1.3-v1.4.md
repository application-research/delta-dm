# V1.3 to V1.4 Migration

If you have an existing database from DDM V1.3, you will need to migrate it to V1.4. This migration is not backwards compatible, so you will not be able to use the V1.4 database with V1.3.

Note: This migration is only necessary as we did not have the automatic migration framework in place prior to this release. Thus, we are using manual SQL to ensure database consistency.
Going forward, this will not be required as the schema established in V1.4 allows for automatic migrations. 

## Manual SQL
Please execute the following SQL statements in your DDM database.

### Connect: SQLite (default)
- Connect to the DB using `sqlite3`,
For instance,

```bash
sqlite3 delta-dm.db
```

### Connect: Postgres
- Connect to the DB using `psql`,
For instance,

```bash
psql postgres://postgres:password@localhost:5432/delta-dm
```

Next, run the following SQL statements:

1. Populate `dataset_id` in `contents` table
```sql
-- Create the new table
CREATE TABLE contents_new (
    comm_p TEXT PRIMARY KEY,
    payload_c_id TEXT,
    size INTEGER NOT NULL,
    padded_size INTEGER NOT NULL,
    dataset_id INTEGER NOT NULL,
    num_replications INTEGER NOT NULL,
    dataset_name TEXT,
    FOREIGN KEY (dataset_id) REFERENCES datasets(id) ON DELETE CASCADE
);

-- Copy the data from the old table
INSERT INTO contents_new (comm_p, payload_c_id, size, padded_size, dataset_id, num_replications, dataset_name)
SELECT comm_p, payload_c_id, size, padded_size, 0, num_replications, dataset_name FROM contents;

-- Set the dataset_id column appropriately
UPDATE contents_new
SET dataset_id = (
    SELECT id FROM datasets WHERE name = contents_new.dataset_name
);

-- Drop the dataset_name column
ALTER TABLE contents_new DROP COLUMN dataset_name;

-- Drop the old table
DROP TABLE contents;

-- Rename the new table to the original name
ALTER TABLE contents_new RENAME TO contents;
```

2. Move `is_self_service` to `ss_is_self_service` in `replications` table
```sql
UPDATE replications SET ss_is_self_service = is_self_service;
ALTER TABLE replications DROP COLUMN is_self_service;
```