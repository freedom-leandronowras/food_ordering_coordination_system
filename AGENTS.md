# AGENTS.md

## Cursor Cloud specific instructions

### Overview

Food ordering coordination platform ("SoftLunch") with four components:
- **web_ui** — Next.js 16 frontend (pnpm, port 3000)
- **coordination_api** — Go 1.24 REST API (port 8080, MongoDB-backed auth + orders)
- **external_services_mocks** — json-server vendor mocks (ports 4011-4013)
- **deployment** — Pulumi IaC for Kubernetes (optional for local dev)

### Running the dev stack

All services can be started with `mise run dev-local` (see `mise.toml` for full task definitions). This starts MongoDB (Docker), the Next.js dev server, Go API with `air` live-reload, and three json-server vendor mocks.

Alternatively, start services individually — see the `[tasks.dev-local]` section in `mise.toml` for the exact commands and environment variables.

A `web_ui/.env.local` file is required before starting. Minimum contents:
```
NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:8080
JWT_SIGNING_KEY=local-dev-signing-key
```

### Key caveats

- **Docker socket permissions**: In Cloud VM, after Docker is installed you may need `sudo chmod 666 /var/run/docker.sock` before Go tests (testcontainers) or `mise run dev-local` can access Docker.
- **Go version**: The `coordination_api` requires Go 1.24+. The `deployment` module requires Go 1.25 (auto-downloaded by the Go toolchain).
- **`air` for live-reload**: Install via `go install github.com/air-verse/air@latest`. Required by `mise run dev-local`.
- **pnpm build scripts warning**: `pnpm install` warns about ignored build scripts for `@clerk/shared`. This is safe to ignore — Clerk is not used at runtime.

### Lint / Test / Build commands

See `README.md` "Testing" section for the validated command set:
```bash
# Lint
cd web_ui && pnpm run lint
cd coordination_api && go vet ./...

# Tests (require Docker access for testcontainers)
cd coordination_api && go test ./internal/http ./test/behavior
cd web_ui && pnpm exec next typegen && pnpm exec tsc --noEmit
cd deployment && go test ./...

# Build
cd web_ui && pnpm run build
```

### Auth for local dev

`AUTH_ALLOW_SELF_ASSIGN_ROLES=true` (default in dev-local) allows registering with any role including `HIVE_MANAGER` and `INNOVATION_LEAD`. Registration endpoint: `POST /api/auth/register` with `{"email","password","full_name","role"}`.
