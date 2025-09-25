#!/bin/bash

# SESC Backend Deployment Script for Camal
# This script deploys the SESC application using Camal

set -e

echo "🚀 Starting SESC Backend deployment with Camal..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required environment variables are set
check_env_vars() {
    print_status "Checking environment variables..."
    
    required_vars=(
        "POSTGRES_PASSWORD"
        "MINIO_ROOT_USER"
        "MINIO_ROOT_PASSWORD"
        "SSH_HOST"
        "SSH_USER"
    )
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            print_error "Required environment variable $var is not set"
            exit 1
        fi
    done
    
    print_success "All required environment variables are set"
}

# Create production docker-compose file
create_docker_compose() {
    print_status "Creating production docker-compose.yml..."
    
    cat > docker-compose.prod.yml << 'COMPOSE_EOF'
version: '3.8'

networks:
  traefik:
    external: true
  internal:
    external: false

volumes:
  postgres:
  minio-data:
  traefik-certificates:

services:
  traefik:
    image: traefik:v3.0
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    networks:
      - traefik
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik-certificates:/certificates
    command:
      - --api.dashboard=true
      - --api.insecure=false
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --providers.docker.network=traefik
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.email=admin@sesc.online
      - --certificatesresolvers.letsencrypt.acme.storage=/certificates/acme.json
      - --certificatesresolvers.letsencrypt.acme.httpchallenge=true
      - --certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web
      - --entrypoints.web.http.redirections.entrypoint.to=websecure
      - --entrypoints.web.http.redirections.entrypoint.scheme=https
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.traefik.rule=Host(`traefik.sesc.online`)"
      - "traefik.http.routers.traefik.entrypoints=websecure"
      - "traefik.http.routers.traefik.tls.certresolver=letsencrypt"
      - "traefik.http.routers.traefik.service=api@internal"

  postgres:
    image: postgres:15
    container_name: postgres
    restart: unless-stopped
    networks:
      - internal
    environment:
      POSTGRES_DB: sesc
      POSTGRES_USER: sesc
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "sesc"]
      interval: 10s
      retries: 5
      start_period: 30s
      timeout: 10s

  minio:
    image: minio/minio:latest
    container_name: minio
    restart: unless-stopped
    networks:
      - internal
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    volumes:
      - minio-data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 10s
      retries: 5
      start_period: 30s
      timeout: 10s

  backend:
    image: ghcr.io/kozlov-ma/sesc-backend/backend:latest
    container_name: sesc-backend
    restart: unless-stopped
    networks:
      - traefik
      - internal
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_healthy
    environment:
      SESC_DATABASE_ADDRESS: postgres://sesc:${POSTGRES_PASSWORD}@postgres:5432/sesc?sslmode=disable
      SESC_MINIO_ENDPOINT: minio:9000
      SESC_MINIO_ACCESS_KEY: ${MINIO_ROOT_USER}
      SESC_MINIO_SECRET_KEY: ${MINIO_ROOT_PASSWORD}
      SESC_MINIO_USE_SSL: "false"
      SESC_MINIO_BUCKET_NAME: sesc-files
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.backend.rule=Host(`api.sesc.online`)"
      - "traefik.http.routers.backend.entrypoints=websecure"
      - "traefik.http.routers.backend.tls.certresolver=letsencrypt"
      - "traefik.http.services.backend.loadbalancer.server.port=8080"

  frontend:
    image: ghcr.io/kozlov-ma/sesc-backend/frontend:latest
    container_name: sesc-frontend
    restart: unless-stopped
    networks:
      - traefik
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`sesc.online`)"
      - "traefik.http.routers.frontend.entrypoints=websecure"
      - "traefik.http.routers.frontend.tls.certresolver=letsencrypt"
      - "traefik.http.services.frontend.loadbalancer.server.port=3000"
COMPOSE_EOF

    print_success "Production docker-compose.yml created"
}

# Deploy to server
deploy_to_server() {
    print_status "Deploying to server $SSH_HOST..."
    
    # Copy docker-compose file to server
    scp -o StrictHostKeyChecking=no docker-compose.prod.yml $SSH_USER@$SSH_HOST:~/docker-compose.yml
    
    # Deploy on server
    ssh -o StrictHostKeyChecking=no $SSH_USER@$SSH_HOST << EOF
        set -e
        
        echo "🔧 Setting up environment variables..."
        export POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
        export MINIO_ROOT_USER="$MINIO_ROOT_USER"
        export MINIO_ROOT_PASSWORD="$MINIO_ROOT_PASSWORD"
        
        echo "🐳 Logging into GitHub Container Registry..."
        echo "$GITHUB_TOKEN" | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin
        
        echo "🌐 Creating traefik network..."
        docker network create traefik 2>/dev/null || true
        
        echo "📥 Pulling latest images..."
        docker pull ghcr.io/kozlov-ma/sesc-backend/backend:latest
        docker pull ghcr.io/kozlov-ma/sesc-backend/frontend:latest
        
        echo "🛑 Stopping existing containers..."
        docker-compose down || true
        
        echo "🚀 Starting services..."
        docker-compose up -d
        
        echo "⏳ Waiting for services to start..."
        sleep 30
        
        echo "📊 Checking service status..."
        docker-compose ps
        
        echo "🔄 Restarting Traefik for service discovery..."
        docker-compose restart traefik
        sleep 15
        
        echo "🧹 Cleaning up unused images..."
        docker image prune -f
        
        echo "✅ Deployment completed successfully!"
        echo "🌐 Services available at:"
        echo "   - Frontend: https://sesc.online"
        echo "   - Backend API: https://api.sesc.online"
        echo "   - Traefik Dashboard: https://traefik.sesc.online"
EOF

    print_success "Deployment completed successfully!"
}

# Verify deployment
verify_deployment() {
    print_status "Verifying deployment..."
    
    ssh -o StrictHostKeyChecking=no $SSH_USER@$SSH_HOST << EOF
        export POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
        export MINIO_ROOT_USER="$MINIO_ROOT_USER"
        export MINIO_ROOT_PASSWORD="$MINIO_ROOT_PASSWORD"
        
        echo "⏳ Waiting for services to be healthy..."
        sleep 30
        
        echo "🔍 Checking container status..."
        if ! docker-compose ps | grep -q "Up"; then
            echo "❌ Deployment failed - containers not running"
            docker-compose logs
            exit 1
        fi
        
        echo "📊 Final service status:"
        docker-compose ps
        
        echo "✅ Deployment verification successful!"
EOF

    print_success "Deployment verification completed!"
}

# Main deployment function
main() {
    print_status "Starting SESC Backend deployment with Camal..."
    
    check_env_vars
    create_docker_compose
    deploy_to_server
    verify_deployment
    
    print_success "🎉 SESC Backend deployment completed successfully!"
    print_status "Your application is now available at:"
    print_status "  - Frontend: https://sesc.online"
    print_status "  - Backend API: https://api.sesc.online"
    print_status "  - Traefik Dashboard: https://traefik.sesc.online"
}

# Run main function
main "$@"
