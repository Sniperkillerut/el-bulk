FROM golang:1.23-alpine AS builder

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

WORKDIR /app
COPY --from=builder /el-bulk-api .
COPY db/schema.sql ./db/

EXPOSE 8080

CMD ["./el-bulk-api"]
