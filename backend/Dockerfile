## =======================================
## Stage: dev (local development with Air)
## =======================================
FROM golang:1.26-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git curl
# Install Air via fixed version binary for stability
RUN curl -sSfL https://github.com/air-verse/air/releases/download/v1.61.7/air_1.61.7_linux_amd64.tar.gz | tar -xz -C /usr/bin/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT ["air"]

## =======================================
## Stage: builder (compile binary)
## =======================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build deps
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG APP_VERSION=0.0.0-dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${APP_VERSION}" -o /el-bulk-api .

# Build seed binary separately so it can be run via Cloud Run Jobs
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /el-bulk-seed ./seed/

# Build backfill script
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /backfill-treatments ./scripts/backfill_treatments.go

## =======================================
## Stage: runner (minimal production image)
## =======================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /el-bulk-api .
COPY --from=builder /el-bulk-seed .
COPY --from=builder /backfill-treatments .
COPY --from=builder /app/db/schema ./db/schema
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/seed/data ./seed/data

# Optimizations for Cloud Run / Serverless memory limits
# GOGC: Trigger GC more aggressively (default 100) to reduce memory peaks.
# GOMEMLIMIT: Soft memory limit. Set to ~90% of your Cloud Run container RAM.
#   - 512Mi container → 450MiB
#   - 1Gi container  → 900MiB
ENV GOGC=50
ENV GOMEMLIMIT=450MiB

USER appuser

EXPOSE 8080

# Health check for docker-compose and orchestrators
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./el-bulk-api"]
