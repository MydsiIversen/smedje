#!/usr/bin/env bash
set -euo pipefail

cd /opt/smedje
git pull origin main
make build-public
cp ./smedje /usr/local/bin/smedje
systemctl restart smedje-app
rsync -av --delete landing/ /var/www/smedje-landing/
caddy reload --config /etc/caddy/Caddyfile
echo "Deployed."
