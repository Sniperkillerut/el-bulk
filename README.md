# 📦 El Bulk — TCG Web Store

**El Bulk** is a professional-grade TCG (Trading Card Game) marketplace and inventory management system designed for local stores. It handles singles, sealed products, and accessories, while providing a powerful admin suite for order management and bulk buying operations.

> [!TIP]
> Built for performance and privacy, El Bulk features application-level PII encryption and is optimized for low-power ARM64 hardware like the Orange Pi Zero 2W.

---

## ✨ Features

### 🛒 Storefront
- **High-Performance Grid**: Instant filtering for singles and sealed products across all supported TCGs (MTG, Lorcana, etc.).
- **Smart Search**: Context-aware search with support for advanced card attributes (rarity, treatment, condition).
- **Persistent Cart**: Local storage-based cart that handles complex TCG item metadata.
- **Modern UI**: Full dark mode support, glassmorphism aesthetics, and responsive design.

### 🛡️ Security & Privacy
- **AES-256-GCM Encryption**: Sensitive customer data (phones, addresses, ID numbers) is encrypted before hitting the database.
- **mTLS Suport**: Mutual TLS for secure service-to-service communication.
- **Anti-CSRF & Rate-Limiting**: Production-ready security middleware active by default.
- **Provider-Agnostic OAuth**: Seamless login via Google or Facebook.

### 🛠️ Admin Dashboard
- **Live Inventory Control**: Batch update stocks, prices, and locations.
- **Dynamic Theming**: Change the storefront appearance in real-time.
- **Internationalization**: Manage bilingual (ES/EN) translations and track completion percentages.
- **Bounty System**: Track and manage "Wanted" lists and customer offers.
- **Notices CMS**: Rich HTML blog/news system with card-reveal link integration.

---

## 🛠️ Technical Architecture

| Layer | Technology |
| :--- | :--- |
| **Backend** | Go (Chi, SQLX, Air) |
| **Frontend** | Next.js 15 (App Router, Tailwind CSS) |
| **Database** | PostgreSQL 16 (Alpine) |
| **Reverse Proxy** | Caddy (Auto-HTTPS) |
| **Testing** | Vitest (Frontend), Go Test / Sqlmock (Backend) |

---

## ☁️ Cloud Deployment & Cost Analysis

El Bulk is architecture-agnostic. While specialized for ARM64 SBCs, it runs perfectly on major cloud providers.

### 💰 Cost Comparison Table (Est. Monthly)

| Provider | Service Type | Compute (vCPU/RAM) | Database | Est. Cost | Best For |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Oracle Cloud** | VM (ARM) | 4 OCPU / 24GB | Self-hosted (Free) | **$0.00** | Budget / Performance |
| **GCP** | Cloud Run | Serverless (Autoscale) | Cloud SQL (Postgres) | **~$10.00** | High traffic / Scale |
| **AWS** | App Runner | Serverless (Autoscale) | RDS (db.t4g.micro) | **~$15.00** | AWS Ecosystem |
| **DigitalOcean** | Droplet | 1 vCPU / 1GB | Managed DB | **~$21.00** | Simple Billing |
| **Hetzner** | Cloud VPS | 2 vCPU / 2GB | Docker Postgres | **~$4.50** | European Performance |

---

### 🏛️ Oracle Cloud (Recommended Free Tier)
Oracle's **Always Free ARM tier** is the strongest choice for hosting El Bulk for free at production speeds.
1. **Compute**: Use `VM.Standard.A1.Flex` with 4 OCPUs and 24GB RAM.
2. **Setup**: Use the same instructions as the Orange Pi, but without the CPU scaling tweaks.
3. **Storage**: 200GB block storage included for free.
4. **Egress**: 10TB per month covers almost any TCG store.

### 🌩️ Google Cloud Platform (GCP)
Best for users prioritizing high availability and serverless simplicity.
1. **Frontend/Backend**: Deploy to **Cloud Run** using the provided `frontend/Dockerfile` and `backend/Dockerfile`.
2. **Database**: Use **Cloud SQL** (PostgreSQL) with a `db-f1-micro` instance.
3. **Secret Manager**: Store `ENCRYPTION_KEY` and `JWT_SECRET` in GCP Secrets.

### 📦 AWS (Classic Infrastructure)
1. **ECS/App Runner**: Use **App Runner** for the simplest managed container experience.
2. **RDS**: Use a `db.t4g.micro` PostgreSQL instance (covered by the **12-month free tier**).
3. **S3**: Use for automated database backups via AWS CLI.

