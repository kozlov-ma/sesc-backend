#!/bin/bash

# Helper script to fix MinIO proxy configuration
# Use this if MinIO health check fails after reboot

set -e

SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_sesc}"
SSH_USER="${SSH_USER:-kanyewest}"
SSH_HOST="${SSH_HOST:-158.160.165.141}"

echo "🔧 Fixing MinIO proxy configuration on $SSH_HOST..."

ssh -i "$SSH_KEY" "$SSH_USER@$SSH_HOST" "docker exec kamal-proxy kamal-proxy deploy sesc-backend-minio \
  --target='sesc-backend-minio:9000' \
  --host='s3.sesc-fiit.ru' \
  --tls \
  --deploy-timeout='60s' \
  --health-check-path='/minio/health/live' \
  --health-check-interval='1s'"

echo "✅ MinIO proxy configured successfully!"
echo ""
echo "Verify with:"
echo "  ssh -i $SSH_KEY $SSH_USER@$SSH_HOST 'docker exec kamal-proxy kamal-proxy list'"

