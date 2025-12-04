#!/bin/bash
set -e

echo "=== Setting up OM Proxy Authentication Test Infrastructure ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}Error: docker-compose is not installed. Please install docker-compose.${NC}"
    exit 1
fi

# Use docker compose (newer) or docker-compose (older)
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo -e "${GREEN}✓ Docker is running${NC}"

# Start services
echo "Starting Docker containers..."
cd "$(dirname "$0")"
$DOCKER_COMPOSE up -d

echo "Waiting for services to be ready..."
sleep 10

# Wait for KDC to be healthy
echo "Waiting for KDC to be ready..."
timeout=60
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if docker exec om-test-kdc kadmin.local -q "list_principals" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ KDC is ready${NC}"
        break
    fi
    echo "Waiting for KDC... ($elapsed/$timeout seconds)"
    sleep 2
    elapsed=$((elapsed + 2))
done

if [ $elapsed -ge $timeout ]; then
    echo -e "${RED}Error: KDC did not become ready in time${NC}"
    exit 1
fi

# Create test user principal
echo "Creating test user principal..."
docker exec om-test-kdc kadmin.local -q "addprinc -pw testpass123 testuser@EXAMPLE.COM" || true

# Create proxy service principal
echo "Creating proxy service principal..."
docker exec om-test-kdc kadmin.local -q "addprinc -randkey HTTP/proxy.example.com@EXAMPLE.COM" || true

# Create keytab for proxy
echo "Creating keytab for proxy..."
docker exec om-test-kdc kadmin.local -q "ktadd -k /tmp/krb5.keytab HTTP/proxy.example.com@EXAMPLE.COM" || true
docker exec om-test-kdc cat /tmp/krb5.keytab > squid/krb5.keytab || true

# Set permissions
chmod 644 squid/krb5.keytab 2>/dev/null || true
