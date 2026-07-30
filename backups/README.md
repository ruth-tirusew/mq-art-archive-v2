# Database backups

Created with `pg_dump` against local Docker Postgres (`mq-postgres`).

## Restore this backup

```bash
# From repo root — restore custom-format dump (recommended)
docker exec -i mq-postgres pg_restore \
  --clean --if-exists --no-owner --dbname=mq \
  < backups/mq-full-YYYYMMDD-HHMMSS.dump

# Or plain SQL
docker exec -i mq-postgres psql -U mq -d mq < backups/mq-full-YYYYMMDD-HHMMSS.sql
```

If `pg_restore` via stdin fails on your Docker setup, copy the file in first:

```bash
docker cp backups/mq-full-YYYYMMDD-HHMMSS.dump mq-postgres:/tmp/restore.dump
docker exec mq-postgres pg_restore --clean --if-exists --no-owner -U mq -d mq /tmp/restore.dump
```

After restore, refresh the web/admin apps.
