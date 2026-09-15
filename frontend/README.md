# MBTI92 frontend

The client is an Angular 20 single-page application built with Vite and Tailwind CSS.

Use Node.js 22 LTS, then run:

```bash
npm ci
npm run dev
```

Available checks and builds:

```bash
npm run test:ci
npm run build:ng
npm run build
```

The Vite development server listens on `http://localhost:4200` and proxies `/api` to the Go API at `http://localhost:8080`. See the repository's root `README.md` for database setup and production embedding instructions.
