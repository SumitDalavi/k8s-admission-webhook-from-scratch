#!/bin/bash
set -e

echo "Generating local test certs..."
mkdir -p /tmp/webhook-certs
openssl req -x509 -newkey rsa:2048 -keyout /tmp/webhook-certs/tls.key -out /tmp/webhook-certs/tls.crt -days 1 -nodes -subj "/CN=localhost"

echo "Building webhook..."
go build -o webhook ./cmd/webhook

echo "Starting webhook server in background..."
# We patch main.go arguments or just run it. Wait, main.go hardcodes /etc/webhook/certs.
# We will use docker to mount the certs.
docker build -t admission-webhook:test .
docker run -d --name webhook-test -p 8443:8443 -v /tmp/webhook-certs:/etc/webhook/certs admission-webhook:test

sleep 2

echo "Running integration tests against localhost:8443..."
# Send mock AdmissionReview request
curl -k -v -X POST "https://localhost:8443/validate" \
  -H "Content-Type: application/json" \
  -d '{"apiVersion":"admission.k8s.io/v1","kind":"AdmissionReview","request":{"uid":"123","object":{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","labels":{"secure":"true"}}}}}'

echo "Cleaning up..."
docker stop webhook-test
docker rm webhook-test
