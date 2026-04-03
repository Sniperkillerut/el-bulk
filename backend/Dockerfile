FROM golang:1.25-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git curl
# Install Air via fixed version binary for stability
RUN curl -sSfL https://github.com/air-verse/air/releases/download/v1.61.7/air_1.61.7_linux_amd64.tar.gz | tar -xz -C /usr/bin/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT ["air"]

FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build deps
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /el-bulk-api .

# Final minimal image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /el-bulk-api .
COPY db/schema.sql ./db/

USER appuser

EXPOSE 8080

CMD ["./el-bulk-api"]
