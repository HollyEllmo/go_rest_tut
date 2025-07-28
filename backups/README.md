# Database Backups

This directory contains database backups for the Go REST Tutorial project.

## Backup Files

- `ecom_testdata_*.sql` - Backups with test data (users, products, orders, inventory)
- `ecom_backup_*.sql` - Regular backups
- `ecom_*_timestamp.sql` - Custom named backups

## Usage

### Create Backup
```bash
# Create backup with timestamp
make backup

# Create backup with custom name
make backup mybackup
# or
./scripts/backup_db.sh mybackup
```

### Restore Database
```bash
# Restore from latest testdata backup
make restore-testdata

# Restore from specific file
make restore backups/ecom_testdata_20250728_114811.sql
# or
./scripts/restore_testdata.sh backups/specific_backup.sql
```

## What's Included in Test Data

- **4 Users**: John, Jane, Mike, Sarah (password: "password")
- **10 Products**: Electronics with realistic prices and stock
- **4 Orders**: Mix of completed/pending orders
- **21+ Inventory Movements**: Stock in/out/returns/restocks

## Important Notes

- Always stop the application before restoring
- Docker container `go_rest_tut_mysql` must be running
- Backups include structure + data + triggers + routines
- Test data passwords are bcrypt hashed