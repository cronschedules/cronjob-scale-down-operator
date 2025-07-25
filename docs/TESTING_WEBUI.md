# Web UI Testing

Test the Web UI feature locally or in a cluster.

## Prerequisites

- Kubernetes cluster
- kubectl configured
- Go 1.23+ (for local development)

## Local Development

```bash
# Build and run locally
go build -o bin/manager ./cmd/main.go
kubectl apply -f config/crd/bases/
./bin/manager --webui-addr=:8082
```

```bash
# Create test resources
kubectl apply -f examples/quick-test.yaml
```

Access: http://localhost:8082

## Helm Installation

```bash
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update
helm install test-operator cronschedules/cronjob-scale-down-operator \
  --set webui.enabled=true
```

```bash
# Port forward to access
kubectl port-forward deployment/test-operator-cronjob-scale-down-operator 8082:8082
```

Access: http://localhost:8082

## Configuration

### Custom Port

```bash
# Local
./bin/manager --webui-addr=:8080

# Helm
helm install test-operator cronschedules/cronjob-scale-down-operator \
  --set webui.service.port=8080
```

### Production Ingress

```yaml
webui:
  enabled: true
  ingress:
    enabled: true
    hosts:
      - host: cronjob-ui.example.com
        paths:
          - path: /
            pathType: Prefix
```

## API Endpoints

```bash
# List all cron jobs
curl http://localhost:8082/api/v1/cronjobs

# Get specific cron job
curl http://localhost:8082/api/v1/cronjobs/default/job-name
```

## Troubleshooting

**Web UI not loading:**
- Check operator logs: `kubectl logs deployment/your-operator`
- Verify port availability: `lsof -i :8082`

**No data displayed:**
- Verify resources exist: `kubectl get cronjobscaledown`
- Check RBAC permissions

**Auto-refresh issues:**
- Check browser console for errors
- Clear browser cache
