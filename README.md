# shlink-cli

A command-line client for [Shlink](https://shlink.io), the self-hosted URL shortener.

## Features

- Full CRUD for short URLs (create, list, get, edit, delete)
- Tag management (list, stats, rename, delete)
- Visit statistics (global, per-URL, per-tag, per-domain, orphan)
- Domain management and redirect configuration
- Health check
- Three output formats: **table** (default), **json**, **plain**
- Configuration via environment variables or flags
- Cross-platform: Linux, macOS, Windows (amd64 / arm64)

## Installation

### Pre-built binaries

Download the latest release from the [Releases](../../releases) page and add it to your `PATH`.

### Build from source

```bash
git clone https://github.com/yourorg/shlink-cli.git
cd shlink-cli
make build          # binary in ./bin/shlink
make install        # installs to $GOPATH/bin
```

## Configuration

| Priority | Setting     | Flag          | Environment variable |
|----------|-------------|---------------|----------------------|
| 1 (high) | Server URL  | `--server`    | `SHLINK_SERVER`      |
| 1 (high) | API key     | `--api-key`   | `SHLINK_API_KEY`     |

Set environment variables for convenience:

```bash
export SHLINK_SERVER=https://s.example.com
export SHLINK_API_KEY=your-api-key-here
```

## Usage

```
shlink [global flags] <command> <subcommand> [flags] [args]
```

### Global flags

| Flag            | Default | Description                        |
|-----------------|---------|------------------------------------|
| `--server`      | —       | Shlink server URL                  |
| `--api-key`     | —       | Shlink API key                     |
| `-o, --output`  | `table` | Output format: table, json, plain  |
| `--api-version` | `3`     | Shlink REST API version            |

---

### Short URLs

```bash
# List short URLs (paginated)
shlink urls list
shlink urls list --page 2 --per-page 20
shlink urls list --search "example" --tag marketing
shlink urls list --order-by dateCreated-DESC

# Get details of a specific short URL
shlink urls get my-slug

# Create a short URL
shlink urls create https://example.com/very/long/path
shlink urls create https://example.com --slug custom-slug --title "My Link"
shlink urls create https://example.com --tag promo --tag newsletter --max-visits 100

# Edit an existing short URL
shlink urls edit my-slug --long-url https://new-destination.com
shlink urls edit my-slug --title "Updated title" --tag new-tag
shlink urls edit my-slug --clear-max-visits

# Delete a short URL
shlink urls delete my-slug
shlink urls rm my-slug      # alias
```

### Tags

```bash
shlink tags list
shlink tags stats
shlink tags rename old-name new-name
shlink tags delete my-tag
shlink tags rm tag1 tag2    # delete multiple at once
```

### Visits

```bash
# Global visit summary
shlink visits global

# Visits for a specific short URL
shlink visits list my-slug
shlink visits list my-slug --start-date 2024-01-01T00:00:00Z --exclude-bots

# Orphan visits (hits with no matching short URL)
shlink visits orphan

# Visits grouped by tag or domain
shlink visits tag marketing
shlink visits domain s.example.com
```

### Domains

```bash
shlink domains list

# Set redirect URLs for a domain
shlink domains set-redirects --domain s.example.com \
  --base https://example.com \
  --not-found https://example.com/404 \
  --invalid https://example.com/invalid
```

### Health

```bash
shlink health                          # uses SHLINK_SERVER (no API key required)
shlink health --server https://s.example.com
shlink health -o json
```

## Output formats

```bash
shlink urls list -o table   # default human-readable table
shlink urls list -o json    # raw API JSON (pretty-printed)
shlink urls list -o plain   # tab-separated values (for awk, cut, etc.)
```

## Building for all platforms

```bash
make dist               # creates binaries in ./dist/
make dist-checksums     # also generates sha256 checksums
```

## Development

```bash
make test         # run tests
make lint         # golangci-lint
make vet          # go vet
make fmt          # format source
make check        # vet + lint
make tidy         # tidy go.mod/go.sum
make clean        # remove build artefacts
```

## Docker

```bash
make docker-build
docker run --rm \
  -e SHLINK_SERVER=https://s.example.com \
  -e SHLINK_API_KEY=your-key \
  shlink:latest urls list
```

## License

MIT
