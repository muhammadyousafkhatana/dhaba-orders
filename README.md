# Dhaba Orders

A Kubernetes-based restaurant ordering application built as a microservices project.

## Architecture

| Component | Technology | Responsibility |
|---|---|---|
| Web | React, Vite, Nginx | Customer ordering interface |
| Menu API | Python, FastAPI | Menu and item management |
| Orders API | Go | Order validation and creation |
| Kitchen Worker | Java, Spring Boot | Processes queued orders |
| Stats API | PHP, PDO | Order and revenue statistics |
| Database | PostgreSQL | Persistent application data |
| Queue/cache | Redis | Kitchen order queue |
| Monitoring | Shell scripts | Service health checks |

The Stats API is currently a PHP/PDO implementation; it is not yet a full Laravel application.

## Infrastructure

- AWS EC2 running Ubuntu
- Single-node K3s Kubernetes cluster
- Kubernetes namespaces: `staging` and `production`
- Kustomize base manifests and environment overlays
- Persistent volumes for PostgreSQL and Redis
- Container images published to GitHub Container Registry (GHCR)
- GitHub Actions for image builds and deployment
- PostgreSQL credentials supplied through Kubernetes Secret `dhaba-db`

This lab uses one EC2 node rather than the assignment's three-node cluster. It uses GHCR rather than AWS ECR.

## Application URLs

| Environment | Frontend |
|---|---|
| Staging | `http://EC2_PUBLIC_IP:30080` |
| Production | `http://EC2_PUBLIC_IP:30081` |

Restrict inbound security-group access to these ports to trusted client IP addresses.

## Deployment

### Prerequisites

- K3s and kubectl installed on EC2
- Git repository cloned to `~/dhaba-orders`
- Namespaces `staging` and `production`
- Kubernetes Secret `dhaba-db` in each namespace, containing key `password`
- GHCR images accessible to the cluster

Create the database Secret in each namespace if it does not already exist:

```bash
kubectl create secret generic dhaba-db -n staging \
  --from-literal=password='REPLACE_WITH_DATABASE_PASSWORD'

kubectl create secret generic dhaba-db -n production \
  --from-literal=password='REPLACE_WITH_DATABASE_PASSWORD'
```

Do not commit database passwords or private keys.

### Manual deployment

```bash
kubectl apply -k k8s/overlays/staging
kubectl apply -k k8s/overlays/production
```

Wait for deployments:

```bash
kubectl get deployments,pods,svc,pvc -n staging
kubectl get deployments,pods,svc,pvc -n production
```

### GitHub Actions deployment

Configure these repository Actions secrets:

- `EC2_HOST`: current EC2 public IPv4 address or DNS name
- `EC2_USER`: `ubuntu`
- `EC2_SSH_KEY`: complete SSH private key for the EC2 instance

The private key belongs in GitHub Secrets, never in the repository.

Workflow behavior:

1. A push to `main` builds and publishes the five application images.
2. After all builds succeed, the workflow deploys the updated manifests to staging over SSH.
3. Production is deployed separately through **Actions → Deploy production → Run workflow**.
4. Configure a GitHub Environment named `production` with required reviewers if an approval gate is available for the repository.

The workflows use the `latest` image tag. For stronger release traceability, immutable commit-SHA image tags should be introduced in a future improvement.

## Verification

Health endpoints:

- Menu API: `/health`
- Orders API: `/health`
- Stats API: `/health`

Example in-cluster checks:

```bash
kubectl get pods -n staging
kubectl get pods -n production
kubectl logs deployment/kitchen-worker -n staging --tail=30
```

The staging and production smoke tests completed successfully during project validation:

- All seven deployments became Ready in both namespaces.
- Both frontend NodePorts returned HTTP 200.
- PostgreSQL initialized the application tables.
- Menu API returned the five seeded menu items.
- An order was created, queued through Redis and processed by the kitchen worker.
- The test order reached `READY` in PostgreSQL.
- Stats API reported the expected order count and revenue.

## Operational notes and limitations

- This is a single-node lab cluster, not a highly available production cluster.
- Local-path PVCs are tied to the EC2 node; they are not a substitute for replicated storage or backups.
- The PostgreSQL container uses an unconfined AppArmor profile because the host's enforced profile blocked required PostgreSQL process signals. This is a lab-specific security trade-off and should be replaced with a suitable restrictive profile for production.
- The PHP stats service should be migrated to a complete Laravel application if strict Laravel compliance is required.
- The current monitoring scripts provide basic health checks; centralized metrics, alerting and Slack integration remain future improvements.
- A solo project cannot provide independent peer review. Production approval can be configured through GitHub Environment protection where supported.

## Repository layout

```text
web/
menu-api/
orders-api/
kitchen-worker/
stats-api/
ops-monitor/
scripts/monitor/
database/init/
k8s/base/
k8s/overlays/staging/
k8s/overlays/production/
.github/workflows/
```
