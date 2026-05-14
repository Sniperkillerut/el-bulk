# Backend Optimization Roadmap

This document tracks planned performance and stability improvements for the `el-bulk` Go backend, specifically tuned for high-traffic environments and `db-f1-micro` database constraints.

## 🚀 Priority 1: Low-Hanging Fruit (Immediate Stability)

### 1.1 Gzip Response Compression
- **Goal**: Reduce payload sizes for large product lists and facet data.
- **Impact**: ~70% reduction in egress bandwidth, faster UI load times.
- **Implementation**: Add `chi.Middleware.Compress` to `main.go`.
- **Status**: [x] Completed

### 1.2 Strict Connection Pooling
- **Goal**: Prevent the `53300: too many connections` error on Micro instances.
- **Impact**: High stability for the database under concurrent load.
- **Implementation**: Enforce `SetMaxOpenConns` and `SetMaxIdleConns` via environment variables.
- **Status**: [x] Completed

## 🧠 Priority 2: Memory & Query Efficiency

### 2.1 In-Memory Caching (Static Entities)
- **Goal**: Cache `Settings`, `TCGs`, `Categories`, and `Translations`.
- **Impact**: Significant reduction in DB query volume (20-30%), sub-millisecond response times for core data.
- **Status**: [x] Completed

### 2.2 Database Indexing Audit
- **Goal**: Ensure columns used in `GetFacets` and `ListWithFilters` have optimal composite indexes.
- **Impact**: Lower DB CPU usage and faster search results.
- **Status**: [x] Completed

### 2.3 Search Result Caching
- **Goal**: Cache the first few pages of common search queries (e.g., TCG front pages).
- **Impact**: Dramatically reduces DB load during high-traffic browsing sessions.
- **Status**: [x] Completed

## 📊 Priority 3: Long-term Scalability

### 3.1 Background Worker Pool
- **Goal**: Implement a managed worker pool (e.g., using a library like `River` or a robust internal queue).
- **Impact**: Offloads CSV imports, image processing, and price syncing from the API request lifecycle.
- **Status**: [ ] Pending

### 3.2 Asset CDN Migration (Cloudflare R2 / GCP Buckets)
- **Goal**: Offload image serving from the Go backend to a dedicated CDN.
- **Impact**: Dramatically reduces server egress bandwidth and improves image load times globally.
- **Status**: [ ] Pending

### 3.3 Advanced Monitoring & Alerting
- **Goal**: Integrate Sentry and Google Cloud Monitoring (Cloud Trace/Metrics).
- **Impact**: Real-time visibility into database connection saturation and automatic alerts for API errors.
- **Status**: [ ] Pending

### 3.4 API Versioning & Breaking Change Shield
- **Goal**: Transition to `/api/v1` and implement strict schema validation for bulk imports.
- **Impact**: Prevents invalid data from entering the database during large-scale imports.
- **Status**: [ ] Pending
