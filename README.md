# Food Ordering Coordination System

Food ordering platform with:
- `web_ui`: Next.js frontend.
- `coordination_api`: Go backend with domain logic, MongoDB persistence, JWT auth, and RBAC.
- `external_services_mocks`: vendor mock data served by `json-server`.
- `deployment`: Pulumi stack (organized in constructs + stack composition) for local Kubernetes deployment.
- `mise.toml` at repo root with atomic tasks for build, deploy, smoke, and local live development.

## Current Auth Model

Clerk has been removed from runtime. Authentication is MongoDB-backed and handled by the Go API.

- Passwords are hashed with `bcrypt`.
- JWTs are signed/verified with `HS256` using `JWT_SIGNING_KEY`.
- JWT claims include `sub` (member id), `role`, `user_id`, `email`, and `exp`.
- Frontend stores the session token in cookie `focs_session`.
- RBAC roles:
  - `MEMBER`
  - `HIVE_MANAGER`
  - `INNOVATION_LEAD`

### Auth Endpoints

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`
- `GET /api/auth/members?domain=...` (manager roles only)

`AUTH_ALLOW_SELF_ASSIGN_ROLES` controls whether register requests can set manager roles.
- Recommended production value: `false`.
- Local dev default in `mise run dev-local`: `true` (to allow quick manager account creation).

## Core API Endpoints

- `GET /api/menus`
- `GET /api/vendors`
- `POST /api/orders`
- `GET /api/members/{memberId}/orders`
- `GET /api/members/{memberId}/credits`
- `POST /api/members/{memberId}/credits` (manager roles only)

## Local Development (Live Reload)

Run everything locally with live reload:

```bash
mise run dev-local
```

What starts:
- Web UI: `http://127.0.0.1:3000`
- API: `http://127.0.0.1:8080`
- Vendor mocks: ports `4011`, `4012`, `4013`
- MongoDB container: `food-ordering-mongo-dev` on `27017`

Stop only the dev Mongo container:

```bash
mise run dev-local-stop
```

## Local Kubernetes Deployment

Run full image build + Pulumi deploy + rollout + smoke checks:

```bash
mise run deploy-local
```

Atomic tasks are available (examples):
- `mise run build-api-image`
- `mise run build-vendor-image`
- `mise run build-web-image`
- `mise run pulumi-config`
- `mise run pulumi-up`
- `mise run rollout`
- `mise run smoke`
- `mise run status`

## Environment

Use `.env.example` as baseline.

Important variables:
- `JWT_SIGNING_KEY` (required)
- `AUTH_TOKEN_TTL_SECONDS` (default `3600`)
- `AUTH_ALLOW_SELF_ASSIGN_ROLES` (default `false`)
- `NEXT_PUBLIC_API_BASE_URL` (frontend)
- `NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS` (optional frontend allowlist)
- `MONGODB_URI`, `MONGODB_DATABASE`, `VENDOR_URLS`, `PORT` (API runtime)

For `mise run deploy-local`, values are read from `web_ui/.env.local` (or `WEB_ENV_FILE`) and synced into Pulumi config.

## Architecture Notes

- Vendor fan-out/fan-in aggregation is concurrent in Go.
- Orders/credits/events are persisted in MongoDB.
- Users are persisted in MongoDB (`users` collection with unique indexes on `email` and `member_id`).
- Deployment code is split into reusable constructs and stack composition:
  - `deployment/internal/constructs`
  - `deployment/internal/stacks/local_stack.go`

## Testing

Validated command set:

```bash
cd coordination_api && go test ./internal/http ./test/behavior
cd web_ui && pnpm exec next typegen && pnpm exec tsc --noEmit
cd deployment && go test ./...
```

Known caveat:
- `cd coordination_api && go test ./...` currently includes existing adapter tests that expect `../../../../mock/pizza_place_db.json`. This is a pre-existing fixture-path issue outside the auth migration.
