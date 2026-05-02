#!/usr/bin/env bash
set -euo pipefail

# Create system user
if ! id -u smedje &>/dev/null; then
    useradd --system --shell /usr/sbin/nologin --home-dir /opt/smedje smedje
    echo "Created user: smedje"
fi

# Create directories
mkdir -p /opt/smedje /etc/smedje /var/www/smedje-landing /var/log/caddy

# Copy systemd service
cp deploy/systemd/smedje-app.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable smedje-app

# Create env file if missing
if [ ! -f /etc/smedje/env ]; then
    cp deploy/.env.example /etc/smedje/env
    chmod 600 /etc/smedje/env
    echo "Created /etc/smedje/env — fill in values before starting"
fi

echo "Setup complete. Next steps:"
echo "  1. Edit /etc/smedje/env with your secrets"
echo "  2. Run: make build-public"
echo "  3. Run: deploy/scripts/deploy.sh"
