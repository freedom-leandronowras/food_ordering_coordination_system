# Local Kubernetes Deployment

This folder contains a Pulumi stack that deploys the full Food Ordering Coordination System to a local Kubernetes cluster.

## Architecture

- `internal/constructs/`
  - `workload.go`: reusable Deployment + Service construct
  - `mongodb.go`: MongoDB construct built on top of `workload.go`
- `internal/stacks/`
  - `local_stack.go`: composes MongoDB, vendor mocks, API, and web UI
- `main.go`
  - Pulumi entrypoint for the local stack

## One-command local deploy

From repository root:

```bash
cd deployment && mise run deploy-local
```

The `mise` task will:
1. Build images from each service Dockerfile.
2. Configure Pulumi stack values.
3. Deploy resources to local Kubernetes.
4. Run rollout checks.
5. Run smoke tests using temporary `kubectl port-forward` sessions.

## Required web auth values

`deployment/mise.toml` task `deploy-local` reads these from `web_ui/.env.local` (or `WEB_ENV_FILE`):

- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
- `CLERK_SECRET_KEY`

## Optional env overrides

- `PULUMI_STACK` (default: `local`)
- `K8S_NAMESPACE` (default: `food-ordering`)
- `WEB_API_BASE_URL` (default: `http://127.0.0.1:18081`)
- `PULUMI_BACKEND_URL` (default: local file backend under `deployment/.pulumi`)
- `PULUMI_CONFIG_PASSPHRASE` (default: `local-dev-passphrase`)
