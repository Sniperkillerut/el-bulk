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
- **Status**: [ ] Pending

## 📊 Priority 3: Long-term Scalability

### 3.1 Background Task Processing
- **Goal**: Move heavy work (like image processing or price updates) to background workers.
- **Impact**: Improved API responsiveness for users.
- **Status**: [ ] Pending
