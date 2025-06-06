#!/bin/bash

# SSL Certificate Initialization Script for SESC
# This script initializes Let's Encrypt certificates for sesc.online and api.sesc.online

set -e

# Configuration
DOMAINS=(sesc.online api.sesc.online)
EMAIL="admin@sesc.online"  # Replace with your email
STAGING=0  # Set to 1 for staging certificates (for testing)

# Paths
DOCKER_COMPOSE_FILE="/opt/sesc/docker-compose.simple.yml"
CERTBOT_DATA_PATH="/opt/sesc/certbot"
NGINX_CONF_PATH="/opt/sesc/nginx"

echo "🚀 Starting SSL certificate initialization..."

# Create required directories
mkdir -p "$CERTBOT_DATA_PATH/conf"
mkdir -p "$CERTBOT_DATA_PATH/www"
mkdir -p "$CERTBOT_DATA_PATH/logs"

# Check if certificates already exist
if [ -d "$CERTBOT_DATA_PATH/conf/live/sesc.online" ]; then
    echo "⚠️  Certificates already exist. Do you want to recreate them? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "🛑 Exiting without changes."
        exit 0
    fi
    echo "🗑️  Removing existing certificates..."
    rm -rf "$CERTBOT_DATA_PATH/conf/live"
    rm -rf "$CERTBOT_DATA_PATH/conf/archive"
    rm -rf "$CERTBOT_DATA_PATH/conf/renewal"
fi

# Download recommended TLS parameters
echo "📥 Downloading recommended TLS parameters..."
if [ ! -f "$CERTBOT_DATA_PATH/conf/options-ssl-nginx.conf" ]; then
    curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$CERTBOT_DATA_PATH/conf/options-ssl-nginx.conf"
fi

if [ ! -f "$CERTBOT_DATA_PATH/conf/ssl-dhparams.pem" ]; then
    curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$CERTBOT_DATA_PATH/conf/ssl-dhparams.pem"
fi

# Create dummy certificates for nginx to start
echo "🔧 Creating dummy certificates..."
for domain in "${DOMAINS[@]}"; do
    path="/etc/letsencrypt/live/$domain"
    mkdir -p "$CERTBOT_DATA_PATH/conf/live/$domain"
    
    # Generate dummy certificate
    openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
        -keyout "$CERTBOT_DATA_PATH/conf/live/$domain/privkey.pem" \
        -out "$CERTBOT_DATA_PATH/conf/live/$domain/fullchain.pem" \
        -subj "/CN=localhost" >/dev/null 2>&1
done

# Start nginx with dummy certificates
echo "🌐 Starting nginx with dummy certificates..."
# First start infrastructure services
docker-compose -f "$DOCKER_COMPOSE_FILE" up -d postgres minio
sleep 20
# Then start application services
docker-compose -f "$DOCKER_COMPOSE_FILE" up -d backend frontend
sleep 15
# Finally start nginx
docker-compose -f "$DOCKER_COMPOSE_FILE" up -d nginx

# Wait for nginx to start
echo "⏳ Waiting for nginx to start..."
sleep 10

# Request real certificates
echo "🔒 Requesting Let's Encrypt certificates..."

# Set staging flag if needed
staging_arg=""
if [ $STAGING -eq 1 ]; then
    staging_arg="--staging"
    echo "⚠️  Using staging environment for testing"
fi

for domain in "${DOMAINS[@]}"; do
    echo "📝 Requesting certificate for $domain..."
    
    # Remove dummy certificate
    rm -rf "$CERTBOT_DATA_PATH/conf/live/$domain"
    
    # Request certificate
    docker run --rm \
        -v "$CERTBOT_DATA_PATH/conf:/etc/letsencrypt" \
        -v "$CERTBOT_DATA_PATH/www:/var/www/certbot" \
        -v "$CERTBOT_DATA_PATH/logs:/var/log/letsencrypt" \
        certbot/certbot \
        certonly \
        --webroot \
        --webroot-path=/var/www/certbot \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        $staging_arg \
        -d "$domain"
    
    if [ $? -eq 0 ]; then
        echo "✅ Certificate for $domain obtained successfully"
    else
        echo "❌ Failed to obtain certificate for $domain"
        exit 1
    fi
done

# Reload nginx with real certificates
echo "🔄 Reloading nginx with real certificates..."
docker-compose -f "$DOCKER_COMPOSE_FILE" exec nginx nginx -s reload

# Set up auto-renewal
echo "⚙️  Setting up auto-renewal..."
crontab -l 2>/dev/null | grep -v "certbot renew" | crontab -
(crontab -l 2>/dev/null; echo "0 12 * * * cd /opt/sesc && docker run --rm -v /opt/sesc/certbot/conf:/etc/letsencrypt -v /opt/sesc/certbot/www:/var/www/certbot certbot/certbot renew --quiet && docker-compose -f docker-compose.simple.yml exec nginx nginx -s reload") | crontab -

echo "✅ SSL certificate initialization complete!"
echo ""
echo "📋 Summary:"
echo "  - Certificates obtained for: ${DOMAINS[*]}"
echo "  - Auto-renewal configured (daily at 12:00)"
echo "  - Nginx is running with SSL enabled"
echo ""
echo "🌐 Your sites should now be accessible at:"
echo "  - https://sesc.online (frontend)"
echo "  - https://api.sesc.online (backend API)"
echo ""
echo "🔍 To check certificate status:"
echo "  docker run --rm -v /opt/sesc/certbot/conf:/etc/letsencrypt certbot/certbot certificates"