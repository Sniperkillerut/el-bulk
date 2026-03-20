# El Bulk — TCG Web Store

Local TCG store web app. Buy singles, sealed, accessories. We also buy bulk.

**Stack**: Go (chi + sqlx) → PostgreSQL ← Next.js 14 (App Router, TypeScript)

---

## Quick Start (Docker — recommended)

```bash
# 1. Copy env template
cp backend/.env.example backend/.env

# 2. Start all services (postgres + backend + frontend)
docker compose up --build -d

# 3. Run seed data (first time only)
cd backend
go run ./seed/main.go
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Admin panel: http://localhost:3000/admin/login

Default admin credentials: `admin` / `elbulk2024!`

---

## Local Dev (no Docker)

**Requirements**: Go 1.23+, Node 20+, PostgreSQL 16

```bash
# 1. Create DB and run schema
psql -U postgres -c "CREATE DATABASE elbulk; CREATE USER elbulk WITH PASSWORD 'elbulkpass'; GRANT ALL ON DATABASE elbulk TO elbulk;"
psql -U elbulk -d elbulk -f backend/db/schema.sql

# 2. Backend
cd backend
cp .env.example .env        # edit DATABASE_URL for local postgres
go run main.go              # → :8080

# 3. Seed (in a separate terminal)
cd backend
go run ./seed/main.go

# 4. Frontend
cd frontend
npm install
cp .env.local.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:8080
npm run dev                 # → :3000
```

---

## URL Structure

| URL | Description |
|-----|-------------|
| `/` | Homepage |
| `/{tcg}/singles` | Singles grid (mtg, pokemon, lorcana, onepiece, yugioh) |
| `/{tcg}/sealed` | Sealed product |
| `/accessories` | Accessories (cross-TCG) |
| `/bulk` | Bring Your Bulk info page |
| `/product/:id` | Product detail |
| `/admin/login` | Admin login |
| `/admin/dashboard` | Product management |

## API

| Method | Endpoint | Auth |
|--------|----------|------|
| GET | `/api/products?tcg=mtg&category=singles&search=&foil=&treatment=&condition=` | — |
| GET | `/api/products/:id` | — |
| GET | `/api/tcgs` | — |
| POST | `/api/admin/login` | — |
| POST | `/api/admin/products` | Bearer JWT |
| PUT | `/api/admin/products/:id` | Bearer JWT |
| DELETE | `/api/admin/products/:id` | Bearer JWT |

---

## Adding a New TCG

1. Seed products with `tcg: "newgame"`
2. Add the TCG to `KNOWN_TCGS` in `frontend/src/lib/types.ts`
3. Add labels to `TCG_LABELS` and `TCG_SHORT` in the same file

No new page files needed — the `[tcg]` dynamic route handles it.
