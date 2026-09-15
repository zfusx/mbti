# MBTI92 Mini Platform Plan

## 1. Overview
- **Goal**: Deliver a single-page MBTI test experience powered by Angular 20 (standalone components + signals) with a Go backend service and PostgreSQL JSONB storage.
- **Assets**: `mbti92_en.json` (questions), `mbti92_result_logic_only.json` (scoring spec), `mbti_results_catalog_en_pure.json` (result copy).

## 2. Architecture
1. **Client (Angular 20 SPA)**
   - Vite-based Angular build for faster dev server.
   - Tailwind CSS with design tokens for colors/spacing.
   - State handled via Angular signals + router data resolvers.
2. **Backend (Go)**
   - Go 1.25+ using Chi for routing.
   - Modular packages: `questions`, `scoring`, `results`, `storage`.
   - JSON API consumed by the SPA.
3. **Database (PostgreSQL with JSONB)**
   - Tables:
     - `mbti_questions (id INT PK, dimension TEXT, payload JSONB)`
     - `mbti_results (type CHAR(4) PK, payload JSONB)`
     - `mbti_sessions (id UUID PK, answers JSONB, created_at TIMESTAMPTZ)`
   - Use `jsonb_path_ops` GIN indexes for searching/filtering.

## 3. Data Flow
1. Angular fetches question batches (`GET /questions?random=true`).
2. User answers stored locally; on submit SPA posts payload to Go service.
3. Go service:
   - Validates payload using `mbti92_result_logic_only.json`.
   - Computes tallies, percentages, type.
   - Persists session summary (optional) in PostgreSQL.
4. SPA receives `{type, tally, pairs, completion, confidence}` and pulls extended copy via `GET /results/:type`.

## 4. API Endpoints (Go)
| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/questions` | Return all or randomized subset; query params `limit`, `random`. |
| POST | `/api/sessions` | Create session if you need persistence; returns session ID. |
| POST | `/api/sessions/{id}/answers` | Accept answers array, run scoring, store results. |
| GET | `/api/results/{type}` | Fetch narrative data for a type. |

## 5. Frontend Screens
1. **Landing**: intro, CTA.
2. **Quiz**:
   - Card carousel with progress bar and optional question map.
   - Tailwind components for accessibility (focus styles, keyboard navigation).
   - Random order toggle (client-side shuffle or server-driven).
3. **Review & Submit**: quick overview before scoring.
4. **Results**:
   - Display four-letter type, pair tallies, confidence gap, descriptive copy, matches/careers.
   - Share/retake buttons.

## 6. Implementation Steps
1. **Backend**
   - Initialize Go module, add router, configuration, DB migrations (e.g., goose).
   - Write ingestion script to load JSON assets into PostgreSQL JSONB tables.
   - Implement scoring as pure function with tests mirroring the logic spec.
2. **Frontend**
   - Scaffold Angular 20 project with Tailwind; set up routing + layout shell.
   - Build question service + store, quiz components, results component.
   - Integrate API calls and add optimistic UI states (loading, error, resume session).
3. **Testing**
   - Go unit tests for scoring + handlers, integration tests using `httptest`.
   - Angular unit tests for components/services and Cypress (or Playwright) e2e covering happy path.
4. **DevOps**
   - Dockerfiles for Go API and PostgreSQL migrations; Vite build for Angular.
   - CI pipeline (GitHub Actions) that runs Go tests, Angular tests, and builds docker images.
   - Deploy: SPA to static hosting (Vercel/S3+CloudFront); Go API to Fly.io/Render with managed Postgres.

## 7. Enhancements (Later)
- Add authentication or anonymous session tracking.
- Support questionnaire variants (short/long) by tagging questions in JSONB.
- Add analytics dashboard using stored session data.
- Internationalization by adding more locale JSON files and toggling in the SPA.
