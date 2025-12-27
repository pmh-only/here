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
- **Colorized Interface**: Beautiful ANSI-colored prompts and messages for better UX

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
export SSH_PASSWORD="your-password"  # Optional: enable authentication
export UNAUTHENTICATED_TIMEOUT="30m"  # Optional: default is 30m

./here-server
```

## Usage

### Client-Side Setup

#### Basic Usage

```bash
# Forward local port 3000
ssh -R1:localhost:3000 yourdomain.com -p 2222

# If password authentication is enabled, you'll see mode selection:
# Welcome to HereServer!
#
# Select mode:
#   1) Anonymous (30m session timeout)
#   2) Login (unlimited session, requires password)
# Enter choice (1 or 2): 1
#
# Anonymous mode selected
#
# You requested 1 service(s):
# #0 (-R1) -> abc-123-def-456-789.yourdomain.com
#
# Note: Anonymous session will timeout after 30m
# Press Ctrl+C or Ctrl+D to exit

# Or select login mode:
# Enter choice (1 or 2): 2
#
# Password: [enter your password]
# Authentication successful!
#
# You requested 1 service(s):
# #0 (-R1) -> abc-123-def-456-789.yourdomain.com
#
# Authenticated session - no timeout
# Press Ctrl+C or Ctrl+D to exit
```

Your service is now accessible at: `http://abc-123-def-456-789.yourdomain.com:8080`

**Important Notes**:
- **Colorized Interface**: The prompts use ANSI colors for better readability:
  - 🎨 Cyan for headers and prompts
  - ✅ Green for success messages
  - ❌ Red for errors and interrupts
  - ⚡ Yellow for warnings and important info
  - 🔗 Green for tunnel URLs
- **Mode Selection**: When password is configured, users choose between:
  - **Anonymous Mode**: Quick access with session timeout (default 30 minutes)
  - **Login Mode**: Requires password, provides unlimited session duration
- **No Password Set**: If `SSH_PASSWORD` is not configured, users connect directly without mode selection or timeout
- **Flexible Access**: Users decide based on their needs - quick test vs long-running service
- **Automatic Cleanup**: Anonymous sessions are automatically cleaned up when timeout expires or connection drops
- **Cancelling**: Press Ctrl+C or Ctrl+D at any time to cancel input or exit the session

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

# SSH password for authentication (default: none, no auth required)
# If set, clients will be prompted for password on connection
export SSH_PASSWORD="your-secure-password"

# Unauthenticated session timeout (default: 30m)
# Authenticated users have unlimited session duration
# Accepts Go duration format: 30m, 1h, 2h30m, etc.
export UNAUTHENTICATED_TIMEOUT="30m"
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
Environment="SSH_PASSWORD=your-secure-password-here"
Environment="UNAUTHENTICATED_TIMEOUT=30m"
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

## Security

### Password Authentication & Mode Selection

HereServer supports optional password authentication with a mode selection system:

```bash
# Enable password authentication on server
export SSH_PASSWORD="your-secure-password"
./here-server
```

**How It Works**:

When `SSH_PASSWORD` is configured, users are presented with two options:
1. **Anonymous Mode**: No password required, but session expires after configured timeout (default 30 minutes)
2. **Login Mode**: Requires password authentication, provides unlimited session duration

This gives users flexibility:
- **Quick testing**: Use anonymous mode for temporary tunnels
- **Production services**: Use login mode for long-running services

**Important Security Notes**:

- The password is transmitted over the SSH connection (which is encrypted)
- Password authentication is handled within the SSH session, not through OpenSSH's authentication layer
- This is a simple shared password for all users
- The password is stored as plain text in environment variables
- Anonymous mode provides resource protection through automatic timeout
- Consider using strong, randomly generated passwords
- For production use, consider implementing:
  - Rate limiting for failed authentication attempts
  - IP-based access controls
  - Integration with proper authentication systems

### Session Timeout

HereServer implements automatic session timeout to protect against abandoned connections:

- **Anonymous sessions**: Automatically terminated after configured timeout (default 30 minutes)
- **Authenticated sessions**: No timeout - unlimited duration
- **No password configured**: No timeout for all users
- **Configurable**: Set `UNAUTHENTICATED_TIMEOUT` environment variable (e.g., `15m`, `1h`, `2h30m`)

**Benefits**:
- Prevents resource exhaustion from forgotten or abandoned tunnels
- Encourages authentication for long-running services
- Automatic cleanup reduces memory usage
- Users can choose between quick access or long-running sessions
- Server admins control timeout policy

**Example Scenarios**:

```bash
# Scenario 1: No password - all users have unlimited sessions
# (No SSH_PASSWORD set)
./here-server

# Scenario 2: With password and default timeout
export SSH_PASSWORD="secure-password"
export UNAUTHENTICATED_TIMEOUT="30m"  # Default, can be omitted
./here-server

# Scenario 3: Shorter timeout to encourage authentication
export SSH_PASSWORD="secure-password"
export UNAUTHENTICATED_TIMEOUT="10m"  # 10 minute timeout
./here-server

# Scenario 4: Generous timeout for development
export SSH_PASSWORD="secure-password"
export UNAUTHENTICATED_TIMEOUT="2h"  # 2 hour timeout
./here-server
```

**Recommendations**:

```bash
# Generate a strong password
openssl rand -base64 32

# Use with environment variable
export SSH_PASSWORD="$(openssl rand -base64 32)"
```

### Best Practices

1. **Use HTTPS**: Put nginx or another reverse proxy in front with SSL/TLS
2. **Firewall Rules**: Restrict SSH port access to known IP ranges
3. **Monitor Logs**: Watch for authentication failures and suspicious activity
4. **Strong Passwords**: Use long, random passwords if using password authentication
5. **Regular Updates**: Keep the server and dependencies up to date

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

### Session Timeout Issues

If your tunnel keeps disconnecting after the timeout period:

```bash
# Option 1: Choose Login mode (option 2) when prompted
ssh -R1:localhost:3000 user@server -p 2222
# Select mode:
#   1) Anonymous (30m session timeout)
#   2) Login (unlimited session, requires password)
# Enter choice (1 or 2): 2
# Password: [enter password]

# Option 2: Adjust timeout on server (requires server config change)
# On server: export UNAUTHENTICATED_TIMEOUT="2h"

# Option 3: Use autossh for automatic reconnection
autossh -M 0 -o "ServerAliveInterval 60" -R1:localhost:3000 user@server -p 2222
```

Check server logs for timeout messages:
```bash
journalctl -u here -f | grep -i timeout
```

### Authentication Failed

If you're getting "Authentication failed" immediately:

- Make sure you press **Enter** after typing your password/choice
- The password prompt waits for you to complete typing before submission
- Check that the password matches the `SSH_PASSWORD` environment variable on the server

### Input Issues

**Mode selection (1 or 2)**: You should see your input echoed back as you type. Press Enter to submit.

**Password**: For security, the password is NOT echoed - you won't see what you type. Just type your password and press Enter.

**Enter key not working**: The server accepts both `\r` (carriage return) and `\n` (newline) as Enter, which should work with all SSH clients and operating systems.

**Cancelling/Exiting**:
- Press **Ctrl+C** to interrupt and cancel at any time (during input prompts or active tunnel)
- Press **Ctrl+D** to terminate the session at any time
- Both will display `^C` or `^D` and exit gracefully
- Works during: mode selection, password entry, AND active tunnel operation
- The server actively monitors for these control characters throughout the session

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
