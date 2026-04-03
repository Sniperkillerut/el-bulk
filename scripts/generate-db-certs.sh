#!/bin/bash

# Configuration
CERT_DIR="./certs"
DAYS=365
DB_HOST="db"

# Create certs directory
mkdir -p "$CERT_DIR"

echo "Generating Root CA..."
openssl req -new -x509 -days "$DAYS" -nodes -out "$CERT_DIR/ca.crt" -keyout "$CERT_DIR/ca.key" -subj "/CN=Root CA"

echo "Generating Server Key and Certificate for $DB_HOST..."
openssl req -new -nodes -out "$CERT_DIR/server.csr" -keyout "$CERT_DIR/server.key" -subj "/CN=$DB_HOST" \
    -addext "subjectAltName = DNS:$DB_HOST, DNS:localhost, IP:127.0.0.1"
openssl x509 -req -in "$CERT_DIR/server.csr" -days "$DAYS" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial -out "$CERT_DIR/server.crt" \
    -copy_extensions copyall

echo "Generating Client Key and Certificate for backend..."
openssl req -new -nodes -out "$CERT_DIR/client.csr" -keyout "$CERT_DIR/client.key" -subj "/CN=backend" \
    -addext "subjectAltName = DNS:backend, DNS:localhost, IP:127.0.0.1"
openssl x509 -req -in "$CERT_DIR/client.csr" -days "$DAYS" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial -out "$CERT_DIR/client.crt" \
    -copy_extensions copyall

# Set permissions for keys as required by Postgres
chmod 0600 "$CERT_DIR/server.key" "$CERT_DIR/client.key"

# Clean up CSRs
rm "$CERT_DIR/server.csr" "$CERT_DIR/client.csr"

echo "Certificates generated in $CERT_DIR"
ls -l "$CERT_DIR"
