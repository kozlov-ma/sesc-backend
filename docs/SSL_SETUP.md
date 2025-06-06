# SSL Setup Guide for SESC Project

This guide explains how to set up SSL certificates using Let's Encrypt for your SESC deployment.

## Overview

The SESC project uses nginx as a reverse proxy with Let's Encrypt SSL certificates to provide secure HTTPS access to:
- **Frontend**: `https://sesc.online`
- **Backend API**: `https://api.sesc.online`

## Prerequisites

1. **Domain Configuration**: Ensure your domains point to your server:
   - `sesc.online` → Your server IP
   - `api.sesc.online` → Your server IP
   - `www.sesc.online` → Your server IP (optional, will redirect to sesc.online)

2. **Server Access**: SSH access to your production server

3. **Ports Open**: Ensure ports 80 and 443 are open on your server

## Initial SSL Setup

### Step 1: Deploy to Server

First, make sure your application is deployed to the server:

```bash
# This will deploy the application with dummy SSL certificates
# The workflow will automatically create dummy certificates if real ones don't exist
```

### Step 2: Configure Email

Edit the email address in the SSL initialization script:

```bash
ssh user@your-server
cd /opt/sesc
nano scripts/init-ssl.sh
# Change EMAIL="admin@sesc.online" to your actual email
```

### Step 3: Run SSL Initialization

```bash
# Make the script executable
chmod +x scripts/init-ssl.sh

# Run the SSL initialization (this will take a few minutes)
./scripts/init-ssl.sh
```

The script will:
1. Download recommended TLS parameters
2. Create dummy certificates temporarily
3. Start nginx with dummy certificates
4. Request real Let's Encrypt certificates
5. Reload nginx with real certificates
6. Set up automatic certificate renewal

### Step 4: Verify SSL Setup

Check that your certificates are working:

```bash
# Check certificate status
docker run --rm -v /opt/sesc/certbot/conf:/etc/letsencrypt certbot/certbot certificates

# Test your sites
curl -I https://sesc.online
curl -I https://api.sesc.online
```

## Certificate Management

### Checking Certificate Expiry

```bash
# View all certificates and their expiry dates
docker run --rm -v /opt/sesc/certbot/conf:/etc/letsencrypt certbot/certbot certificates

# Check a specific certificate
openssl x509 -in /opt/sesc/certbot/conf/live/sesc.online/fullchain.pem -noout -dates
```

### Manual Certificate Renewal

Certificates are automatically renewed, but you can manually renew them:

```bash
# Renew all certificates
docker run --rm -v /opt/sesc/certbot/conf:/etc/letsencrypt -v /opt/sesc/certbot/www:/var/www/certbot certbot/certbot renew

# Reload nginx after renewal
docker-compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

### Auto-Renewal

The initialization script sets up a cron job that:
- Runs daily at 12:00 PM
- Checks for certificates expiring in 30 days
- Automatically renews them if needed
- Reloads nginx after successful renewal

View the cron job:
```bash
crontab -l
```

## Testing SSL Configuration

### SSL Labs Test

Test your SSL configuration quality:
- Visit: https://www.ssllabs.com/ssltest/
- Enter: `sesc.online` and `api.sesc.online`
- Should achieve A+ rating

### Local Testing

```bash
# Test HTTPS redirect
curl -I http://sesc.online
# Should return 301 redirect to https://

# Test API endpoint
curl -I https://api.sesc.online/health
# Should return your API response

# Test security headers
curl -I https://sesc.online
# Should include security headers like HSTS, X-Frame-Options, etc.
```

## Troubleshooting

### Common Issues

1. **Certificate Request Failed**
   ```bash
   # Check if domains point to your server
   nslookup sesc.online
   nslookup api.sesc.online
   
   # Check if ports 80/443 are accessible
   nc -zv your-server-ip 80
   nc -zv your-server-ip 443
   ```

2. **Nginx Won't Start**
   ```bash
   # Check nginx logs
   docker-compose -f docker-compose.prod.yml logs nginx
   
   # Test nginx configuration
   docker-compose -f docker-compose.prod.yml exec nginx nginx -t
   ```

3. **Certificate Renewal Issues**
   ```bash
   # Check renewal logs
   docker run --rm -v /opt/sesc/certbot/conf:/etc/letsencrypt -v /opt/sesc/certbot/logs:/var/log/letsencrypt certbot/certbot renew --dry-run
   ```

### Emergency Recovery

If SSL breaks and you need to quickly restore service:

```bash
# Stop nginx
docker-compose -f docker-compose.prod.yml stop nginx

# Temporarily expose services directly (NOT for production)
docker-compose -f docker-compose.prod.yml run -d -p 80:80 frontend
docker-compose -f docker-compose.prod.yml run -d -p 8080:8080 backend

# Fix SSL issues, then restart nginx
docker-compose -f docker-compose.prod.yml start nginx
```

## Security Notes

1. **Certificate Security**: 
   - Private keys are stored in `/opt/sesc/certbot/conf/live/`
   - Ensure proper file permissions (600 for private keys)
   - Regular backups of `/opt/sesc/certbot/conf/`

2. **Nginx Security**:
   - Security headers are configured in nginx.conf
   - Rate limiting is enabled for API endpoints
   - CORS is properly configured for API

3. **Monitoring**:
   - Set up monitoring for certificate expiry
   - Monitor nginx access/error logs
   - Set up alerts for SSL-related failures

## Advanced Configuration

### Custom Certificate Paths

If you need to use different certificate paths, modify `nginx/nginx.conf`:

```nginx
ssl_certificate /path/to/your/fullchain.pem;
ssl_certificate_key /path/to/your/privkey.pem;
```

### Staging Certificates (for testing)

To test with staging certificates first:

```bash
# Edit scripts/init-ssl.sh
# Change STAGING=0 to STAGING=1
# Run the script
./scripts/init-ssl.sh
```

### Additional Domains

To add more domains, update:
1. `scripts/init-ssl.sh` - add domains to DOMAINS array
2. `nginx/nginx.conf` - add server blocks for new domains

## Support

For issues related to:
- **Let's Encrypt**: https://community.letsencrypt.org/
- **Nginx**: https://nginx.org/en/docs/
- **This setup**: Create an issue in the project repository