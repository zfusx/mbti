#!/usr/bin/env bash
set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required to run this script. Install jq and try again." >&2
  exit 1
fi

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

: "${DATABASE_URL:?Set DATABASE_URL before running this script}"

echo "Applying MBTI schema..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$REPO_ROOT/db/schema.sql"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

QUESTIONS_TSV="$TMP_DIR/questions.tsv"
RESULTS_TSV="$TMP_DIR/results.tsv"

jq -r '.[] | [.id, .dimension, tostring] | @tsv' "$REPO_ROOT/mbti92_en.json" > "$QUESTIONS_TSV"
jq -r '.types | to_entries[] | [.key, (.value | tostring)] | @tsv' "$REPO_ROOT/mbti_results_catalog_en_pure.json" > "$RESULTS_TSV"

echo "Seeding questions and results data..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 <<SQL
TRUNCATE TABLE mbti_questions RESTART IDENTITY CASCADE;
\copy mbti_questions (id, dimension, payload) FROM '$QUESTIONS_TSV' WITH (FORMAT csv, DELIMITER E'\t', QUOTE E'\x01');

TRUNCATE TABLE mbti_results;
\copy mbti_results (type, payload) FROM '$RESULTS_TSV' WITH (FORMAT csv, DELIMITER E'\t', QUOTE E'\x01');
SQL

echo "Done. Verify with: psql \"\$DATABASE_URL\" -c 'SELECT count(*) FROM mbti_questions;'"
