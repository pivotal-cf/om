#!/bin/bash
# Test script to verify proxy authentication

set -e

echo "=== Testing Proxy Authentication ==="

# Test 1: Get Kerberos ticket
echo "1. Getting Kerberos ticket..."
echo "testpass123" | kinit testuser@EXAMPLE.COM

if [ $? -eq 0 ]; then
    echo "✓ Kerberos ticket obtained"
    klist
else
    echo "✗ Failed to get Kerberos ticket"
    exit 1
fi

# Test 2: Test proxy with curl
echo ""
echo "2. Testing proxy with curl..."
export HTTP_PROXY=http://proxy.example.com:3128
export HTTPS_PROXY=http://proxy.example.com:3128

# Test HTTP request through proxy
response=$(curl -s -o /dev/null -w "%{http_code}" --proxy $HTTP_PROXY http://www.example.com)

if [ "$response" = "200" ] || [ "$response" = "301" ] || [ "$response" = "302" ]; then
    echo "✓ Proxy authentication successful (HTTP $response)"
else
    echo "✗ Proxy authentication failed (HTTP $response)"
    exit 1
fi

# Test HTTPS request through proxy
response=$(curl -s -o /dev/null -w "%{http_code}" --proxy $HTTPS_PROXY https://www.example.com)

if [ "$response" = "200" ] || [ "$response" = "301" ] || [ "$response" = "302" ]; then
    echo "✓ HTTPS proxy authentication successful (HTTP $response)"
else
    echo "✗ HTTPS proxy authentication failed (HTTP $response)"
    exit 1
fi

echo ""
echo "=== All Tests Passed ==="

