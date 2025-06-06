# SESC Backend Deployment Guide

This guide explains how to deploy the SESC application to production using GitHub Actions, Docker, and Traefik with automatic SSL certificates.

## ⚠️ IMPORTANT: GitHub Secrets Required

**Before deployment, you MUST configure these GitHub secrets in your repository:**

Go to your repository → Settings → Secrets and variables → Actions, and add:

- `POSTGRES_PASSWORD` = `password`
- `MINIO_ROOT_USER` = `minioadmin`  
- `MINIO_ROOT_PASSWORD` = `minioadmin`

The deployment will fail if these secrets are not configured in GitHub.

## Prerequisites

1. **Server Requirements:**
   - Ubuntu/Debian server with Docker and Docker Compose installed
   - Domain names pointing to your server:
     - `sesc.online` (for frontend)
     - `api.sesc.online` (for backend API)
     - `traefik.sesc.online` (optional, for Traefik dashboard)

2. **GitHub Secrets:**
   Configure the following secrets in your GitHub repository settings:
   - `SSH_HOST` - Your server's IP address or hostname
   - `SSH_USER` - SSH username for your server
   - `SSH_KEY` - Private SSH key for authentication
   - `GHCR_TOKEN` - GitHub Container Registry token with package read permissions
   - `POSTGRES_PASSWORD` - Database password (default: `password`)
   - `MINIO_ROOT_USER` - MinIO username (default: `minioadmin`)
   - `MINIO_ROOT_PASSWORD` - MinIO password (default: `minioadmin`)

## Architecture

The deployment consists of the following services:

- **Traefik** - Reverse proxy with automatic SSL certificates via Let's Encrypt
- **PostgreSQL** - Database server
- **MinIO** - Object storage server
- **Backend** - SESC API server
- **Frontend** - React application

## Automatic Deployment

### GitHub Actions Workflow

The deployment is automated via GitHub Actions (`deploy.yml`). It:

1. Connects to your server via SSH
2. Pulls the latest Docker images from GitHub Container Registry
3. Updates the docker-compose configuration
4. Restarts all services with zero downtime

### Triggering Deployment

Deployment happens automatically when:
- Code is pushed to the `main` branch
- Manual trigger via GitHub Actions UI

## Manual Deployment

If you need to deploy manually:

1. **Connect to your server:**
   ```bash
   ssh user@your-server.com
   ```

2. **Create the docker-compose.yml file:**
   ```bash
   # The GitHub Action will create this file automatically
   # Or copy docker-compose.prod.yml from the repository
   ```

3. **No environment variables needed:**
   ```bash
   # Environment variables are now passed from GitHub secrets
   # No manual setup required on the server
   ```

4. **Login to GitHub Container Registry:**
   ```bash
   echo "YOUR_GHCR_TOKEN" | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
   ```

5. **Create Traefik network:**
   ```bash
   docker network create traefik
   ```

6. **Deploy:**
   ```bash
   docker-compose pull
   docker-compose up -d
   ```

## Configuration

### GitHub Secrets Configuration

**IMPORTANT:** This deployment uses `config.yml` for application configuration and GitHub secrets for sensitive credentials.

Configure these secrets in your GitHub repository (Settings → Secrets and variables → Actions):

```
POSTGRES_PASSWORD=password
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

These values are automatically passed to the server during deployment - no manual server configuration required.

### SSL Certificates

Traefik automatically obtains SSL certificates from Let's Encrypt for:
- `sesc.online`
- `api.sesc.online`
- `traefik.sesc.online` (if enabled)

No manual certificate management is required.

### Domains

Update your DNS records to point to your server:
```
sesc.online        A    YOUR_SERVER_IP
api.sesc.online    A    YOUR_SERVER_IP
traefik.sesc.online A   YOUR_SERVER_IP
```

## Monitoring

### Service Status

Check if all services are running:
```bash
docker-compose ps
```

### Logs

View logs for specific services:
```bash
# All services
docker-compose logs

# Specific service
docker-compose logs backend
docker-compose logs traefik
```

### Traefik Dashboard

Access the Traefik dashboard at `https://traefik.sesc.online` to monitor:
- Active routes
- SSL certificate status
- Service health

## Troubleshooting

### Common Issues

1. **SSL Certificate Issues:**
   ```bash
   # Check Traefik logs
   docker-compose logs traefik
   
   # Restart Traefik
   docker-compose restart traefik
   ```

2. **Database Connection Issues:**
   ```bash
   # Check PostgreSQL logs
   docker-compose logs postgres
   
   # Verify database is ready
   docker-compose exec postgres pg_isready -U sesc
   ```

3. **Image Pull Issues:**
   ```bash
   # Re-login to GitHub Container Registry
   echo "YOUR_TOKEN" | docker login ghcr.io -u YOUR_USERNAME --password-stdin
   
   # Pull images manually
   docker pull ghcr.io/kozlov-ma/sesc-backend/backend:latest
   docker pull ghcr.io/kozlov-ma/sesc-backend/frontend:latest
   ```

### Service Restart

Restart individual services without downtime:
```bash
# Restart backend only
docker-compose up -d --no-deps backend

# Restart frontend only
docker-compose up -d --no-deps frontend
```

### Complete Reset

If you need to start fresh:
```bash
# Stop and remove all containers
docker-compose down

# Remove volumes (WARNING: This deletes all data!)
docker-compose down -v

# Clean up images
docker system prune -f

# Redeploy
docker-compose up -d
```

## Security Considerations

1. **Change Default Passwords:** Always set secure passwords in GitHub secrets
2. **SSH Key Security:** Keep your SSH private key secure and rotate regularly
3. **GitHub Token:** Use a token with minimal required permissions
4. **Firewall:** Configure your server firewall to only allow necessary ports (80, 443, 22)
5. **Regular Updates:** Keep Docker images updated by running deployments regularly

## Backup

### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U sesc sesc > backup.sql

# Restore backup
docker-compose exec -T postgres psql -U sesc sesc < backup.sql
```

### MinIO Backup

```bash
# Access MinIO data
docker-compose exec minio ls /data
```

## Support

For deployment issues:
1. Check the GitHub Actions logs
2. Review service logs using `docker-compose logs`
3. Verify DNS configuration
4. Check firewall settings
5. Ensure all environment variables are set correctly