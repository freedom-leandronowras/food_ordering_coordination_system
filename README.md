## Overview
In an office setting, different companies have their innovation departments in the same physical location. Managers ensure everything runs smoothly, but there's no coordination for food ordering.

This system integrates food vendors and allows users to make and audit orders.

## Setup
 
### Prerequisites
 
| Tool | Version | Install |
|---|---|---|
| Docker | any recent | [docs.docker.com](https://docs.docker.com/get-docker/) |
| Node.js | 22+ | [nodejs.org](https://nodejs.org/) or `nvm install 22` |
| pnpm | 10+ | `corepack enable && corepack prepare pnpm@latest --activate` |
| Go | 1.24+ | [go.dev/dl](https://go.dev/dl/) |
| air | latest | `go install github.com/air-verse/air@latest` |
| mise | latest | `curl https://mise.jdx.dev/install.sh \| sh` |
| curl | any | Pre-installed on most systems |
 
### First-time setup
 
Run one command from the repository root:

```bash
mise run first-time-setup
```

This task:
- installs `web_ui` dependencies (only when `web_ui/node_modules` is missing)
- downloads `coordination_api` Go module dependencies
- creates `web_ui/.env.local` with local defaults only when the file is missing or empty
 
## Local Development (Live Reload)
 
Run everything locally with live reload:
 
```bash
mise run dev-local
```
 
What starts:
 
| Service | URL |
|---|---|
| Web UI | http://127.0.0.1:3000 |
| API | http://127.0.0.1:8080 |
| Vendor mock (pizza) | http://127.0.0.1:4011 |
| Vendor mock (sushi) | http://127.0.0.1:4012 |
| Vendor mock (taco) | http://127.0.0.1:4013 |
| MongoDB | 127.0.0.1:27017 (container `food-ordering-mongo-dev`) |
 
## Core API Endpoints
 
- `GET /api/menus`
- `GET /api/vendors`
- `POST /api/orders`
- `GET /api/members/{memberId}/orders`
- `GET /api/members/{memberId}/credits`
- `POST /api/members/{memberId}/credits` (manager roles only)
 
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
 
