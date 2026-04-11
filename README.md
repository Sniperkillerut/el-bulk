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

For a **complete, step-by-step deployment guide**, see:

👉 **[docs/GCP_DEPLOYMENT.md](docs/GCP_DEPLOYMENT.md)**

It covers everything from project creation to production monitoring:

| Section | What It Covers |
| :--- | :--- |
| **Prerequisites** | Tools, accounts, and values you'll need |
| **GCP Project Setup** | APIs, billing, and region config |
| **Cloud SQL** | PostgreSQL instance, IAM auth, connection names |
| **Cloud Storage** | Image bucket creation and public access |
| **Secret Manager** | All 6 secrets with generation commands |
| **IAM Permissions** | Exact roles for Cloud Run and Cloud Build |
| **Cloud Build** | One-command automated deployment |
| **Manual Deployment** | Per-service Docker build & deploy |
| **Database Seeding** | Production vs development seed modes |
| **Custom Domains & SSL** | Domain mapping, DNS records, auto SSL |
| **OAuth Setup** | Google & Facebook provider configuration |
| **SMTP / Email** | SendGrid, Brevo, and Gmail options |
| **Monitoring & Alerting** | Logs, dashboards, and alert policies |
| **Troubleshooting** | CORS, DB, OAuth, cold starts, images |
| **Environment Reference** | Complete table of all env vars and secrets |

### Quick Start (TL;DR)

```bash
# 1. Enable APIs & create infra
gcloud services enable run.googleapis.com sqladmin.googleapis.com \
    secretmanager.googleapis.com artifactregistry.googleapis.com cloudbuild.googleapis.com
gcloud artifacts repositories create el-bulk-repo --repository-format=docker --location=us-central1
gcloud sql instances create el-bulk-db --database-version=POSTGRES_16 --tier=db-f1-micro --region=us-central1

# 2. Create secrets (see full guide for all 6)
echo -n "$(openssl rand -base64 32)" | gcloud secrets create ELBULK_JWT_SECRET --data-file=-

# 3. Deploy everything
gcloud builds submit --config cloudbuild.yaml \
    --substitutions=_DB_CONNECTION_NAME="PROJECT:REGION:INSTANCE",_GCS_BUCKET="your-bucket"
```

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