# Shorty

A fast and efficient URL shortener with file upload support and flexible authentication.

## Features

- **URL Shortening**: Create short, memorable links from long URLs
- **Custom Names**: Optional custom names for shortened URLs
- **File Uploads**: Upload and serve files with auto-generated short links (S3-compatible storage)
- **API**: RESTful API (`/v1/*`) with API key authentication
- **Web UI**: Modern, responsive web interface for managing links
- **Flexible Authentication**:
  - OAuth (GitLab/Gitea support)
  - Username/Password login
  - No authentication mode (public access)
- **Auto Cleanup**: Automatic removal of expired shortened URLs and orphaned files
- **Redis Backend**: Fast caching and session management
- **Docker Ready**: Includes Dockerfile for easy deployment

## Quick Start

### Requirements
- Go 1.21+
- Redis 6.0+
- (Optional) S3-compatible storage for file uploads

### Installation

1. Clone the repository
2. Copy and configure `config.yaml`:
```yaml
app:
  listen: :1106
  key: "your-secret-key"
  auth:
    user: "admin"
    password: "your-password"
  base_url: "https://your-domain.com"

redis:
  host: "127.0.0.1"
  port: "6379"
```

3. Build and run:
```bash
go build
./shorty
```

Access the web UI at `http://localhost:1106`

## Configuration

### Authentication Methods

**Option 1: Username/Password (Default)**
```yaml
app:
  auth:
    user: "admin"
    password: "secure-password"
```

**Option 2: OAuth (GitLab/Gitea)**
```yaml
oauth:
  enable: true
  client_id: "your-client-id"
  client_secret: "your-client-secret"
  base_url: "https://gitlab.example.com"
```

**Option 3: API Only (No Web UI Auth)**
- Leave `oauth.enable: false` and `app.auth` empty

### File Uploads (S3)
```yaml
s3:
  enable: true
  endpoint: "minio.example.com"
  bucket: "shorty"
  key:
    access: "your-access-key"
    secret: "your-secret-key"
  cleanup_interval: "1h"
```

## API Usage

### Authentication
Use the `Authorization` header with your API key:
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://your-domain.com/v1/shorty
```

### Create Short URL
```bash
POST /v1/shorty
{
  "url": "https://example.com/very/long/url",
  "name": "optional-custom-name"
}
```

### List All URLs
```bash
GET /v1/list
```

### Rename URL
```bash
PATCH /v1/oldname/newname
```

### Delete URL
```bash
DELETE /v1/shortname
```

## API Key Types

- **App.Key**: Cryptographic key for generating API tokens (required)
- **App.Token**: Optional static token for direct API access

## Environment Variables

All configuration can be set via environment:
- `LISTEN`: Server address (default: `:1106`)
- `KEY`: Secret key for API tokens (required)
- `TOKEN`: Static API token (optional)
- `AUTH_USER`: Username for web login
- `AUTH_PASSWORD`: Password for web login
- `OAUTH_ENABLE`: Enable OAuth (true/false)
- `OAUTH_CLIENT_ID`: OAuth client ID
- `OAUTH_CLIENT_SECRET`: OAuth client secret
- `OAUTH_BASE_URL`: OAuth provider URL
- `REDIS_HOST`: Redis host (default: `127.0.0.1`)
- `REDIS_PORT`: Redis port (default: `6379`)
- `S3_ENABLE`: Enable S3 uploads (true/false)
- `S3_ENDPOINT`: S3 endpoint URL
- `S3_BUCKET`: S3 bucket name

## Development

### Build Frontend
```bash
cd ui
bun install
bun run build
cd ..
go build
```

### Run Linter
```bash
golangci-lint run ./...
```

## Architecture

- **Backend**: Go with Fiber framework
- **Frontend**: SvelteKit with Tailwind CSS
- **Storage**: Redis (session/cache) + S3-compatible storage (files)
- **Authentication**: Pluggable (OAuth, Username/Password, or None)

## Performance

- Redis caching for fast URL lookups
- 10-minute auth token cache
- Automatic cleanup of expired URLs and orphaned files
- Optimized for high-throughput scenarios

## Docker

```bash
docker build -t shorty .
docker run -p 1106:1106 \
  -e KEY="your-secret-key" \
  -e AUTH_PASSWORD="your-password" \
  -e REDIS_HOST="redis-host" \
  shorty
```
