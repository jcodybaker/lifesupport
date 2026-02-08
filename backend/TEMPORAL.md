# Temporal Worker Integration

This document describes the Temporal worker integration in the Life Support backend.

## Overview

The backend has been restructured to use the Cobra CLI tool, providing two main commands:
- `lifesupport-backend http` - Runs the HTTP API server
- `lifesupport-backend worker` - Runs the Temporal worker

Both services run from the same binary but can be deployed independently.

## Building

```bash
go build -o lifesupport-backend
```

## Running the HTTP Server

Start the HTTP API server:

```bash
./lifesupport-backend http
```

With custom port:

```bash
./lifesupport-backend http --port 3000
```

## Running the Temporal Worker

Start the Temporal worker:

```bash
./lifesupport-backend worker
```

With custom configuration:

```bash
./lifesupport-backend worker \
  --temporal-host localhost:7233 \
  --task-queue lifesupport-tasks \
  --max-concurrent-activities 20 \
  --max-concurrent-workflows 10
```

## Environment Variables

### HTTP Server
- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - HTTP server port (overridden by `--port` flag)

### Temporal Worker
- `TEMPORAL_HOST` - Temporal server address (overridden by `--temporal-host` flag)
- `TEMPORAL_TASK_QUEUE` - Task queue name (overridden by `--task-queue` flag)

## Temporal Setup

To run the Temporal worker, you need a running Temporal server. You can start one locally using Docker:

```bash
# Start Temporal server with docker-compose
docker-compose -f temporal-docker-compose.yml up -d
```

Or use Temporal's development server:

```bash
temporal server start-dev
```

## Adding Workflows and Activities

Workflows and activities are defined in `pkg/worker/workflows.go`. 

### Example Workflow

```go
func MyWorkflow(ctx workflow.Context, input string) (string, error) {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity, input).Get(ctx, &result)
    return result, err
}
```

### Example Activity

```go
func MyActivity(ctx context.Context, input string) (string, error) {
    log.Printf("Processing: %s", input)
    return "processed: " + input, nil
}
```

### Registering Workflows/Activities

Add your workflow/activity to the registration functions in `pkg/worker/workflows.go`:

```go
func RegisterWorkflows(w interface{}) {
    w.(interface{ RegisterWorkflow(interface{}, ...interface{}) }).RegisterWorkflow(SampleWorkflow)
    w.(interface{ RegisterWorkflow(interface{}, ...interface{}) }).RegisterWorkflow(MyWorkflow)
}

func RegisterActivities(w interface{}) {
    w.(interface{ RegisterActivity(interface{}, ...interface{}) }).RegisterActivity(SampleActivity)
    w.(interface{ RegisterActivity(interface{}, ...interface{}) }).RegisterActivity(MyActivity)
}
```

## Deployment

### Docker Deployment

You can deploy the HTTP server and worker as separate containers:

```dockerfile
# Dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o lifesupport-backend

FROM debian:bookworm-slim
COPY --from=builder /app/lifesupport-backend /usr/local/bin/
ENTRYPOINT ["lifesupport-backend"]
CMD ["http"]
```

Deploy HTTP server:
```bash
docker run -e DATABASE_URL="..." lifesupport-backend http
```

Deploy worker:
```bash
docker run -e TEMPORAL_HOST="temporal:7233" lifesupport-backend worker
```

### Kubernetes Deployment

Create separate deployments for the HTTP server and worker:

```yaml
# http-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lifesupport-http
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: lifesupport-http
        image: lifesupport-backend:latest
        command: ["lifesupport-backend", "http"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: url

---
# worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lifesupport-worker
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: lifesupport-worker
        image: lifesupport-backend:latest
        command: ["lifesupport-backend", "worker"]
        env:
        - name: TEMPORAL_HOST
          value: "temporal.default.svc.cluster.local:7233"
```

## Monitoring

The worker includes built-in metrics and logging. Check the logs for:

```
Starting Temporal worker on task queue: lifesupport-tasks
Connected to Temporal server at: localhost:7233
```

## Testing

To test workflows locally, use Temporal's test suite (see Temporal SDK documentation for details).

You can also trigger workflows manually using the Temporal CLI:

```bash
temporal workflow start \
  --task-queue lifesupport-tasks \
  --type SampleWorkflow \
  --input '"test-name"'
```
