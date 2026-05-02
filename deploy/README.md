# Smedje VPS Deployment

Deploy the public demo at app.smedje.net with self-hosted analytics.

## Prerequisites

- Ubuntu 24.04 (or similar)
- Go 1.22+
- Caddy 2
- Docker + docker-compose
- Node.js 20+ (for building the frontend)

## DNS Records

Point these to your server IP:

| Record | Type | Value |
|--------|------|-------|
| smedje.net | A | <SERVER_IP> |
| www.smedje.net | A | <SERVER_IP> |
| app.smedje.net | A | <SERVER_IP> |
| analytics.smedje.net | A | <SERVER_IP> |

## Quick Start

1. Clone and set up:
   ```bash
   git clone https://github.com/MydsiIversen/smedje.git /opt/smedje
   cd /opt/smedje
   sudo bash deploy/scripts/initial-setup.sh
   ```

2. Generate secrets and fill in /etc/smedje/env:
   ```bash
   smedje password --length 32 --charset alphanum  # UMAMI_DB_PASSWORD
   smedje password --length 64 --charset alphanum  # UMAMI_APP_SECRET
   ```

3. Start Umami:
   ```bash
   cd /opt/smedje/deploy
   cp .env.example .env  # fill in values
   docker compose up -d
   ```

4. Create website in Umami:
   - Visit https://analytics.smedje.net
   - Default login: admin / umami
   - Change password immediately
   - Add website → copy the Website ID → put in /etc/smedje/env as UMAMI_ID

5. Set up Caddy:
   ```bash
   # Generate basicauth hash
   caddy hash-password
   # Update Caddyfile with the hash
   cp deploy/caddy/Caddyfile /etc/caddy/Caddyfile
   caddy reload --config /etc/caddy/Caddyfile
   ```

6. Build and deploy:
   ```bash
   cd /opt/smedje
   bash deploy/scripts/deploy.sh
   ```

## Verification

```bash
curl -s https://smedje.net | head -5         # Landing page
curl -s https://app.smedje.net/healthz       # App health
curl -s https://app.smedje.net/api/version   # Version info
```

## Logs

```bash
journalctl -u smedje-app -f                  # App logs
tail -f /var/log/caddy/smedje-app.log         # Caddy access logs
docker compose -f deploy/docker-compose.yml logs -f umami  # Umami
```

## Backup

Umami data lives in a Docker volume:
```bash
docker compose -f deploy/docker-compose.yml exec umami-db \
  pg_dump -U umami umami > umami-backup-$(date +%Y%m%d).sql
```

## Upgrade

```bash
cd /opt/smedje
bash deploy/scripts/deploy.sh
```
