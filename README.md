# MBTI92

MBTI92 is a 92-question personality questionnaire with balanced 20-, 40-, and 92-question modes. It combines an Angular frontend, a Go HTTP API, and PostgreSQL JSONB storage.

## Repository layout

- `frontend/` — Angular 20 client built with Vite and Tailwind CSS.
- `backend/` — Go API using Chi and pgx.
- `db/` — PostgreSQL schema.
- `scripts/` — database seeding and frontend embedding helpers.
- `mbti92_en.json` — question bank, evenly split across EI, SN, TF, and JP.
- `mbti_results_catalog_en_pure.json` — narrative content for all 16 types.

## Local setup

Requirements: Go 1.25+, Node.js 22 LTS, Docker, `psql`, `jq`, and `rsync`.

```bash
cp .env.example .env
# Replace the example password in both POSTGRES_PASSWORD and DATABASE_URL.
docker compose up -d

set -a
source .env
set +a
./scripts/load_data.sh
```

Run the API:

```bash
cd backend
go run ./cmd/api
```

Run the frontend in another terminal:

```bash
cd frontend
npm ci
npm run dev
```

The frontend is available at `http://localhost:4200` and proxies `/api` to the Go service on port 8080.

## Tests and builds

```bash
cd backend
go vet ./...
go test -race ./...

cd ../frontend
npm run test:ci
npm run build:ng
npm run build

cd ..
./scripts/embed_frontend.sh
git diff --exit-code backend/web/static
```

The API embeds the production frontend from `backend/web/static`. After changing frontend code, run the Vite build and `scripts/embed_frontend.sh`, then commit the updated embedded assets.

## Configuration

The API requires `DATABASE_URL`. Optional variables:

- `HTTP_PORT` — default `8080`.
- `ALLOWED_ORIGIN` — leave unset for same-origin production deployments; set to the development frontend origin when needed.
- `QUESTION_COUNT` — full questionnaire size, default `92`; must be divisible by four.

Never commit `.env`, database passwords, tokens, SSH keys, or production configuration.

## API

- `GET /healthz`
- `GET /api/questions?limit=20|40|92&random=true`
- `POST /api/sessions`
- `POST /api/sessions/{id}/answers`
- `GET /api/results/{type}`

Submitted answers are anonymous but are stored in PostgreSQL. Deployments should define an appropriate retention policy and disclose storage behavior to users.

## Content notice

The repository contains personality-test questions and result narratives. Confirm that you have the right to publish and redistribute this content before making the repository public. “MBTI” may be a protected trademark in some jurisdictions.
