# Here

A lightweight, self-hosted reverse tunnel service that exposes your local services to the internet through SSH tunneling. Similar to ngrok or localtunnel, but you do not need any binary installation.

## Demo video

[Demonstration Video](https://github.com/user-attachments/assets/32ca5a56-e204-4dd9-8598-8ff80d06e50c)


## Usage

**No installation required** – simply use your built-in OpenSSH client.

```
ssh here.pmh.so -R0:localhost:8080

```

The output will look like this:

```
Here: simple & nodeps tunnel

1 item

│ Tunnel #0 (-R0)
│ https://jpbxrjvfl9.here.pmh.so

↑/k up • ↓/j down • c copy url • p pause/resume ...
Note: Unauthenticated session will timeout after 5m0s
```

And that's it! You can now share your `localhost:8080` service using `https://jpbxrjvfl9.here.pmh.so`

## More features

### Multiple Services at Once

`here` supports multiple `-R` flags, allowing you to expose several services simultaneously.

```
ssh here.pmh.so -R0:localhost:8080 -R1:localhost:9000
```

_But what does `-R0:...` mean? Can I use `-R1234:...` instead?_

Yes, you can! The port number in that part of the argument will be ignored by the server. Duplicated `-Rn:` values also work, so you can use it like this:

```
ssh here.pmh.so -R0:localhost:8080 -R0:localhost:9000
```

### Expose Services on Your Internal Network

You can specify any domain name instead of `localhost`. Your OpenSSH client will resolve the domain, so the server doesn't need to have access to your internal network.

```
ssh here.pmh.so -R0:service.local:80
```

### Custom Subdomain Name

You can override the random subdomain assignment by specifying your preferred subdomain name:

```
ssh here.pmh.so -R myfancy-service:0:localhost:8080
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

| Variable                  | Default       | Description                                                                   |
| ------------------------- | ------------- | ----------------------------------------------------------------------------- |
| `SSH_LISTEN_ADDR`         | `:2222`       | The address and port where the SSH server listens for incoming connections    |
| `HTTP_LISTEN_ADDR`        | `:8080`       | The address and port where the HTTP reverse proxy listens                     |
| `HTTP_HOST_SUFFIX`        | `.local`      | The domain suffix appended to generated subdomains (e.g., `.yourdomain.com`)  |
| `HTTP_HOST_PREFIX`        | `http://`     | The URL scheme prefix for generated links (`http://` or `https://`)           |
| `SSH_PASSWORD`            | _(none)_      | Password for authenticated login mode. If not set, authentication is disabled |
| `UNAUTHENTICATED_TIMEOUT` | `30m`         | Session timeout for anonymous connections (e.g., `5m`, `1h`, `30m`)           |
| `DATA_PATH`               | `/data`       | Directory path where persistent data (like SSH host keys) is stored           |
| `SSH_DOMAIN`              | `here.pmh.so` | For display in usage page                                                     |
| `SSH_CLOSE_MESSAGE`       | `true`        | For disable SSH connection close message page                                 |

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
- **HTTPS backends and virtual-host–based web servers are not supported** because the OpenSSH client does not provide the origin domain information to HereServer. As a result, the correct values for the SNI field and the Host header cannot be determined.

## Copyright

&copy; 2025. Minhyeok Park <pmh_only@pmh.codes>. MIT Licensed.
