# Here

A lightweight, self-hosted reverse tunnel service that exposes your local services to the internet through SSH tunneling. Similar to ngrok or localtunnel, but you control the infrastructure.

## Features

- **SSH-Based Tunneling**: Secure connections using standard SSH protocol with remote port forwarding
- **HTTP Reverse Proxy**: Routes incoming HTTP requests through SSH tunnels to your local services
- **UUID-Based Routing**: Each tunnel gets a unique identifier for clean URL management
- **Concurrent Sessions**: Support multiple tunnels and clients simultaneously
- **Thread-Safe**: Protected against race conditions with proper synchronization
- **Graceful Shutdown**: Clean resource cleanup on termination
- **Proxy Headers**: Automatic X-Forwarded-For, X-Real-IP, and other standard headers
- **Memory Safe**: Automatic cleanup of stale connections and mappings

## How It Works

1. Client connects via SSH with remote port forwarding: `-R1:localhost:PORT`
2. Server generates a unique UUID for the forwarded port
3. HTTP requests to `{UUID}.yourdomain.com` are proxied through the SSH tunnel
4. Your local service receives the request and responds
5. Response flows back through the tunnel to the original client

```
Internet Client → HTTP Server (Port 8080) → SSH Tunnel → Your Local Service
                       ↓
              UUID-based routing
              (abc-123-def.yourdomain.com)
```

## Installation

### Prerequisites

- Go 1.25.5 or later
- Domain name with wildcard DNS (\*.yourdomain.com)

### Build from Source

```bash
git clone https://github.com/pmh-only/here.git
cd here
go build -o here-server .
```

### Run the Server

```bash
# Create data directory for SSH host key
mkdir -p data

# Run here
export DATA_PATH="./data"
export SSH_LISTEN_ADDR=":2222"
export HTTP_LISTEN_ADDR=":8080"
export HTTP_HOST_SURFFIX=".yourdomain.com"

./here-server
```

## Usage

### Client-Side Setup

#### Basic Usage

```bash
# Forward local port 3000
ssh -R1:localhost:3000 yourdomain.com -p 2222

# Server will respond with:
# Welcome to HereServer!
# You requested 1 service(s):
# 1 -> abc-123-def-456-789
```

Your service is now accessible at: `http://abc-123-def-456-789.yourdomain.com:8080`

#### Multiple Ports

```bash
# Forward multiple ports in one connection
ssh -R1:localhost:3000 \
    -R2:localhost:8000 \
    user@your-server.com -p 2222
```

### Server-Side Setup

#### Configuration

Set environment variables to customize the server:

```bash
# SSH server listen address (default: :2222)
export SSH_LISTEN_ADDR=":2222"

# HTTP server listen address (default: :8080)
export HTTP_LISTEN_ADDR=":8080"

# Data directory for SSH host key (default: /data)
export DATA_PATH="/var/lib/here"

# Host suffix for URL routing (default: empty)
export HOST_SUFFIX=".tunnel.example.com"
```

#### DNS Setup

Configure wildcard DNS to point to your server:

```
*.tunnel.example.com.  IN  A  203.0.113.10
```

#### Systemd Service

Create `/etc/systemd/system/here.service`:

```ini
[Unit]
Description=Here Reverse Tunnel Service
After=network.target

[Service]
Type=simple
User=here
Group=here
WorkingDirectory=/opt/here
Environment="SSH_LISTEN_ADDR=:2222"
Environment="HTTP_LISTEN_ADDR=:8080"
Environment="DATA_PATH=/var/lib/here"
Environment="HOST_SUFFIX=.tunnel.example.com"
ExecStart=/opt/here/here-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable here
sudo systemctl start here
```

#### Reverse Proxy with Nginx

To add HTTPS support, use nginx as a reverse proxy:

```nginx
server {
    listen 443 ssl http2;
    server_name *.tunnel.example.com;

    ssl_certificate /etc/letsencrypt/live/tunnel.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tunnel.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## Proxy Headers

The server automatically adds standard proxy headers to forwarded requests:

- **X-Forwarded-For**: Chain of client IPs through proxies
- **X-Real-IP**: Immediate client IP address
- **X-Forwarded-Proto**: Original protocol (http/https)
- **X-Forwarded-Host**: Original host header from client

Your backend application can use these headers to get the real client information:

```javascript
// Express.js example
app.set("trust proxy", true);
app.get("/", (req, res) => {
  const clientIP = req.ip; // Uses X-Forwarded-For
  res.send(`Your IP: ${clientIP}`);
});
```

## Architecture

### Components

- **SSH Server** (port 2222): Accepts SSH connections with remote port forwarding
- **HTTP Server** (port 8080): Reverse proxy that routes requests through tunnels
- **Mapping Registry**: Thread-safe map of UUIDs to SSH connections
- **Tunnel Manager**: Handles SSH channel creation and cleanup

### Request Flow

```
1. HTTP Request arrives at http://abc-123-def.yourdomain.com:8080
2. Server extracts UUID (abc-123-def) from Host header
3. Looks up SSH connection associated with UUID
4. Opens "forwarded-tcpip" channel through SSH tunnel
5. Proxies HTTP request through channel to client's local service
6. Response flows back through same channel
7. Server sends response to original HTTP client
```

## Troubleshooting

### Connection Refused

```bash
# Check if server is running
systemctl status here

# Check if port is listening
ss -tln | grep -E '2222|8080'
```

### Tunnel Not Working

```bash
# Verify SSH connection
ssh -v -R1:localhost:3000 user@server -p 2222

# Check server logs for UUID
journalctl -u here -n 50
```

### DNS Not Resolving

```bash
# Test wildcard DNS
nslookup abc-123-def.tunnel.example.com

# Should return your server IP
```

### Service Unavailable

This usually means the SSH connection was closed.

## Development

### Project Structure

```
.
├── main.go          # Entry point and HereServer struct
├── model.go         # Data structures for SSH protocol
├── config.go        # Configuration helpers
├── ssh.go           # SSH server implementation
├── ssh_config.go    # SSH host key management
├── http.go          # HTTP reverse proxy
├── http_config.go   # HTTP configuration
└── tunnel.go        # SSH channel to net.Conn adapter
```

### Building

```bash
# Build
go build -o here-server .

# Run tests
go test ./...

# Check for issues
go vet ./...

# Format code
go fmt ./...
```

### Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

&copy; 2025. Minhyeok Park <pmh_only@pmh.codes>. MIT Licensed.

## Acknowledgments

Built with:

- [gliderlabs/ssh](https://github.com/gliderlabs/ssh) - SSH server library
- [golang.org/x/crypto/ssh](https://golang.org/x/crypto/ssh) - SSH protocol implementation
- [google/uuid](https://github.com/google/uuid) - UUID generation