---

## 🚀 Local Development

### Prerequisites
- Docker & Docker Compose
- *Optional*: Go 1.23+, Node 20+, PostgreSQL 16 (for native dev)

### Docker Setup (Recommended)
```bash
# 1. Clone and enter
git clone https://github.com/Sniperkillerut/el-bulk.git && cd el-bulk

# 2. Environment Setup
cp backend/.env.example backend/.env
# Edit backend/.env and set your ENCRYPTION_KEY (32 chars)

# 3. Boot Services
docker compose -f docker-compose.dev.yml up --build

# 4. Data Initialization (New Section)
# For more detailed instructions, see the "Initialization & Setup Scripts" section below.
```
- **Storefront**: http://localhost:3000
- **Admin Login**: `admin` / `elbulk2024!`

---

## 🛠️ Initialization & Setup Scripts

Before starting the application for the first time, or when setting up a new environment, you need to run several scripts to prepare the infrastructure and data.

### 1. Database Certificate Generation
El Bulk uses **mTLS (Mutual TLS)** for secure database connections. You must generate the certificates before starting the services if they are not already present in the `/certs` directory.

```bash
# Run from the project root
bash ./scripts/generate-db-certs.sh
```
This script will:
- Generate a Root CA.
- Create Server Key and Certificate for the database.
- Create Client Key and Certificate for the backend service.
- Set appropriate permissions (0600) for the private keys.

### 2. Database Seeding
To populate the database with initial products, categories, and translation keys, run the seeding script:

```bash
# Option A: Run via Docker (Recommended if services are running)
docker exec -it el_bulk_backend go run ./seed/main.go

# Option B: Run natively (Requires local Go installation)
cd backend
go run ./seed/main.go
```
The seeding script will:
- Initialize the product taxonomy (Singles, Sealed, Accessories).
- Add initial stock for all supported TCGs.
- Populate the translation table with bilingual keys (ES/EN).

---

## 🛡️ Security Deep Dive

### PII Encryption
El Bulk encrypts sensitive fields at the application level. To rotate keys or migrate plain-text data:
```bash
cd backend
go run scripts/migrate_pii/main.go
```

### Database mTLS
For production, the database requires certificate-based authentication. Certificates are located in `/certs` and managed by the backend during the connection handshake.

---

## 🍊 Orange Pi Zero 2W (4GB) Production Guide

This guide assumes you are using **Debian 12 (Bookworm)** or **Armbian 23.x**.

### 1. Hardware Preparation
- **Power**: Use a high-quality 5V 3A USB-C power supply.
- **Cooling**: A heat sink on the H618 chip is **mandatory** for production loads.
- **Storage**: For database longevity, use a USB 3.0 SSD. If using a TF card, ensure it is Class 10/A2.

### 2. OS Optimization
Prepare the system for high I/O and RAM efficiency:
```bash
# Update and install dependencies
sudo apt update && sudo apt upgrade -y
sudo apt install docker.io docker-compose zram-config ufw -y

# Enable ZRAM (Crucial for 4GB RAM)
sudo systemctl enable zram-config
sudo systemctl start zram-config

# Tune Swappiness
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### 3. Production Deployment
Use the optimized production compose file which includes the **Caddy** reverse proxy.

```bash
cd el-bulk
# Configure Prod Env
cp .env.example .env.prod
# Add SITE_DOMAIN, ACME_EMAIL, and security secrets

# Deploy
docker compose -f docker-compose.prod.yml up -d
```

### 4. Reverse Proxy & SSL
The included `Caddyfile` automatically provisions SSL certificates via Let's Encrypt.
- Ensure ports **80** and **443** are forwarded at your router.
- Set `SITE_DOMAIN=store.yourdomain.com` in your production env.

### 5. Automated Backups
Schedule daily database dumps to prevent data loss:
```bash
# Edit crontab: crontab -e
00 03 * * * docker exec el_bulk_db_prod pg_dump -U elbulk elbulk > /home/pi/backups/elbulk_$(date +\%Y\%m\%d).sql
```

---

## 🧪 Testing

### Backend
```bash
cd backend
go test ./... -v -cover
```

### Frontend
```bash
cd frontend
npm run test:coverage
```

---

## 🗺️ System Roadmap
- [ ] Stripe/MercadoPago Integration
- [ ] WhatsApp Order Integration
- [ ] Advanced TCGPlayer/Scryfall Stock Sync
- [ ] Multi-store multi-inventory support

---

## ⚖️ License
MIT License. Created by Sniperkillerut.
