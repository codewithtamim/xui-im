## Local Development Setup

1. Create a directory named `x-ui` in the project root
2. Rename `.env.example` to `.env`
3. Run the application:
   ```bash
   go run main.go
   ```
   Or build and run:
   ```bash
   go build -o x-ui main.go
   ./x-ui
   ```

## Running with Docker

### Quick Start

1. Ensure `db/` and `cert/` directories exist (they will be created automatically if missing):
   ```bash
   mkdir -p db cert
   ```

2. Build and start the container:
   ```bash
   docker compose up -d --build
   ```

3. Access the panel at **http://localhost:2053** (default credentials: `admin` / `admin`)

### Network Configuration

The `docker-compose.yml` supports two modes depending on your platform:

**Mac / Windows (Docker Desktop)** — Use port mapping (default in this setup):
- `network_mode: host` is commented out
- Ports `2053` (panel) and `2096` (subscription) are exposed
- Access at `http://localhost:2053`

**Linux** — Use host network for full Xray compatibility:
- In `docker-compose.yml`, uncomment `network_mode: host` and comment out the `ports` section
- The container shares the host network; no port mapping needed
- Access at `http://localhost:2053` (or your server IP)

### Useful Commands

```bash
# View logs
docker compose logs -f 3xui

# Stop the container
docker compose down

# Rebuild after code changes
docker compose up -d --build
```
