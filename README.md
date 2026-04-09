# 📦 El Bulk — TCG Web Store

**El Bulk** is a professional-grade TCG (Trading Card Game) marketplace and inventory management system designed for local stores. It handles singles, sealed products, and accessories, while providing a powerful admin suite for order management and bulk buying operations.

---

## ✨ Features

### 🛒 Storefront
- **High-Performance Grid**: Instant filtering for singles and sealed products across all supported TCGs (MTG, Lorcana, Pokemon, etc.).
- **Smart Search**: Context-aware search with support for advanced card attributes (rarity, treatment, condition).
- **Persistent Cart**: Local storage-based cart that handles complex TCG item metadata.
- **Modern UI**: Full dark mode support, glassmorphism aesthetics, and responsive design.

### 🛡️ Security & Privacy
- **AES-256-GCM Encryption**: Sensitive customer data (phones, addresses, ID numbers) is encrypted before hitting the database.
- **mTLS Support**: Mutual TLS for secure database communication, now powered by dynamic environment-based certificate provisioning.
- **Anti-CSRF & Rate-Limiting**: Production-ready security middleware active by default.
- **Provider-Agnostic OAuth**: Seamless login via Google or Facebook.
- **Granular Cookie Consent**: Privacy-first tracking that requires explicit opt-in for analytics and marketing cookies.

### 🛠️ Admin Dashboard
- **Live Inventory Control**: Batch update stocks, prices, and locations.
- **Dynamic Theming**: Change the storefront appearance in real-time.
- **Internationalization**: Manage bilingual (ES/EN) translations and track completion percentages.
- **Bounty System**: Track and manage "Wanted" lists and customer offers.
- **Notices CMS**: Rich HTML blog/news system with card-reveal link integration.

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

## 📈 Analytics & Marketing Configuration Guide

El Bulk is pre-integrated with industry-standard tracking tools. Using these in tandem allows you to understand both *what* users are doing and *why* they are doing it.

### 📊 Quantitative Analysis (Data & Trends)
These tools provide "Aggregated Data" to track overall store health and campaign performance.

