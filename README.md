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

---

## Deployment (Self-Hosting on Orange Pi / Raspberry Pi)

The **Orange Pi Zero 2W (4GB RAM)** is the recommended self-hosting target. 

### 1. Production Docker Setup
On your Pi, use the production-optimized compose file (not provided in dev, but follows this pattern):

```yaml
services:
  db: { image: postgres:16-alpine, ... }
  backend:
    build: { context: ./backend, target: prod }
    environment: { PORT: "8080", ... }
  frontend:
    build: { context: ./frontend, target: prod }
    ports: ["80:3000"]
```

### 2. Cloudflare Tunnel (Zero Trust)
To expose your store to the internet without opening ports:

1.  **Install `cloudflared`** on the Orange Pi.
2.  **Authenticate**: `cloudflared tunnel login`.
3.  **Create Tunnel**: `cloudflared tunnel create el-bulk-tunnel`.
4.  **Configure**: Create `config.yml`:
    ```yaml
    tunnel: <TUNNEL_ID>
    credentials-file: /home/pi/.cloudflared/<TUNNEL_ID>.json
    ingress:
      - hostname: store.yourdomain.com
        service: http://localhost:80
      - service: http_status:404
    ```
5.  **Route DNS**: `cloudflared tunnel route dns el-bulk-tunnel store.yourdomain.com`.
6.  **Run**: `cloudflared tunnel run el-bulk-tunnel`.

### 3. Automated Backups (Crontab)
To automate the daily database dump, add this to your Pi’s crontab (`crontab -e`):

```bash
# Backup every day at 3:00 AM
00 03 * * * docker exec el_bulk_db pg_dump -U elbulk elbulk > /home/pi/backups/elbulk_$(date +\%Y\%m\%d).sql

# (Optional) Delete backups older than 30 days
00 04 * * * find /home/pi/backups/ -name "*.sql" -type f -mtime +30 -delete
```

### 4. Performance Tips for Pi
- **Swap**: Ensure at least 2GB of swap space is enabled (`sudo fallocate -l 2G /swapfile`).
- **Storage**: Boot from a USB SSD if possible to avoid MicroSD card wear from PostgreSQL writes.
