# 🌩️ El Bulk — GCP Production Deployment Guide

**Last Updated:** 2026-04-12

A complete, step-by-step guide to deploying the El Bulk TCG Store on Google Cloud Platform using **Cloud Run**, **Cloud SQL**, **Cloud Storage**, and **Cloud Build**.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [GCP Project Setup](#2-gcp-project-setup)
3. [Artifact Registry (Docker Repository)](#3-artifact-registry-docker-repository)
4. [Cloud SQL (PostgreSQL Database)](#4-cloud-sql-postgresql-database)
5. [Cloud Storage (Image Uploads)](#5-cloud-storage-image-uploads)
6. [Secret Manager (Sensitive Configuration)](#6-secret-manager-sensitive-configuration)
7. [IAM Permissions](#7-iam-permissions)
8. [Automated Deployment (Cloud Build)](#8-automated-deployment-cloud-build)
9. [Manual Deployment (Alternative)](#9-manual-deployment-alternative)
10. [Database Initialization & Seeding](#10-database-initialization--seeding)
11. [Custom Domains & SSL](#11-custom-domains--ssl)
12. [DNS Configuration](#12-dns-configuration)
13. [OAuth Provider Setup](#13-oauth-provider-setup)
14. [SMTP / Email Setup](#14-smtp--email-setup)
15. [Monitoring & Alerting](#15-monitoring--alerting)
16. [Backups & Disaster Recovery](#16-backups--disaster-recovery)
17. [Cost Optimization](#17-cost-optimization)
18. [Troubleshooting](#18-troubleshooting)
19. [Maintenance Operations](#19-maintenance-operations)
20. [Environment Variable Reference](#20-environment-variable-reference)
21. [CI/CD: Automated Triggers](#21-cicd-automated-triggers)

---

## 1. Prerequisites

### Tools Required

Install these on your local machine before starting:

```bash
# Google Cloud SDK
curl -sSL https://sdk.cloud.google.com | bash
exec -l $SHELL
gcloud init

# Docker (for local testing)
# Install Docker Desktop or Docker Engine: https://docs.docker.com/get-docker/

# Verify installations
gcloud --version
docker --version
```

### Accounts Required

| Account | Purpose | Where to Get |
|:---|:---|:---|
| **Google Cloud** | Infrastructure hosting | [console.cloud.google.com](https://console.cloud.google.com) |
| **Google OAuth** | Customer social login | [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials) |
| **Facebook OAuth** | Customer social login (optional) | [developers.facebook.com](https://developers.facebook.com) |
| **SMTP Provider** | Newsletter / transactional email | SendGrid, Brevo, Mailgun, or Gmail App Password |
| **Domain Registrar** | Custom domain (elbulk.com) | Namecheap, GoDaddy, Google Domains, etc. |

### Information You'll Need

Before starting, gather these values:

| Value | Example | Notes |
|:---|:---|:---|
| GCP Project ID | `my-elbulk-prod` | Globally unique |
| GCP Region | `us-central1` | Choose closest to your customers |
| Domain | `elbulk.com` | Your storefront domain |
| Admin credentials | `admin` / `YourSecurePass!` | For the admin dashboard |

---

## 2. GCP Project Setup

### 2.1 Create Project

```bash
# Create a new project (or use existing)
gcloud projects create my-elbulk-prod --name="El Bulk Production"

# Set as active project
gcloud config set project my-elbulk-prod
```

### 2.2 Enable Billing

> [!CAUTION]
> Cloud SQL and Cloud Run **require billing** to be enabled. Go to:
> [console.cloud.google.com/billing](https://console.cloud.google.com/billing)
> and link a billing account to your project.

### 2.3 Enable Required APIs

```bash
gcloud services enable \
    run.googleapis.com \
    sqladmin.googleapis.com \
    secretmanager.googleapis.com \
    artifactregistry.googleapis.com \
    cloudbuild.googleapis.com \
    cloudtrace.googleapis.com \
    compute.googleapis.com \
    iam.googleapis.com
```

### 2.4 Set Default Region

```bash
gcloud config set run/region us-central1
gcloud config set compute/region us-central1
```

---

## 3. Artifact Registry (Docker Repository)

Cloud Build needs a Docker repository to store your container images.

```bash
# Create repository
gcloud artifacts repositories create el-bulk-repo \
    --repository-format=docker \
    --location=us-central1 \
    --description="El Bulk container images"

# Verify it was created
gcloud artifacts repositories list --location=us-central1
```

**Expected output:**
```
REPOSITORY    FORMAT  LOCATION       DESCRIPTION
el-bulk-repo  DOCKER  us-central1    El Bulk container images
```

---

## 4. Cloud SQL (PostgreSQL Database)

### 4.1 Create the Instance

```bash
gcloud sql instances create el-bulk-db \
    --database-version=POSTGRES_16 \
    --edition=ENTERPRISE \
    --tier=db-f1-micro \
    --region=us-central1 \
    --storage-type=SSD \
    --storage-size=10GB \
    --storage-auto-increase \
    --backup-start-time="03:00" \
    --availability-type=zonal \
    --database-flags=log_min_duration_statement=1000
```

> [!NOTE]
> **Tier selection:**
> - `db-f1-micro` (~$9/month) — Good for development/small stores
> - `db-g1-small` (~$27/month) — Recommended for production with moderate traffic
> - `db-custom-2-4096` (~$50/month) — For high-traffic stores

### 4.2 Create the Database

```bash
gcloud sql databases create elbulk --instance=el-bulk-db
```

### 4.3 Set the Default User Password

```bash
# Set a strong password for the postgres user
gcloud sql users set-password postgres \
    --instance=el-bulk-db \
    --password="$(openssl rand -base64 24)"
```

> [!TIP]
> Save this password securely. You'll need it for the `DATABASE_URL` secret later if not using IAM authentication.

### 4.4 Get Connection Name

This value is critical — you'll use it in every deployment command.

```bash
gcloud sql instances describe el-bulk-db --format="value(connectionName)"
```

**Save the output.** It looks like: `my-elbulk-prod:us-central1:el-bulk-db`

### 4.5 (Recommended) Enable IAM Authentication

IAM auth eliminates password management — Cloud Run authenticates using its service account identity.

```bash
# Enable IAM auth on the instance
gcloud sql instances patch el-bulk-db \
    --database-flags=cloudsql.iam_authentication=on
```

Then create an IAM database user for the Cloud Run service account:

```bash
# Get the Cloud Run service account email
# (Correctly identifies the default compute service account)
SA_EMAIL=$(gcloud iam service-accounts list \
    --filter="email ~ ^[0-9]+-compute@developer.gserviceaccount.com$" \
    --format="value(email)")
echo "Service Account: $SA_EMAIL"

# Create IAM user in Cloud SQL
# NOTE: The suffix ".gserviceaccount.com" MUST be stripped for this command
IAM_USER=$(echo "$SA_EMAIL" | sed 's/\.gserviceaccount\.com$//')

gcloud sql users create "$IAM_USER" \
    --instance=el-bulk-db \
    --type=CLOUD_IAM_SERVICE_ACCOUNT
```

Then grant the IAM user access to the database by running this block (it uses the `$IAM_USER` variable from the step above):

```bash
# Connect to the database and grant permissions
# You will be prompted for the postgres user password
# We ensure the database exists first just in case Step 4.2 was skipped
gcloud sql databases create elbulk --instance=el-bulk-db 2>/dev/null || echo "Database exists"

gcloud sql connect el-bulk-db --user=postgres --database=elbulk <<EOF
GRANT ALL PRIVILEGES ON DATABASE elbulk TO "$IAM_USER";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "$IAM_USER";
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "$IAM_USER";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "$IAM_USER";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "$IAM_USER";
EOF
```



---

## 5. Cloud Storage (Image Uploads)

Product images and admin-uploaded media are stored in Google Cloud Storage.

### 5.1 Create the Bucket

```bash
# Replace with your actual bucket name (must be globally unique)
gcloud storage buckets create gs://elbulk-media-prod \
    --location=us-central1 \
    --default-storage-class=STANDARD \
    --uniform-bucket-level-access
```

### 5.2 Make Images Publicly Readable

The storefront needs to serve images directly from the bucket:

```bash
gcloud storage buckets add-iam-policy-binding gs://elbulk-media-prod \
    --member="allUsers" \
    --role="roles/storage.objectViewer"
```

### 5.3 (Optional) Set CORS for Direct Uploads

If you plan to upload images directly from the browser in the future:

```bash
cat > /tmp/cors.json << 'EOF'
[
  {
    "origin": ["https://elbulk.com", "https://api.elbulk.com"],
    "method": ["GET", "PUT", "POST"],
    "responseHeader": ["Content-Type"],
    "maxAgeSeconds": 3600
  }
]
EOF

gcloud storage buckets update gs://elbulk-media-prod --cors-file=/tmp/cors.json
```

### 5.4 Verify

```bash
gcloud storage ls gs://elbulk-media-prod
```

---

## 6. Secret Manager (Sensitive Configuration)

Secrets are mounted into Cloud Run containers at runtime. They are **never** baked into Docker images.

### 6.1 Generate Secrets

```bash
# Generate cryptographically secure values
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -hex 16)

echo "JWT_SECRET:     $JWT_SECRET"
echo "ENCRYPTION_KEY: $ENCRYPTION_KEY"
```

> [!CAUTION]
> **Save these values securely (password manager).** If you lose `ENCRYPTION_KEY`, all encrypted PII (customer phones, addresses, ID numbers) becomes **permanently unrecoverable**.

### 6.2 Create Each Secret

```bash
# Database URL
# If using IAM auth (uses the $IAM_USER variable defined in Step 4.5):
echo -n "user=$IAM_USER dbname=elbulk sslmode=disable" | \
    gcloud secrets create ELBULK_DB_URL --data-file=-

# If using password auth:
echo -n "postgres://<USER>:<PASSWORD>@/<DB_NAME>?host=/cloudsql/<CONNECTION_NAME>&sslmode=disable" | \
    gcloud secrets create ELBULK_DB_URL --data-file=-

# JWT Secret
echo -n "$JWT_SECRET" | \
    gcloud secrets create ELBULK_JWT_SECRET --data-file=-

# Encryption Key (32 characters for AES-256)
echo -n "$ENCRYPTION_KEY" | \
    gcloud secrets create ELBULK_ENCRYPTION_KEY --data-file=-

# SMTP Password
echo -n "your-smtp-password" | \
    gcloud secrets create ELBULK_SMTP_PASS --data-file=-

# Google OAuth Client Secret
echo -n "GOOGLExxxxx.apps.googleusercontent.com-secret" | \
    gcloud secrets create ELBULK_GOOGLE_CLIENT_SECRET --data-file=-

# Facebook OAuth Client Secret (optional)
echo -n "facebook-app-secret-here" | \
    gcloud secrets create ELBULK_FACEBOOK_CLIENT_SECRET --data-file=-
```

### 6.3 Verify Secrets Exist

```bash
gcloud secrets list --format="table(name, createTime)"
```

**Expected output:**
```
NAME                            CREATE_TIME
ELBULK_DB_URL                   2026-04-11...
ELBULK_ENCRYPTION_KEY           2026-04-11...
ELBULK_FACEBOOK_CLIENT_SECRET   2026-04-11...
ELBULK_GOOGLE_CLIENT_SECRET     2026-04-11...
ELBULK_JWT_SECRET               2026-04-11...
ELBULK_SMTP_PASS                2026-04-11...
```

### 6.4 Update a Secret Value

If you ever need to rotate a secret:

```bash
echo -n "new-secret-value" | \
    gcloud secrets versions add ELBULK_JWT_SECRET --data-file=-
```

---

## 7. IAM Permissions

Cloud Run's service account needs specific roles to access GCP resources.

### 7.1 Identify the Service Account

```bash
# Cloud Run uses the Compute Engine default service account unless customized
SA_EMAIL=$(gcloud iam service-accounts list \
    --filter="email ~ ^[0-9]+-compute@developer.gserviceaccount.com$" \
    --format="value(email)")
echo "Service Account: $SA_EMAIL"
```

### 7.2 Grant Required Roles

```bash
# Fetch current project ID
PROJECT_ID=$(gcloud config get-value project)

# Verify Project ID before proceeding (Safe for interactive shell)
if [ -z "$PROJECT_ID" ]; then
    echo "❌ ERROR: PROJECT_ID is empty. Run 'gcloud config set project [NAME]' first."
else
    echo "✅ Project ID detected: $PROJECT_ID"
fi

# Access Cloud SQL
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/cloudsql.client"

# IAM-based database login
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/cloudsql.instanceUser"

# Upload images to Cloud Storage
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/storage.objectAdmin"

# Read secrets from Secret Manager
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/secretmanager.secretAccessor"

# Send traces to Cloud Trace
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/cloudtrace.agent"

# Write Logs to Cloud Logging
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/logging.logWriter"
```

### 7.3 Grant Cloud Build Permissions

Cloud Build needs permissions to enable APIs (Step 0) and deploy services to Cloud Run.

> [!IMPORTANT]
> **Check which Service Account is running your build.**
> By default, Cloud Build uses `[PROJECT_NUMBER]@cloudbuild.gserviceaccount.com`. However, in some configurations, it may default to the **Compute Engine default service account** `[PROJECT_NUMBER]-compute@developer.gserviceaccount.com`.
> If your build fails with "Permission Denied", verify the service account in the build logs and grant the roles to that specific account.

```bash
# Get Project Details
PROJECT_ID=$(gcloud config get-value project)
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")

# Identify the Build Service Account (usually cloudbuild, but check logs)
BUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
# OR if your build uses compute SA:
# BUILD_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

# Enable services (Required for Step 0)
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$BUILD_SA" \
    --role="roles/serviceusage.serviceUsageAdmin"

# Deploy to Cloud Run
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$BUILD_SA" \
    --role="roles/run.admin"

# Act as Service Account
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$BUILD_SA" \
    --role="roles/iam.serviceAccountUser"

# Access Secrets
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$BUILD_SA" \
    --role="roles/secretmanager.secretAccessor"

# Write Logs (Required if using Compute SA for build execution)
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$BUILD_SA" \
    --role="roles/logging.logWriter"
```

---

## 8. Automated Deployment (Cloud Build)

This is the **recommended** deployment method. The `cloudbuild.yaml` file in the repo root handles everything:
1. Enables required APIs
2. Builds backend and frontend Docker images
3. Pushes images to Artifact Registry
4. Deploys both services to Cloud Run

### 8.1 First Deployment

```bash
# From the project root directory
gcloud builds submit \
    --config cloudbuild.yaml \
    --substitutions=\
_TAG=$(git rev-parse --short HEAD 2>/dev/null || echo "latest"),\
_DB_CONNECTION_NAME="my-elbulk-prod:us-central1:el-bulk-db",\
_GCS_BUCKET="elbulk-media-prod",\
_FRONTEND_ORIGIN="https://elbulk.com",\
_API_URL="https://api.elbulk.com",\
_GOOGLE_CLIENT_ID="your-google-client-id.apps.googleusercontent.com",\
_FACEBOOK_CLIENT_ID="your-facebook-app-id",\
_SMTP_HOST="smtp.sendgrid.net",\
_SMTP_PORT="587",\
_SMTP_USER="apikey",\
_SMTP_FROM="El Bulk <notices@elbulk.com>",\
_GA_ID="G-XXXXXXXXXX",\
_META_PIXEL_ID="1234567890",\
_GOOGLE_ADS_ID="AW-XXXXXXXXXX",\
_HOTJAR_ID="1234567",\
_BACKEND_INTERNAL_URL="https://el-bulk-backend-xxxxx-uc.a.run.app"
```

> [!IMPORTANT]
> The first deployment takes **5-10 minutes**. Subsequent deployments are faster due to Docker layer caching.

### 8.2 Verify Deployment

```bash
# Check backend
gcloud run services describe el-bulk-backend --region=us-central1 \
    --format="value(status.url)"

# Check frontend
gcloud run services describe el-bulk-frontend --region=us-central1 \
    --format="value(status.url)"
```

You should get URLs like:
```
https://el-bulk-backend-xxxxx-uc.a.run.app
https://el-bulk-frontend-xxxxx-uc.a.run.app
```

### 8.3 Test Health Endpoint

```bash
BACKEND_URL=$(gcloud run services describe el-bulk-backend \
    --region=us-central1 --format="value(status.url)")

curl -s "$BACKEND_URL/health" | jq
```

**Expected output:**
```json
{ "status": "ok" }
```

### 8.4 Subsequent Deployments

For future deploys, run the same `gcloud builds submit` command. Only the changed layers will rebuild.

---

## 9. Manual Deployment (Alternative)

Use this if you need to deploy individual services without Cloud Build.

### 9.1 Build & Push Backend

```bash
PROJECT_ID=$(gcloud config get-value project)
REGION="us-central1"

# Build
docker build -t $REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/backend:latest ./backend

# Push
docker push $REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/backend:latest

# Deploy
gcloud run deploy el-bulk-backend \
    --image=$REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/backend:latest \
    --region=$REGION \
    --platform=managed \
    --memory=512Mi \
    --cpu=1 \
    --min-instances=0 \
    --max-instances=5 \
    --add-cloudsql-instances=my-elbulk-prod:us-central1:el-bulk-db \
    --set-secrets="DATABASE_URL=ELBULK_DB_URL:latest,ENCRYPTION_KEY=ELBULK_ENCRYPTION_KEY:latest,JWT_SECRET=ELBULK_JWT_SECRET:latest,SMTP_PASS=ELBULK_SMTP_PASS:latest,GOOGLE_CLIENT_SECRET=ELBULK_GOOGLE_CLIENT_SECRET:latest,FACEBOOK_CLIENT_SECRET=ELBULK_FACEBOOK_CLIENT_SECRET:latest" \
    --set-env-vars="STORAGE_TYPE=gcp,GCP_BUCKET_NAME=elbulk-media-prod,APP_ENV=production,INSTANCE_CONNECTION_NAME=my-elbulk-prod:us-central1:el-bulk-db,DB_IAM_AUTH=true,GOOGLE_CLOUD_PROJECT=$PROJECT_ID,FRONTEND_ORIGIN=https://elbulk.com,GOOGLE_CLIENT_ID=your-google-id,FACEBOOK_CLIENT_ID=your-fb-id,SMTP_HOST=smtp.sendgrid.net,SMTP_PORT=587,SMTP_USER=apikey,SMTP_FROM=notices@elbulk.com,LOG_LEVEL=INFO,LOG_FORMAT=json" \
    --allow-unauthenticated
```

### 9.2 Build & Push Frontend

```bash
# Build with build-time args
docker build \
    --build-arg NEXT_PUBLIC_API_URL=https://api.elbulk.com \
    --build-arg NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX \
    --build-arg NEXT_PUBLIC_META_PIXEL_ID=1234567890 \
    --build-arg NEXT_PUBLIC_GOOGLE_ADS_ID=AW-XXXXXXXXXX \
    --build-arg NEXT_PUBLIC_HOTJAR_ID=1234567 \
    -t $REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/frontend:latest \
    ./frontend

# Push
docker push $REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/frontend:latest

# Deploy
gcloud run deploy el-bulk-frontend \
    --image=$REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/frontend:latest \
    --region=$REGION \
    --platform=managed \
    --memory=256Mi \
    --cpu=1 \
    --min-instances=0 \
    --max-instances=10 \
    --allow-unauthenticated
```

> [!WARNING]
> **Frontend build-time variables**: `NEXT_PUBLIC_*` variables are embedded into the frontend at **build time**. If you change them, you must **rebuild** the frontend image. They won't take effect just by updating Cloud Run env vars.

---

## 10. Database Initialization & Seeding

The backend automatically initializes the database schema on first startup (via `db.Initialize()`) using the files in `db/schema/`. No manual schema setup is needed.

### 10.1 First-Time Seed (Minimal — Production)

After the first deployment, run the seed to create the admin account and basic data:

```bash
# Get the backend service URL
BACKEND_URL=$(gcloud run services describe el-bulk-backend \
    --region=us-central1 --format="value(status.url)")

# Option A: Run seed via Cloud Run Jobs (recommended)
gcloud run jobs create el-bulk-seed \
    --image=$REGION-docker.pkg.dev/$PROJECT_ID/el-bulk-repo/backend:latest \
    --region=us-central1 \
    --set-cloudsql-instances=my-elbulk-prod:us-central1:el-bulk-db \
    --set-secrets="DATABASE_URL=ELBULK_DB_URL:latest,ENCRYPTION_KEY=ELBULK_ENCRYPTION_KEY:latest" \
    --set-env-vars="INSTANCE_CONNECTION_NAME=my-elbulk-prod:us-central1:el-bulk-db,DB_IAM_AUTH=true,ADMIN_USERNAME=admin,ADMIN_PASSWORD=YourSecurePassword!" \
    --command="go" \
    --args="run,./seed/,--mode=minimal"

gcloud run jobs execute el-bulk-seed --region=us-central1 --wait

# Option B: Connect via Cloud SQL Proxy and run locally
# (See section 19.2 for Cloud SQL Proxy setup)
cd backend && DATABASE_URL="..." go run ./seed/ --mode=minimal
```

### 10.2 Seeding Modes

| Mode | Use Case | Data Created |
|:---|:---|:---|
| `minimal` | **Production** | Admin account, TCGs, categories, settings, themes, translations, 1 reference product |
| `full` | **Development/Demo** | Everything in minimal + hundreds of products, 30+ customers, 100+ orders, bounties, offers |

> [!CAUTION]
> **Never run `--mode=full` on production.** It creates fake customers, orders, and test data.

---

## 11. Custom Domains & SSL

### Architecture Overview

El Bulk uses a **subdomain-based** architecture:

| Service | URL | Purpose |
|:---|:---|:---|
| Frontend | `https://elbulk.com` | Storefront & admin UI |
| Backend | `https://api.elbulk.com` | REST API |

### 11.1 Map Domains to Cloud Run

```bash
# Frontend domain
gcloud beta run domain-mappings create \
    --service=el-bulk-frontend \
    --domain=elbulk.com \
    --region=us-central1

# Backend (API) domain
gcloud beta run domain-mappings create \
    --service=el-bulk-backend \
    --domain=api.elbulk.com \
    --region=us-central1
```

### 11.2 Get DNS Records

After creating domain mappings, Cloud Run tells you what DNS records to create:

```bash
gcloud beta run domain-mappings describe \
    --domain=elbulk.com \
    --region=us-central1
```

**Note the `resourceRecords` field** — you'll need these values for Step 12.

> [!NOTE]
> SSL certificates are **automatically provisioned** by Google once DNS is configured. This can take 15-30 minutes after DNS propagates.

---

## 12. DNS Configuration

Configure these records at your domain registrar (Namecheap, GoDaddy, Cloudflare, etc.):

### 12.1 Required DNS Records

| Type | Name | Value | TTL |
|:---|:---|:---|:---|
| `A` | `@` (root) | `216.239.32.21`, `216.239.34.21`, `216.239.36.21`, `216.239.38.21` | 300 |
| `AAAA` | `@` (root) | `2001:4860:4802:32::15`, `2001:4860:4802:34::15`, `2001:4860:4802:36::15`, `2001:4860:4802:38::15` | 300 |
| `CNAME` | `api` | `ghs.googlehosted.com.` | 300 |
| `CNAME` | `www` | `ghs.googlehosted.com.` | 300 |

> [!IMPORTANT]
> The exact IP addresses come from Step 11.2. Cloud Run domain mappings provide specific IPs for root domains and CNAME targets for subdomains.

### 12.2 Verify DNS Propagation

```bash
# Check frontend
dig elbulk.com +short

# Check API
dig api.elbulk.com +short

# Verify SSL
curl -I https://elbulk.com
curl -I https://api.elbulk.com/health
```

DNS propagation can take **up to 48 hours** (usually 15-60 minutes).

---

## 13. OAuth Provider Setup

### 13.1 Google OAuth

1. Go to [Google Cloud Console → APIs & Services → Credentials](https://console.cloud.google.com/apis/credentials)
2. Click **Create Credentials** → **OAuth 2.0 Client ID**
3. Application type: **Web application**
4. Name: `El Bulk Production`
5. **Authorized JavaScript origins:**
   ```
   https://elbulk.com
   ```
6. **Authorized redirect URIs:**
   ```
   https://api.elbulk.com/api/auth/google/callback
   ```
7. Click **Create**
8. Copy the **Client ID** and **Client Secret**

**Where to use these values:**
- `Client ID` → `_GOOGLE_CLIENT_ID` in Cloud Build substitutions
- `Client Secret` → `ELBULK_GOOGLE_CLIENT_SECRET` in Secret Manager

### 13.2 Facebook OAuth (Meta for Developers)

> [!IMPORTANT]
> **Avoid Generic Emails**: Facebook often blocks account registration using `admin@`, `info@`, or `support@` emails. Use a personal-sounding email (e.g., `your.name@elbulk.com`) or simply use your personal Facebook account to manage the developer app.

> [!WARNING]
> **Real Name Policy**: Facebook requires personal accounts to use real human names. Do not name your administrative account "El Bulk Store" or it may be suspended. Use your personal profile and create a **Meta Business Portfolio** for the company.

#### Step-by-Step Configuration:

1.  Go to [Meta for Developers](https://developers.facebook.com/) and log in.
2.  Click **My Apps** -> **Create App**.
3.  Select **"Authenticate and request data from users with Facebook Login"**.
4.  Enter your **App Name** (e.g., "El Bulk Storefront").
5.  **Business Portfolio**: Select your newly created Meta Business Portfolio from the dropdown to ensure professional ownership.
6.  In the left sidebar, go to **App Settings** → **Basic**:
    - **Privacy Policy URL**: `https://elbulk.com/privacy` (Required to go live).
    - **User Data Deletion**: `https://elbulk.com/privacy` (You can point this to your privacy page or a dedicated instructions page).
7.  In the left sidebar, go to **Use Cases** (or "Add Product") → **Facebook Login** → **Settings**:
    - Under **Client OAuth Settings**, add these to **Valid OAuth Redirect URIs**:
      ```
      https://api.elbulk.com/api/auth/facebook/callback
      http://localhost:8080/api/auth/facebook/callback
      ```
8.  Go back to **App Settings** → **Basic**:
    - Copy the **App ID** and **App Secret**.

**Where to use these values:**
- `App ID` → `_FACEBOOK_CLIENT_ID` in Cloud Build substitutions.
- `App Secret` → `ELBULK_FACEBOOK_CLIENT_SECRET` in Secret Manager.

> [!CAUTION]
> Facebook requires your app to be in **Live mode** for customers to log in. You must provide a valid Privacy Policy URL and optionally submit for "App Review" if you request advanced permissions (though standard `email` and `public_profile` usually don't require review).

---

## 14. SMTP / Email Setup

El Bulk sends emails for newsletter notifications. Choose one of these providers:

### Option A: SendGrid (Recommended)

1. Create account at [sendgrid.com](https://sendgrid.com)
2. Go to **Settings → API Keys → Create API Key** (Full Access)
3. Configure:
   ```
   SMTP_HOST=smtp.sendgrid.net
   SMTP_PORT=587
   SMTP_USER=apikey
   SMTP_PASS=<your-api-key>
   SMTP_FROM=El Bulk <notices@elbulk.com>
   ```

### Option B: Brevo (Free Tier)

1. Create account at [brevo.com](https://brevo.com)
2. Go to **SMTP & API → SMTP Settings**
3. Configure:
   ```
   SMTP_HOST=smtp-relay.brevo.com
   SMTP_PORT=587
   SMTP_USER=<your-brevo-email>
   SMTP_PASS=<your-smtp-key>
   SMTP_FROM=El Bulk <notices@elbulk.com>
   ```

### Option C: Gmail App Password (Not Recommended for Production)

1. Enable 2FA on your Google Account
2. Go to [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
3. Generate an App Password
4. Configure:
   ```
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=your-email@gmail.com
   SMTP_PASS=<app-password>
   SMTP_FROM=El Bulk <your-email@gmail.com>
   ```

---

## 15. Monitoring & Alerting

### 15.1 Cloud Run Dashboard

View real-time metrics at:
`https://console.cloud.google.com/run?project=my-elbulk-prod`

Key metrics to watch:
- **Request count** — Traffic patterns
- **Request latency (p50, p95, p99)** — Performance
- **Container instance count** — Scaling behavior
- **Memory utilization** — Memory pressure

### 15.2 Cloud Logging

View structured backend logs:

```bash
# Stream live backend logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=el-bulk-backend" \
    --limit=50 \
    --format="table(timestamp, jsonPayload.message)"

# Search for errors
gcloud logging read "resource.type=cloud_run_revision AND severity=ERROR" \
    --limit=20
```

Or use the **Logs Explorer** in the GCP Console:
`https://console.cloud.google.com/logs/query?project=my-elbulk-prod`

### 15.3 Cloud SQL Monitoring

Check database health:
`https://console.cloud.google.com/sql/instances/el-bulk-db/overview?project=my-elbulk-prod`

Key metrics:
- **CPU utilization** — Upgrade tier if consistently >80%
- **Memory utilization** — Watch for OOMs
- **Disk usage** — Auto-increase is enabled
- **Active connections** — Default max is ~25 for f1-micro

### 15.4 Set Up Alerts (Recommended)

```bash
# Alert on high error rate
gcloud monitoring policies create --policy-from-file=/dev/stdin << 'EOF'
{
  "displayName": "El Bulk - High Error Rate",
  "conditions": [{
    "displayName": "Cloud Run 5xx errors > 5% for 5 minutes",
    "conditionThreshold": {
      "filter": "resource.type=\"cloud_run_revision\" AND metric.type=\"run.googleapis.com/request_count\" AND metric.labels.response_code_class=\"5xx\"",
      "comparison": "COMPARISON_GT",
      "thresholdValue": 10,
      "duration": "300s"
    }
  }],
  "notificationChannels": []
}
EOF
```

Consider creating alerts for:
- Backend 5xx error rate > 5%
- Request latency p95 > 2s
- Cloud SQL CPU > 80%
- Cloud SQL storage > 80%

---

## 16. Backups & Disaster Recovery

### 16.1 Automated Backups (Already Enabled)

Cloud SQL was provisioned with `--backup-start-time="03:00"`, meaning daily backups at 3 AM UTC.

Verify backup status:

```bash
gcloud sql backups list --instance=el-bulk-db
```

### 16.2 On-Demand Backup

Before any major migration or data change:

```bash
gcloud sql backups create --instance=el-bulk-db --description="Pre-migration backup"
```

### 16.3 Restore From Backup

```bash
# List available backups
gcloud sql backups list --instance=el-bulk-db

# Restore (replaces all data!)
gcloud sql backups restore <BACKUP_ID> --restore-instance=el-bulk-db
```

### 16.4 Export Database (Manual Snapshot)

```bash
# Export to Cloud Storage
gcloud sql export sql el-bulk-db gs://elbulk-media-prod/backups/elbulk-$(date +%Y%m%d).sql \
    --database=elbulk
```

---

## 17. Cost Optimization

### Estimated Monthly Costs

| Service | Tier | Estimated Cost |
|:---|:---|:---|
| Cloud SQL | `db-f1-micro` | ~$9/mo |
| Cloud Run (Backend) | 512Mi, 0-5 instances | ~$2-10/mo |
| Cloud Run (Frontend) | 256Mi, 0-10 instances | ~$2-15/mo |
| Cloud Storage | <1GB images | ~$0.02/mo |
| Secret Manager | 6 secrets | ~$0.06/mo |
| Artifact Registry | ~2GB images | ~$0.20/mo |
| **Total** | | **~$15-35/mo** |

### Cost Reduction Tips

1. **Set `min-instances=0`** — Allows scale-to-zero when idle (at the cost of cold starts)
2. **Use `db-f1-micro`** — Sufficient for <1000 daily users
3. **Enable `storage-auto-increase`** — Only pay for storage you use
4. **Set lifecycle rules** on Cloud Storage to auto-delete old images

---

## 18. Troubleshooting

### 18.1 "CORS Blocked" Errors

**Symptom:** Frontend shows `Access to fetch blocked by CORS policy`

**Fix:** Ensure `FRONTEND_ORIGIN` in the backend matches your frontend URL exactly:
```bash
# Check current value
gcloud run services describe el-bulk-backend --region=us-central1 \
    --format="table(spec.template.spec.containers.env)" | grep FRONTEND

# Update if wrong
gcloud run services update el-bulk-backend --region=us-central1 \
    --update-env-vars="FRONTEND_ORIGIN=https://elbulk.com"
```

### 18.2 "Database Connection Refused"

**Symptom:** Backend logs show `failed to connect to database`

**Checklist:**
1. Is `--add-cloudsql-instances` set in the deploy command?
2. Is `INSTANCE_CONNECTION_NAME` correct?
3. Does the service account have `roles/cloudsql.client`?
4. If using IAM auth, does the service account have `roles/cloudsql.instanceUser`?
5. Check the `DATABASE_URL` secret format matches what `db.go` expects

```bash
# Check the secret value (careful — this is sensitive)
gcloud secrets versions access latest --secret=ELBULK_DB_URL
```

### 18.3 "Schema Initialization Failed"

**Symptom:** Tables don't exist, `relation "product" does not exist`

**Cause:** The `db/schema/` directory wasn't copied into the Docker image.

**Fix:** Verify the Dockerfile contains:
```dockerfile
COPY --from=builder /app/db/schema ./db/schema
COPY --from=builder /app/db/migrations ./db/migrations
```

Then rebuild and redeploy.

### 18.4 OAuth Login Redirects to Wrong URL

**Symptom:** After Google/Facebook login, user is redirected to `localhost` or wrong domain

**Fix:** Ensure `FRONTEND_ORIGIN` is set correctly in the backend, and the OAuth redirect URIs match exactly:
- Google: `https://api.elbulk.com/api/auth/google/callback`
- Facebook: `https://api.elbulk.com/api/auth/facebook/callback`

### 18.5 "Cold Start" Latency

**Symptom:** First request after idle period takes 3-8 seconds

**Fix:** Set minimum instances to 1 (increases cost ~$10/mo):
```bash
gcloud run services update el-bulk-backend --region=us-central1 --min-instances=1
```

### 18.6 Images Not Loading

**Symptom:** Product images show broken image placeholder

**Checklist:**
1. Is `STORAGE_TYPE=gcp` set?
2. Is `GCP_BUCKET_NAME` correct?
3. Is the bucket publicly readable? (Step 5.2)
4. Is the bucket name in `next.config.ts` → `remotePatterns`?
   - Add: `{ protocol: 'https', hostname: 'storage.googleapis.com' }`

### 18.7 Viewing Logs

```bash
# Backend logs (last 100 entries)
gcloud logging read \
    "resource.type=cloud_run_revision AND resource.labels.service_name=el-bulk-backend" \
    --limit=100 --format="table(timestamp,severity,jsonPayload.message)"

# Frontend logs
gcloud logging read \
    "resource.type=cloud_run_revision AND resource.labels.service_name=el-bulk-frontend" \
    --limit=50
```

---

## 19. Maintenance Operations

### 19.1 Update Environment Variables

```bash
# Add or update a single variable
gcloud run services update el-bulk-backend --region=us-central1 \
    --update-env-vars="LOG_LEVEL=DEBUG"

# Remove a variable
gcloud run services update el-bulk-backend --region=us-central1 \
    --remove-env-vars="DEPRECATED_VAR"
```

### 19.2 Connect to Production Database

Use the Cloud SQL Proxy to connect from your local machine:

```bash
# Install proxy
gcloud components install cloud-sql-proxy

# Start proxy (runs in background)
cloud-sql-proxy my-elbulk-prod:us-central1:el-bulk-db &

# Connect with psql
psql "host=127.0.0.1 port=5432 dbname=elbulk user=postgres"
```

### 19.3 Rollback a Deployment

```bash
# List revisions
gcloud run revisions list --service=el-bulk-backend --region=us-central1

# Route 100% traffic to a previous revision
gcloud run services update-traffic el-bulk-backend \
    --region=us-central1 \
    --to-revisions=el-bulk-backend-00042-abc=100
```

### 19.4 Scale Configuration

```bash
# Increase max instances during peak traffic
gcloud run services update el-bulk-backend --region=us-central1 \
    --max-instances=10

# Increase memory for heavy operations (bulk imports)
gcloud run services update el-bulk-backend --region=us-central1 \
    --memory=1Gi
```

### 19.5 Trigger Manual Price Refresh

```bash
BACKEND_URL=$(gcloud run services describe el-bulk-backend \
    --region=us-central1 --format="value(status.url)")

curl -X POST "$BACKEND_URL/api/admin/prices/refresh" \
    -H "Cookie: admin_token=<your-admin-jwt>"
```

### 19.6 Change Admin Log Level at Runtime

```bash
# Set to DEBUG for troubleshooting
curl -X PUT "$BACKEND_URL/api/admin/logs/level" \
    -H "Cookie: admin_token=<your-admin-jwt>" \
    -H "Content-Type: application/json" \
    -d '{"level": "DEBUG"}'

# Set back to INFO
curl -X PUT "$BACKEND_URL/api/admin/logs/level" \
    -H "Cookie: admin_token=<your-admin-jwt>" \
    -H "Content-Type: application/json" \
    -d '{"level": "INFO"}'
```

---

## 20. Environment Variable Reference

### Backend — Secrets (Secret Manager)

| Secret | Required | Description |
|:---|:---:|:---|
| `ELBULK_DB_URL` | ✅ | PostgreSQL connection string |
| `ELBULK_JWT_SECRET` | ✅ | JWT signing key (≥32 chars) |
| `ELBULK_ENCRYPTION_KEY` | ✅ | AES-256 key for PII (exactly 32 hex chars) |
| `ELBULK_SMTP_PASS` | ⚠️ | SMTP password (required for newsletters) |
| `ELBULK_GOOGLE_CLIENT_SECRET` | ⚠️ | Google OAuth secret (required for Google login) |
| `ELBULK_FACEBOOK_CLIENT_SECRET` | ❌ | Facebook OAuth secret (optional) |

### Backend — Environment Variables (Cloud Run)

| Variable | Required | Default | Description |
|:---|:---:|:---|:---|
| `APP_ENV` | ✅ | `development` | Set to `production` for prod security |
| `FRONTEND_ORIGIN` | ✅ | — | CORS allowed origin (e.g., `https://elbulk.com`) |
| `INSTANCE_CONNECTION_NAME` | ✅ | — | Cloud SQL instance connection name |
| `DB_IAM_AUTH` | ✅ | `false` | Enable IAM-based DB auth |
| `STORAGE_TYPE` | ✅ | — | `gcp` for Cloud Storage |
| `GCP_BUCKET_NAME` | ✅ | — | Cloud Storage bucket name |
| `GOOGLE_CLIENT_ID` | ⚠️ | — | Google OAuth client ID |
| `FACEBOOK_CLIENT_ID` | ❌ | — | Facebook OAuth app ID |
| `SMTP_HOST` | ⚠️ | — | SMTP server hostname |
| `SMTP_PORT` | ⚠️ | `587` | SMTP server port |
| `SMTP_USER` | ⚠️ | — | SMTP username |
| `SMTP_FROM` | ⚠️ | — | Email sender address |
| `LOG_LEVEL` | ❌ | `INFO` | `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `LOG_FORMAT` | ❌ | `text` | `text` or `json` (use `json` for GCP) |
| `DB_MAX_OPEN_CONNS` | ❌ | `25` | Max open database connections |
| `DB_MAX_IDLE_CONNS` | ❌ | `5` | Max idle database connections |
| `POKEMON_TCG_API_KEY` | ❌ | — | Pokémon TCG API key for card lookup |
| `GOOGLE_CLOUD_PROJECT` | ❌ | — | Auto-set by Cloud Run; for Cloud Trace |

### Frontend — Build-Time Variables (Docker Build Args)

| Variable | Required | Description |
|:---|:---:|:---|
| `NEXT_PUBLIC_API_URL` | ✅ | Backend API URL (e.g., `https://api.elbulk.com`) |
| `NEXT_PUBLIC_GA_ID` | ❌ | Google Analytics 4 Measurement ID |
| `NEXT_PUBLIC_META_PIXEL_ID` | ❌ | Facebook/Meta Pixel ID |
| `NEXT_PUBLIC_GOOGLE_ADS_ID` | ❌ | Google Ads Conversion ID |
| `NEXT_PUBLIC_HOTJAR_ID` | ❌ | Hotjar Site ID |
| `INTERNAL_API_URL` | ✅ | Backend internal URL for SSR rewrites (provided at runtime via `_BACKEND_INTERNAL_URL`) |

> [!IMPORTANT]
> ✅ = Required for the service to function
> ⚠️ = Required for specific features (login, email)
> ❌ = Optional, service works without it

---

## Deployment Checklist

Use this checklist for your first production deployment:

- [ ] GCP Project created with billing enabled
- [ ] All required APIs enabled (Step 2.3)
- [ ] Artifact Registry repo created (Step 3)
- [ ] Cloud SQL instance created and database `elbulk` exists (Step 4)
- [ ] IAM authentication configured (Step 4.5) *or* password-based user created
- [ ] Cloud Storage bucket created and made public (Step 5)
- [ ] All 6 secrets created in Secret Manager (Step 6)
- [ ] IAM roles granted to Cloud Run service account (Step 7.2)
- [ ] IAM roles granted to Cloud Build service account (Step 7.3)
- [ ] First `gcloud builds submit` completed successfully (Step 8)
- [ ] Health endpoint returns `{"status":"ok"}` (Step 8.3)
- [ ] Database seeded with `--mode=minimal` (Step 10)
- [ ] Can log into admin dashboard with seeded credentials
- [ ] Custom domains mapped (Step 11)
- [ ] DNS records configured and propagated (Step 12)
- [ ] SSL certificates provisioned (auto by Google)
- [ ] Google OAuth configured with correct redirect URI (Step 13.1)
- [ ] CORS working (frontend can call backend API)
- [ ] Images uploading and displaying correctly
- [ ] Newsletter subscription working (if SMTP configured)

---

## 21. CI/CD: Automated Triggers

To automate deployments when the `main` branch is updated while allowing you to work on local branches (like `dv`) without triggering builds:

### 21.1 Create a Trigger
1.  Go to **Cloud Build > Triggers** in the [GCP Console](https://console.cloud.google.com/cloud-build/triggers).
2.  Click **Create Trigger**.
3.  **Name**: `deploy-on-main`
4.  **Event**: Select **Push to a branch**.
5.  **Repository**: Connect your GitHub/GitLab/Bitbucket repository.
6.  **Branch**: Set the filter to `^main$` (uses regex to match only the main branch).
7.  **Configuration**: Select **Cloud Build configuration file (yaml/json)** and set the path to `cloudbuild.yaml`.

### 21.2 Configure Substitutions
Automated triggers require you to define the variables usually passed via the `--substitutions` flag in the UI:

1.  In the Trigger settings, scroll down to **Substitution variables**.
2.  Click **Add Variable** for each required substitution in `cloudbuild.yaml`.
3.  **Keys** must start with an underscore (e.g., `_DB_CONNECTION_NAME`).
4.  **Values** should be the production-specific strings.

**Common Substitutions to add:**
- `_DB_CONNECTION_NAME`
- `_GCS_BUCKET`
- `_API_URL`
- `_FRONTEND_ORIGIN`
- `_GOOGLE_CLIENT_ID`
- `_BACKEND_INTERNAL_URL`

This setup ensures that only merges to `main` go live, while your local development workflow remains isolated.
