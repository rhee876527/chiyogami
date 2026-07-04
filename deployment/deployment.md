# Chiyogami Deployment Guide

## Docker (Recommended)

Docker is the recommended way to deploy Chiyogami. Pre-built images are available on
GHCR. See the [README](../README.md) for quick-start instructions, environment
variables, and the `docker-compose.yml` example.

```bash
docker run -d \
  -v "$(pwd)/pastes:/pastes" \
  -p 127.0.0.1:8000:8000 \
  --restart unless-stopped \
  ghcr.io/rhee876527/chiyogami:latest
```

---

## Podman

Podman can run the same `docker-compose.yml` file using `podman-compose` or
`podman play kube`. The `./pastes` directory must exist before starting:

```bash
mkdir -p ./pastes
podman-compose up -d
```

Or with plain Podman:

```bash
mkdir -p ./pastes
podman run -d \
  -v "$(pwd)/pastes:/pastes:Z" \
  -p 127.0.0.1:8000:8000 \
  --restart unless-stopped \
  ghcr.io/rhee876527/chiyogami:latest
```

The `:Z` relabels the volume for SELinux (required on most Podman hosts).

---

## Systemd Service (Binary Install)

If you prefer to run the binary directly, use the following systemd
service file.

Create `/etc/systemd/system/chiyogami.service`:

```ini
[Unit]
Description=Chiyogami pastebin
After=network.target

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=/home/ubuntu/.chiyogami
ExecStart=/home/ubuntu/.chiyogami/chiyogami-1.5.5-linux-amd64

# Inline environment variables (edit as needed)
Environment=SECRET_KEY=change-me-to-a-random-string
Environment=PORT=8000

Restart=always
RestartSec=5

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
```

Then enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now chiyogami
```

**Adding more environment variables** - simply add additional `Environment=`
lines in the `[Service]` section. The full list of available variables is
documented in the [README](../README.md#environment-variables).

---

## Reverse Proxy Configuration

Chiyogami listens on port **8000** internally. When placed behind a reverse
proxy, you need to forward the real client IP so the application rate-limiter sees the correct origin. Both client supplied headers `X-Real-IP` and `X-Forwarded-For` should be treated as untrusted.

### Nginx (1.18+)

```nginx
server {
    listen 80;
    server_name chiyogami.example.com;

    # Compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 256;
    gzip_types
        text/html
        text/plain
        text/css
        application/json
        application/javascript
        image/svg+xml;

    # Trust the reverse proxy's IP (or subnet) so Nginx can read
    # the real client IP from X-Forwarded-For. Replace with your setup.
    set_real_ip_from 10.0.0.0/8;
    set_real_ip_from 172.16.0.0/12;
    set_real_ip_from 192.168.0.0/16;
    real_ip_header X-Forwarded-For;
    real_ip_recursive on;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

- `set_real_ip_from` - Lists trusted proxy IPs/subnets. With `real_ip_recursive` on, Nginx walks
   the `X-Forwarded-For` chain and replaces the client address with the first non-trusted IP.
- `real_ip_header X-Forwarded-For` - Tells Nginx to read the client IP from
  this header.
- `proxy_set_header X-Real-IP $remote_addr` - Forwards the resolved real IP
  (after `real_ip` module processing) to the backend.
- `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for` - Appends
  Nginx's own IP to the existing chain; the backend sees the full proxy path.

**Protect `/health` from external access:**

```nginx
location = /health {
    return 301 /nonexistent-path;
}
```

### Caddy (2.x)

```caddyfile
chiyogami.example.com {
    encode zstd gzip

    redir /health /nonexistent-path 301

    reverse_proxy 127.0.0.1:8000 {
        header_up X-Real-IP {remote_host}
    }
}
```

Caddy automatically adds `X-Forwarded-For`, `X-Forwarded-Proto`, and
`X-Forwarded-Host`. The `X-Real-IP` header must be set explicitly via
`header_up X-Real-IP {remote_host}`.

### Traefik (v2+)

Traefik automatically sets both `X-Real-IP` and `X-Forwarded-For`. No extra
headers are needed for client IP preservation.

**Docker Compose labels (with compression + health protection):**

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.chiyogami.rule=Host(`chiyogami.example.com`)"
  - "traefik.http.routers.chiyogami.entrypoints=websecure"
  - "traefik.http.routers.chiyogami.tls.certresolver=letsencrypt"
  - "traefik.http.services.chiyogami.loadbalancer.server.port=8000"
  # Compression
  - "traefik.http.middlewares.chiyogami-compress.compress=true"
  # Health endpoint protection
  - "traefik.http.middlewares.chiyogami-health.replacepathregex.regex=^/health$"
  - "traefik.http.middlewares.chiyogami-health.replacepathregex.replacement=/nonexistent-path"
  # Chain both middlewares
  - "traefik.http.routers.chiyogami.middlewares=chiyogami-compress,chiyogami-health"
```

### Apache (2.4+)

```apache
<VirtualHost *:80>
    ServerName chiyogami.example.com

    # Compression
    DeflateCompressionLevel 6
    AddOutputFilterByType DEFLATE text/html text/plain text/css application/json application/javascript image/svg+xml
    Header append Vary Accept-Encoding

    ProxyRequests Off
    ProxyPreserveHost On
    ProxyAddHeaders On

    ProxyPass / http://127.0.0.1:8000/
    ProxyPassReverse / http://127.0.0.1:8000/

    RequestHeader set X-Real-IP "%{REMOTE_ADDR}s"
</VirtualHost>
```

`ProxyAddHeaders On` (default) automatically appends `X-Forwarded-For`,
`X-Forwarded-Host`, and `X-Forwarded-Server`. Set `X-Real-IP` explicitly with
`RequestHeader`.

Requires modules: `proxy`, `proxy_http`, `headers`, `deflate`.

**Protect `/health` from external access:**

```apache
ProxyPass /health !
Redirect 301 /health /nonexistent-path
```

```bash
sudo a2enmod proxy proxy_http headers deflate
sudo systemctl reload apache2
```

### HAProxy (2.0+)

```haproxy
frontend chiyogami_front
    bind *:80
    mode http
    default_backend chiyogami_back

backend chiyogami_back
    mode http
    option forwardfor
    filter compression
    compression algo gzip
    compression type text/html text/plain text/css application/json application/javascript image/svg+xml
    http-request set-header X-Real-IP %[src]
    server chiyogami 127.0.0.1:8000 check
```

`option forwardfor` adds the `X-Forwarded-For` header. `X-Real-IP` is set
explicitly with `http-request set-header X-Real-IP %[src]`.

**Protect `/health` from external access:**

```haproxy
frontend chiyogami_front
    bind *:80
    mode http

    acl is_health path -m str /health
    http-request redirect location /nonexistent-path code 301 if is_health

    default_backend chiyogami_back
```