#### 1. Google Analytics 4 (GA4)
- **What it is**: Your primary cockpit for traffic analysis.
- **Used For**: Tracking page views, product popularity, user demographics, and acquisition paths.
- **How to Obtain**: Go to [Google Analytics 4](https://analytics.google.com/) > **Admin** > **Data Streams** > **Web** > Copy the **Measurement ID** (Format: `G-XXXXXXXXXX`).
- **Variable**: `NEXT_PUBLIC_GA_ID`

#### 2. Google Ads (Conversion Tracking)
- **What it is**: Measurement for your paid search campaigns.
- **Used For**: Calculating ROI on your ad spend. It identifies which specific ads led to a "Checkout" or a "Bounty Offer".
- **How to Obtain**: Go to [Google Ads](https://ads.google.com/) > **Goals** > **Conversions**. In the **Tag Setup** tab, look for the `gtag('config', 'AW-XXXXXXXXXX')` string.
- **Variable**: `NEXT_PUBLIC_GOOGLE_ADS_ID`

#### 3. Meta Pixel (Facebook/Instagram)
- **What it is**: A tracking script for the Meta ecosystem.
- **Used For**: Tracking conversions for Facebook Ads and building "Retargeting" audiences (e.g., showing an ad to someone who added a card to their cart but didn't buy).
- **How to Obtain**: Go to [Meta Events Manager](https://business.facebook.com/events_manager2/) > **Data Sources** > Click your Pixel > **Settings** > Copy the **Pixel ID**.
- **Variable**: `NEXT_PUBLIC_META_PIXEL_ID`

### 🛋️ Qualitative Analysis (User Experience)
These tools provide "Visual Insight" into the actual user experience.

#### 4. Hotjar
- **What it is**: User behavior visualization.
- **Used For**: **Heatmaps** (seeing where people click/scroll) and **Session Recordings** (watching a video replay of a user's visit).
- **How to Obtain**: Sign in to [Hotjar](https://insights.hotjar.com/). Add your domain (`elbulk.com`) to find the **Site ID** in the site list.
- **Variable**: `NEXT_PUBLIC_HOTJAR_ID`

---

> [!TIP]
> **Strategic Coexistence: "The Marketing Stack"**
> It is **recommended** to use all of these simultaneously. They do not conflict; they complement each other:
> 1. **Google Analytics** tells you that 20% of users leave on the "Checkout" page (**The Fact**).
> 2. **Hotjar** lets you watch recordings of those users to see that the "Pay" button is hidden behind a banner on some phones (**The Reason**).
> 3. **Meta Pixel** allows you to send those specific users an automated discount via Instagram to bring them back (**The Solution**).

---

## 🛠️ Technical Architecture

| Layer | Technology |
| :--- | :--- |
| **Backend** | Go (Chi, SQLX, Air) |
| **Frontend** | Next.js 15 (App Router, Tailwind CSS) |
| **Database** | PostgreSQL 16 (Alpine) |
| **Deployment** | Docker / Google Cloud Run |
| **Testing** | Vitest (Frontend), Go Test / Sqlmock (Backend) |

---

## 🚀 Local Development

### Prerequisites
- Docker & Docker Compose
- *Optional*: Go 1.25+, Node 20+

### Docker Setup (WSL / Docker Desktop)
```bash
# 1. Clone and enter
git clone https://github.com/Sniperkillerut/el-bulk.git && cd el-bulk

# 2. Environment Setup
cp .env.example .env
# Edit .env and set your ENCRYPTION_KEY (32-chars) and JWT_SECRET

# 3. Boot Services
docker compose -f docker-compose.dev.yml up --build
```
- **Storefront**: http://localhost:3000
- **Admin Login**: `admin` / `elbulk!2024` (or whatever you set in `.env`)

---

## 🛠️ Data Initialization

### 1. Database Seeding
To populate the database with initial data, use the modular seeding system. It supports two modes: `minimal` (safe for production) and `full` (exhaustive for development/testing).

#### Seeding Modes Comparison

| Feature | 🌱 Minimal Mode | 🌟 Full Mode |
| :--- | :---: | :---: |
| **Admin Account** | ✅ | ✅ |
| **TCGs & Categories** | ✅ | ✅ |
| **Store Settings & Themes** | ✅ | ✅ |
| **Bilingual Translations** | ✅ | ✅ |
| **Blog / Notices** | ✅ | ✅ |
| **Reference Products** | 1 (Black Lotus) | Hundreds (All TCGs) |
| **Scryfall Sync** | ❌ | ✅ (Real MTG Data) |
| **Customers & CRM** | ❌ | ✅ (30+ Profiles) |
| **Order History** | ❌ | ✅ (100+ Orders) |
| **Bounties & Offers** | ❌ | ✅ |

#### Running the Seed

```bash
# Run via Docker (Recommended)
docker exec -it el_bulk_backend_dev go run ./seed/ --mode=minimal  # Production
docker exec -it el_bulk_backend_dev go run ./seed/ --mode=full     # Development

# Run natively (requires local Go/Postgres)
cd backend
go run ./seed/ --mode=minimal
```

---

## 🌩️ Google Cloud Platform (GCP) Production Guide

This guide focuses on a modern serverless stack using **Cloud Run** and **Cloud SQL** for high availability and zero maintenance.

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
3. **Security**: Use **Private IP** with Serverless VPC Access for maximum security.
4. **Credentials**: Create a database user `elbulk` and a database `elbulk`.

### 3. Cloud Storage (GCS)
Create a bucket for product images and card assets:
1. **Creation**:
   ```bash
   gcloud storage buckets create gs://[BUCKET_NAME] --location=us-central1
   ```
2. **Permissions**: To allow the storefront to serve images, make the bucket publicly readable:
   ```bash
   gcloud storage buckets add-iam-policy-binding gs://[BUCKET_NAME] --member="allUsers" --role="roles/storage.objectViewer"
   ```

### 4. Environment Configuration
To ensure production security, El Bulk separates variables into **Secrets** (sensitive data) and **Public Config** (operational settings).

#### 4a. Secrets (GCP Secret Manager)
Create the following secrets. These are mounted into Cloud Run at runtime.

| Secret Name | Description | Value Generation |
| :--- | :--- | :--- |
| `ELBULK_DB_URL` | Postgres connection string | `postgres://[USER]:[PASS]@/[DB]?host=/cloudsql/[CONN]` |
| `ELBULK_JWT_SECRET` | Auth signing key | `openssl rand -base64 32` |
| `ELBULK_ENCRYPTION_KEY` | PII Encryption key | `openssl rand -hex 16` |
| `ELBULK_SMTP_PASS` | Email password | From your SMTP provider. |
| `ELBULK_GOOGLE_CLIENT_SECRET` | OAuth Secret | From Google Cloud Console. |
| `DB_SSL_ROOT_CERT` | Cloud SQL Server CA | Content of **server-ca.pem** (required for verify-full) |

#### 4b. Public Config (Cloud Run Environment)
These variables are passed directly to the `gcloud run deploy` command.

| Variable | Purpose | Suggested Value |
| :--- | :--- | :--- |
| `STORAGE_TYPE` | Storage engine | `gcp` |
| `GCP_BUCKET_NAME` | Cloud Storage bucket | `[YOUR_GCS_BUCKET_NAME]` |
| `APP_ENV` | Runtime environment | `production` |
| `FRONTEND_ORIGIN` | CORS Security | `https://elbulk.com` |
| `SITE_URL` | Email link generation | `https://elbulk.com` |
| `SMTP_HOST` / `PORT` | Email infrastructure | e.g., `smtp.sendgrid.net` / `587` |

> [!NOTE]
> **Next.js Tracking IDs**: The following variables are built into the frontend image at **build-time**. They do not need to be set in Cloud Run environment variables unless using a specific dynamic-runtime configuration.
> - `NEXT_PUBLIC_GA_ID`
> - `NEXT_PUBLIC_META_PIXEL_ID`
> - `NEXT_PUBLIC_GOOGLE_ADS_ID`
> - `NEXT_PUBLIC_HOTJAR_ID`

### 5. Production Deployment (Cloud Run)

**Backend Service**:
```bash
gcloud run deploy el-bulk-backend \
    --image us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/backend \
    --add-cloudsql-instances=[CONNECTION_NAME] \
    --region=us-central1 \
    --set-secrets="DATABASE_URL=ELBULK_DB_URL:latest,ENCRYPTION_KEY=ELBULK_ENCRYPTION_KEY:latest,JWT_SECRET=ELBULK_JWT_SECRET:latest,SMTP_PASS=ELBULK_SMTP_PASS:latest,GOOGLE_CLIENT_SECRET=ELBULK_GOOGLE_CLIENT_SECRET:latest,DB_SSL_ROOT_CERT=DB_SSL_ROOT_CERT:latest" \
    --set-env-vars="STORAGE_TYPE=gcp,GCP_BUCKET_NAME=[BUCKET],APP_ENV=production,FRONTEND_ORIGIN=https://elbulk.com,SITE_URL=https://elbulk.com,SMTP_HOST=[HOST],SMTP_PORT=[PORT],SMTP_USER=[USER],SMTP_FROM=[FROM],GOOGLE_CLIENT_ID=[ID]"
```

**Frontend Service**:
```bash
# Build with build-time variables
docker build \
    --build-arg NEXT_PUBLIC_API_URL=https://api.elbulk.com \
    --build-arg NEXT_PUBLIC_GA_ID=[GA_ID] \
    --build-arg NEXT_PUBLIC_META_PIXEL_ID=[PIXEL_ID] \
    --build-arg NEXT_PUBLIC_GOOGLE_ADS_ID=[ADS_ID] \
    --build-arg NEXT_PUBLIC_HOTJAR_ID=[HOTJAR_ID] \
    -t us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/frontend ./frontend
docker push us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/frontend

# Deploy
gcloud run deploy el-bulk-frontend \
    --image us-central1-docker.pkg.dev/[PROJECT_ID]/el-bulk-repo/frontend \
    --region=us-central1
```

### 6. Domain & SSL (Global Balancer)
Map your naked domain (**elbulk.com**) to the frontend service using the **GCP Global HTTP(S) Load Balancer** with Google-Managed SSL certificates.

### 7. Automated Management
- **Backups**: Cloud SQL handles automated point-in-time recovery (PITR).
- **Monitoring**: Use **Cloud Monitoring** to track request latency and DB usage.
- **Alerting**: Set up an alert for `Database Storage > 80%`.

---

## 🧪 Testing

### Backend
```bash
cd backend
go test ./...
```

### Frontend
```bash
cd frontend
npm run test
```

## ⚖️ License
MIT

## to fix:
- [ ] admin selecting "confirmed" on the dropdown bypasses confirmation modal and wont reduce inventory from storage locations, fix it.
- [ ] when order in "ready for pick up" or "shipped" the confirm button should not appear as the order is already confirmed. (maybe add a confirmed flag?)
- [ ] when admin sets an order as "completed" there should be an alert popup that the chang is final and order will lock and wont be editable anymore.
- [ ] same for cancelled 
- [ ] restore to inventory is allowing for multiple restores,  artifiaclly increasing the stock quantity, this is a bug, fix it.
- [ ] restore to inventory should allow to add stock to multiple storage locations and to add to a newly selected storage location.
