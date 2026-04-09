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
- **Granular Cookie Consent**: Privacy-first tracking that requires explicit opt-in for analytics and marketing cookies.

---

## 🍪 Cookie & Privacy Consent

El Bulk uses a granular consent management system to balance user privacy with store analytics. All non-essential tracking is **disabled by default**.

| Category | Description | Usage | State |
| :--- | :--- | :--- | :--- |
| **Essential** | Core site functionality. | Handles authentication sessions, active theme selection, and persistent cart metadata. | **Always Active** |
| **Analytics** | Traffic and behavior analysis. | Uses **GA4 (Google Analytics)** to help store owners understand which card sets and products are most popular. | **Opt-in** |
| **Marketing** | Personalized advertising. | Placeholder for future integrations like Facebook Pixel or social features. | **Opt-in** |

Users can customize these preferences at any time via the storefront banner. Consent status is persisted in the browser's `localStorage`.

---

## 📈 Analytics & Marketing Setup

To enable tracking, add the following IDs to your `frontend/.env.local` or production environment variables.

### 1. Google Analytics 4 (GA4)
- **Variable**: `NEXT_PUBLIC_GA_ID` (Format: `G-XXXXXXXXXX`)
- **How to get**: Go to [Google Analytics Console](https://analytics.google.com/) > Admin > Data Streams > Web > Copy **Measurement ID**.

### 2. Meta Pixel (Facebook)
- **Variable**: `NEXT_PUBLIC_META_PIXEL_ID` (Format: `1234567890`)
- **How to get**: Go to [Meta Events Manager](https://business.facebook.com/events_manager2/) > Data Sources > Settings > Copy **Pixel ID**.

### 3. Google Ads (Consent Mode v2)
- **Variable**: `NEXT_PUBLIC_GOOGLE_ADS_ID` (Format: `AW-XXXXXXXXXX`)
- **How to get**: Go to [Google Ads](https://ads.google.com/) > Tools & Settings > Conversions > Tag setup > Copy **Conversion ID**.

### 4. Hotjar
- **Variable**: `NEXT_PUBLIC_HOTJAR_ID` (Format: `1234567`)
- **How to get**: Go to [Hotjar Dashboard](https://insights.hotjar.com/) > Sites & Organizations > Copy **Site ID**.

---

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

## 🌩️ Google Cloud Platform (GCP) Production Guide

This guide focuses on a modern serverless stack using **Cloud Run** and **Cloud SQL** for high availability and zero maintenance.

> [!NOTE]
> **Why Caddy is omitted**: For Cloud Run, Caddy is not needed. GCP provides managed SSL certificates and edge routing natively. If you require path-based routing (e.g., `store.com/api`), it is handled by the **GCP Global HTTP(S) Load Balancer** rather than an internal container.

### 1. Infrastructure Preparation
Prepare your GCP environment and enable required services:
```bash
# Enable essential APIs
gcloud services enable \
    run.googleapis.com \
    sqladmin.googleapis.com \
    secretmanager.googleapis.com \
    artifactregistry.googleapis.com

# Create a private Docker repository
gcloud artifacts repositories create el-bulk-repo --repository-format=docker --location=us-central1
```

### 2. Managed Database (Cloud SQL)
Deploy a production-grade PostgreSQL 16 instance:
1. **Instance**: Create a Cloud SQL for PostgreSQL instance (e.g., `el-bulk-db`).
2. **Specs**: For small/medium stores, `db-f1-micro` or `db-g1-small` is sufficient.
3. **Security**: Use **Private IP** with Serverless VPC Access for maximum security (prevents the DB from being exposed to the public internet).
4. **Credentials**: Create a database user `elbulk` and a database `elbulk`.

### 3. Secret Management
 Centralize your production secrets in **Secret Manager**:
 - `ELBULK_JWT_SECRET`: Random 32+ char string.
 - `ELBULK_ENCRYPTION_KEY`: Persistent 32-char AES key.
 - `ELBULK_DB_PASSWORD`: Password for your Cloud SQL user.
 - `ELBULK_GOOGLE_CLIENT_SECRET`: For OAuth integration.
 
 > [!TIP]
 > **Quick Generation Commands**:
 > ```bash
 > # Generate and create the JWT Secret
 > openssl rand -base64 32 | gcloud secrets create ELBULK_JWT_SECRET --data-file=-
 > 
 > # Generate and create the 32-character Encryption Key
 > openssl rand -hex 16 | gcloud secrets create ELBULK_ENCRYPTION_KEY --data-file=-
 > 
 > # Create the DB Password secret (replace [PASSWORD])
 > echo "[PASSWORD]" | gcloud secrets create ELBULK_DB_PASSWORD --replication-policy="automatic" --data-file=-
 > ```
 
 ### 4. Production Deployment (Cloud Run)
Deploy the frontend and backend as independent serverless services:

**Backend Service**:
```bash
# Build and push image
docker build -t us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/backend ./backend
docker push us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/backend

# Deploy to Cloud Run
gcloud run deploy el-bulk-backend \
    --image us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/backend \
    --set-env-vars="DATABASE_URL=postgres://elbulk:[PASSWORD]@/elbulk?host=/cloudsql/[CONNECTION_NAME]" \
    --add-cloudsql-instances=[CONNECTION_NAME] \
    --region=us-central1
```

**Frontend Service**:
```bash
# Build with build-time API URL
docker build --build-arg NEXT_PUBLIC_API_URL=https://api.yourdomain.com -t us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/frontend ./frontend
docker push us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/frontend

# Deploy
gcloud run deploy el-bulk-frontend \
    --image us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/frontend \
    --region=us-central1
```

### 5. Domain & SSL (Global Balancer)
While Cloud Run provides `*.a.run.app` URLs with SSL, it is recommended to use **Cloud Load Balancing**:
- **Global Load Balancer**: Map your domain (e.g., `store.elbulk.com`) to the frontend service.
- **Google-Managed Certificates**: Automatically provisioned and renewed by GCP.
- **Cloud CDN**: Enable for the frontend to cache static MTG card images globally.

### 6. Automated Management
- **Backups**: Cloud SQL handles automated point-in-time recovery (PITR) by default.
- **Monitoring**: Use **Cloud Monitoring** dashboards to track request latency and DB CPU usage.
- **Alerting**: Set up an alert for `Database Storage > 80%`.

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

## 🛠️ Backend Refactoring Roadmap (Phased)

To ensure system stability, we are refactoring the backend logic from a monolithic handler-based architecture to a modern **Service/Repository** pattern.

### Phase 1: Foundation & Generic Store 🏗️
- [x] Implement `BaseStore` with Go generics for standard CRUD (List, Get, Create, Update, Delete).
- [x] Migrate **Categories** to the new architecture as a Proof of Concept.
- [x] **Goal**: Reduce boilerplate by 70% in simple CRUD entities.

### Phase 2: Core CRUD Migration ✅
- [x] Migrate **Themes**, **TCGs**, and **Notices** to use the `BaseStore`.
- [x] Standardize API response rendering across all migrated entities.

### Phase 3: Product Service & Enrichment ✅
- [x] Create `ProductStore` (extending BaseStore) and `ProductService`.
- [x] Centralize "Hot/New" discovery algorithms and pricing logic.
- [x] Decouple HTTP handlers from complex SQL filtering logic.

### Phase 4: Order Service & Transactions ✅
- [x] Implement `OrderService` to handle complex transactional state (Payment -> Confirmation -> Stock Adjustment).
- [x] Move PII encryption/decryption into centralized service middleware or helpers.

### Phase 5: Optimization & Cleanup ✅
- [x] Performance audit of the new Service layer.
- [x] Final removal of legacy redundant handler logic.
- [x] Consolidation of environment variables.


---

### Additional dvelopment:
- [x] the "restore inventory" modal is allowing to restore more items than the ones that were bought, fix it.
- [x] the "restore inventory" modal is not saving what was restored (visually when opening it again), either fix it or block it after correct item restoration.
- [x] restore inventory button should only appear if the order was canceled after confirming it. if the order was canceled before confirming it, it should not appear and automatically clear the correct ammount from the pending storage location, also check if the client canceled it and auto-restre stock (from pending storage location).
- [x] when an order is confirmed its inventory can not be modified anymore, but shipping, pickup and cancel status can be changed.
- [x] allow the admin to change payment method on orders and to modify shipping costs.
- [x] "ready to pick up", "shipped" and "completed" can only be selected after the order is confirmed.
- [x] "Cancelled" should be a status that can be selected at any time, and it should restore the inventory or request the admin to do it with the modal.
- [x] when an order is "Completed" or "canceled" it can't be modified anymore (with the exception of restoring stock for cancelled orders).
- [ ] sometimes "sell us your bulk" modal randomly closes, fix it (could not replicate).
- [x] the "ADD WANTED CARD" admin modal is not hidding the hovered card image after selecting a card, fix it.
- [x] admin /admin/clients/[id] has an order history scrollbar, add the same functionality to client requests and bounty offers also marking pending like on the order history one.
- [x] on client /profile make the bounty offers, client requests and orders scrollable.
- [x] on client orders show shipping as an item in the order items list.
- [x] on client orders, when the client cancel an order refresh the orders status to show the newly canceled status.
- [x] some combinations of cookies are allowing clients to reach checkout without login in, fix it. (hard to replicate, may be dev restarting docker that causes the bug)
- [x] modify accounting so that it takes into consideration "COST BASIS" when making csv reports and show it as outcome or egress, take into consideration product quantity, you can show the creation or modification date as the out data and the order confirmed or completed date as income.
- [x] add after the orange pi deployment guide a gcp deployment guide, use the same structure as the orange pi deployment guide, being thorough and detailed.


## ⚖️ License
MIT License. Created by Sniperkillerut.
