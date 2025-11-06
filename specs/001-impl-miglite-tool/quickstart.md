# Quickstart Guide: MigLite Tool

## Prerequisites
- Go 1.19 or higher
- A supported database (MySQL, PostgreSQL, or SQLite3)
- Database driver dependency added to your project (e.g., github.com/go-sql-driver/mysql, github.com/lib/pq, github.com/mattn/go-sqlite3)

## Installation
```bash
# Clone the repository
git clone <your-repo-url>
cd miglite

# Build the tool
go build -o miglite main.go

# Alternatively, build and install to GOPATH
go install .
```

## Configuration
Create a `miglite.yaml` file in your project root:

```yaml
database:
  driver: mysql  # or postgresql, sqlite3
  dsn: root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local
migrations:
  path: ./migrations
```

Or set the `DATABASE_URL` environment variable:
```bash
export DATABASE_URL=mysql://user:password@localhost:3306/mydb
```

## Creating Your First Migration
```bash
# Create a new migration file
miglite create add-users-table
```

This generates a file in `./migrations/` with the current timestamp: `YYYYMMDD-add-users-table.sql`

Edit the file to define your migration:
```sql
-- Migrate:UP
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);

-- Migrate:DOWN
DROP TABLE users;
```

## Running Migrations
```bash
# Apply pending migrations
miglite up

# Rollback the last migration
miglite down

# Check migration status
miglite status
```

## Example Workflow
1. Create your migration: `miglite create add-users-table`
2. Edit the generated file with your SQL
3. Check status: `miglite status`
4. Apply: `miglite up`
5. Verify: `miglite status`

## Troubleshooting
- If you get "unknown driver" error, make sure you've added the appropriate database driver dependency
- If migrations fail, check that your SQL syntax is correct for your database
- If you get permission errors, verify your database credentials in the config file