# transaction-store

## Plan
- Write server in Go
- Multi-stage Dockerfile for optimized builds
- Consider security with mTLS
- Interface-based design for database connectivity
- Include Kubernetes manifests

### Observability

- Add Prometheus metrics endpoint to Go service
- Include metrics like request latency, request count, DB connection pool stats
- Include a Grafana dashboard definition
- Add structured logging

### Infrastructure as Code (IaC):

- Add Terraform configs for the infrastructure pieces
- Demonstrate with a local setup using kind/k3d
- Include resource requests/limits in K8s manifests

### CI/CD:

- Add GitHub Actions workflow to:
- Run tests
- Build and push Docker images
- Deploy to K8s (use k3d for demonstration)
- Run integration tests

### For the K8s manifests:

- ConfigMaps/Secrets for configuration
- Health check probes
- Pod disruption budget
- HorizontalPodAutoscaler
- ServiceAccount and RBAC rules if needed

### Database considerations:

- Add database migrations strategy
- Consider backup/restore procedures
- Add persistence configuration for Postgres

### Documentation:
- Architecture diagram
- API documentation
- Runbook with common operations

### Other considerations
- Use self-signed certs but document how it would be done in production (e.g., cert-manager)
- Document why specific pool sizes for db connections

### Todos
- Make an interface for logger?

### Final checks
- Is dockerfile multi-stage?
- Are there unit tests?

## How to run

### Get up and running quickly with docker compose

Run:
```
docker-compose up --build
```

This will spin up the server as well as a postgres instance. You can then send requests to the server like so:
```
$ curl -X POST localhost:8080/api/transaction/ -d '{"transactionId":"0f7e46df-c685-4df9-9e23-e75e7ac8ba7a","amount": "99.99","timestamp":"2009-09-28T19:03:12Z"}'
{"id":"0f7e46df-c685-4df9-9e23-e75e7ac8ba7a","status":"success"}
```

or check the server's health:
```
$ curl localhost:8080/health
{"status":"healthy"}
```
### Deploy application into Kubernetes cluster

The application can be deployed into a local k3d cluster using the supplied Makefile. To do so, run:
```
make
```
You can then port-forward the service with:
```
kubectl port-forward deploy/tx-server 8080
```
and reach the API as before with, for example,
```
$ curl -X POST localhost:8080/api/transaction/ -d '{"transactionId":"0f7e46df-c685-4df9-9e23-e75e7ac8ba7b","amount": "99.99","timestamp":"2009-09-28T19:03:12Z"}'
{"id":"0f7e46df-c685-4df9-9e23-e75e7ac8ba7a","status":"success"}
```

To delete the cluster, run:
```
make delete-local-kube-cluster
```
