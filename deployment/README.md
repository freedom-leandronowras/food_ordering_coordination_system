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
mise run deploy-local
```

The `mise` task will:
1. Build images from each service Dockerfile.
2. Configure Pulumi stack values.
3. Deploy resources to local Kubernetes.
4. Run rollout checks.
5. Run smoke tests using temporary `kubectl port-forward` sessions.

## Atomic tasks

Run each phase independently:

```bash
mise tasks
```

Main atomic tasks:

- `preflight` - Validate local prerequisites.
- `kube-check` - Validate cluster connectivity.
- `build-api-image`
- `build-vendor-image`
- `build-web-image`
- `build-images` - Runs all image builds.
- `pulumi-login`
- `pulumi-stack`
- `pulumi-config`
- `pulumi-up`
- `rollout` - Wait for all deployments.
- `smoke` - Temporary port-forward + endpoint checks.
- `status` - Show deployments/pods/services.
- `port-forward-api`
- `port-forward-web`

## Local live development

Run the full stack with live reload on localhost:

```bash
mise run dev-local
```

This keeps running and serves:

- Web UI: `http://127.0.0.1:3000`
- API: `http://127.0.0.1:8080/api/vendors`

Stop the dev Mongo container when you no longer need it:

```bash
mise run dev-local-stop
```

## Required web auth values

Root `mise.toml` task `deploy-local` reads these from `web_ui/.env.local` (or `WEB_ENV_FILE`):

- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
- `CLERK_SECRET_KEY`

## Optional env overrides

- `PULUMI_STACK` (default: `local`)
- `K8S_NAMESPACE` (default: `food-ordering`)
- `WEB_API_BASE_URL` (default: `http://127.0.0.1:18081`)
- `PULUMI_BACKEND_URL` (default: local file backend under `deployment/.pulumi`)
- `PULUMI_CONFIG_PASSPHRASE` (default: `local-dev-passphrase`)
