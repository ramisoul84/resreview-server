#!/bin/bash
set -e

[ -f .env ] && export $(grep -v '^#' .env.dev | xargs)

echo "Running migrations UP...$DB_NAME ..."

for f in migrations/*.up.sql; do
  echo "  → $f"
  docker exec -i postgres psql -U "$DB_USER" -d "$DB_NAME" < "$f"
done

echo "Done."