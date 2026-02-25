# AGENTS.md

## Cursor Cloud specific instructions

### Architecture

- **Next.js frontend** (root `package.json`): runs on port 3000 via `pnpm dev`
- **Go API backend** (`api/`): runs on a configurable port via env vars `PORT`, `MONGODB_URI`, `MONGODB_DATABASE`
- **MongoDB**: required by the Go API at runtime

### Running Services

1. **Docker** must be running before Go tests or the Go API (MongoDB dependency).
   - In the cloud VM, start Docker with: `sudo nohup dockerd > /tmp/dockerd.log 2>&1 &`
   - Docker uses `fuse-overlayfs` storage driver and `iptables-legacy` (required for nested containers in Firecracker VMs).
2. **MongoDB**: `sudo docker run -d --name mongodb -p 27017:27017 mongo:7`
3. **Go API**: `cd api && PORT=8080 MONGODB_URI=mongodb://localhost:27017 MONGODB_DATABASE=food_ordering go run ./_local`
4. **Next.js frontend**: `pnpm dev` (port 3000 by default)

### Lint / Test / Build

- **Frontend lint**: `pnpm lint` (ESLint)
- **Frontend build**: `pnpm build` — Note: currently fails due to a deprecated `instrumentationHook` key in `next.config.ts`. The dev server works fine (shows warning only).
- **Go vet**: `cd api && go vet ./...`
- **Go tests** (require Docker running): `cd api && sudo go test ./... -v -timeout 120s`
  - Tests use `testcontainers-go` to spin up ephemeral MongoDB containers automatically.
  - Must run with `sudo` because Docker requires root in this VM environment.
  - Includes both HTTP integration tests (`internal/http/`) and BDD/Gherkin behavior tests (`test/behavior/`).

### Auth (JWT)

The API uses unverified JWT parsing (no signature check). For testing, generate tokens with any signing key. The JWT must include `sub` (UUID, the member ID), `role` (one of `MEMBER`, `HIVE_MANAGER`, `INNOVATION_LEAD`), and `exp` claims.

### Go version

The project requires Go 1.24. The VM ships with Go 1.22; the update script installs Go 1.24 to `/usr/local/go`.
