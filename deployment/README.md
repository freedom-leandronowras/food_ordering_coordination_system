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

## One-command production deploy (EC2 Kubernetes)

From repository root:

```bash
PROD_API_IMAGE=<registry>/food-coordination-api:<tag> \
PROD_WEB_IMAGE=<registry>/food-web-ui:<tag> \
PROD_VENDOR_IMAGE=<registry>/food-vendor-mocks:<tag> \
PROD_WEB_API_BASE_URL=https://<api-domain> \
PROD_JWT_SIGNING_KEY=<strong-secret> \
mise run deploy-production
```

The production pipeline will:
1. Build images from each service Dockerfile.
2. Push images to your registry.
3. Configure Pulumi production stack values.
4. Deploy resources to Kubernetes with `pulumi up`.
5. Run rollout checks.
6. Run a smoke check against `${PROD_WEB_API_BASE_URL}/api/vendors`.

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
- `prod-preflight` - Validate production deployment env + tools.
- `kube-check-prod` - Validate production cluster connectivity.
- `build-images-prod` - Build all production images.
- `push-images-prod` - Push all production images.
- `pulumi-stack-prod`
- `pulumi-config-prod`
- `pulumi-up-prod`
- `rollout-prod`
- `smoke-prod` - Check production API endpoint.
- `status-prod`
- `deploy-production` - Runs production pipeline end-to-end.

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

## Required auth values

Root `mise.toml` task `deploy-local` reads these from `web_ui/.env.local` (or `WEB_ENV_FILE`):

- `JWT_SIGNING_KEY`
- `NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS` (optional)

## Optional env overrides

- `PULUMI_STACK` (default: `local`)
- `K8S_NAMESPACE` (default: `food-ordering`)
- `WEB_API_BASE_URL` (default: `http://127.0.0.1:18081`)
- `PULUMI_BACKEND_URL` (default: local file backend under `deployment/.pulumi`)
- `PULUMI_CONFIG_PASSPHRASE` (default: `local-dev-passphrase`)
