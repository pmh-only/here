# Here

A lightweight, self-hosted reverse tunnel service that exposes your local services to the internet through SSH tunneling. Similar to ngrok or localtunnel, but you do not need any binary installation.

> **⚠️ MVP Warning**
> This project is currently a Minimum Viable Product. Some features may be blocked or incomplete, and you may encounter bugs. Please use with caution and report any issues you find.

## Demo video

https://github.com/user-attachments/assets/89e2cc2e-96f8-4d08-a554-ca08143c4aaf

## Usage

**No installation required** – simply use your built-in OpenSSH client.

```
ssh here.pmh.so -P0:localhost:8080
```

The output will look like this:

```
Allocated port 11346 for remote forward to localhost:8080

╔═══════════════════════════════╗
║   Welcome to HereServer!      ║
╚═══════════════════════════════╝

Select mode:
  1) Anonymous (5m0s session timeout)
  2) Login (unlimited session, requires password)

Enter choice (1 or 2): 1
✓ Anonymous mode selected

You requested 1 service(s):
  #0 (-R0) -> https://j4qyberht3.here.pmh.so

⏱  Note: Anonymous session will timeout after 5m0s
⌨  Press Ctrl+C or Ctrl+D to exit
```

And that's it! You can now share your `localhost:8080` service using `https://j4qyberht3.here.pmh.so`

## More features

### Multiple Services at Once

`here` supports multiple `-P` flags, allowing you to expose several services simultaneously.

```
ssh here.pmh.so -P0:localhost:8080 -P1:localhost:9000
```

_But what does `-R0:...` mean? Can I use `-R1234:...` instead?_

Yes, you can! The port number in that part of the argument will be ignored by the server. Duplicated `-Rn:` values also work, so you can use it like this:

```
ssh here.pmh.so -P0:localhost:8080 -P0:localhost:9000
```

### Expose Services on Your Internal Network

You can specify any domain name instead of `localhost`. Your OpenSSH client will resolve the domain, so the server doesn't need to have access to your internal network.

```
ssh here.pmh.so -P0:service.local:80
```

### Custom Subdomain Name

You can override the random subdomain assignment by specifying your preferred subdomain name:

```
ssh here.pmh.so -P myfancy_service:0:localhost:8080
```

## Self-Hosting

You can self-host this service using Docker. The image is available at `ghcr.io/pmh-only/here:latest`.

### Quick Start

```bash
docker run -d \
  -p 2222:2222 \
  -p 8080:8080 \
  -v /path/to/data:/data \
  -e HTTP_HOST_SUFFIX=".yourdomain.com" \
  -e HTTP_HOST_PREFIX="https://" \
  ghcr.io/pmh-only/here:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SSH_LISTEN_ADDR` | `:2222` | The address and port where the SSH server listens for incoming connections |
| `HTTP_LISTEN_ADDR` | `:8080` | The address and port where the HTTP reverse proxy listens |
| `HTTP_HOST_SUFFIX` | `.local` | The domain suffix appended to generated subdomains (e.g., `.yourdomain.com`) |
| `HTTP_HOST_PREFIX` | `http://` | The URL scheme prefix for generated links (`http://` or `https://`) |
| `SSH_PASSWORD` | _(none)_ | Password for authenticated login mode. If not set, authentication is disabled |
| `UNAUTHENTICATED_TIMEOUT` | `30m` | Session timeout for anonymous connections (e.g., `5m`, `1h`, `30m`) |
| `DATA_PATH` | `/data` | Directory path where persistent data (like SSH host keys) is stored |

### Docker Compose Example

```yaml
services:
  here:
    image: ghcr.io/pmh-only/here:latest
    ports:
      - "2222:2222"
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      HTTP_HOST_SUFFIX: ".yourdomain.com"
      HTTP_HOST_PREFIX: "https://"
      SSH_PASSWORD: "your-secure-password"
      UNAUTHENTICATED_TIMEOUT: "5m"
    restart: unless-stopped
```

### Notes

- Make sure to configure your DNS to point `*.yourdomain.com` to your server's IP address
- If using HTTPS, you'll need to set up a reverse proxy (like nginx or Caddy) with SSL certificates in front of the HTTP port
- The SSH host key is automatically generated on first run and stored in the data directory

## Limitations

- **Only HTTP connections are supported (not TCP or UDP level)**: This limitation exists because the service uses subdomain-based routing. Supporting raw TCP or UDP would require dedicated IP addresses for each tunnel, which is beyond the scope of this project.

## Copyright

&copy; 2025. Minhyeok Park <pmh_only@pmh.codes>. MIT Licensed.
