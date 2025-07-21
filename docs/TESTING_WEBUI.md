# Testing the Web UI

This guide helps you quickly test the new Web UI feature of the CronJob Scale Down Operator.

## Prerequisites

- Kubernetes cluster (minikube, kind, or any cluster)
- kubectl configured
- Go 1.23+ installed

## Quick Test Setup

### 1. Build and Run the Operator Locally

```bash
# Clone and build
git clone https://github.com/z4ck404/cronjob-scale-down-operator.git
cd cronjob-scale-down-operator
go build -o bin/manager ./cmd/main.go

# Install CRDs in your cluster
kubectl apply -f config/crd/bases/

# Run locally with web UI enabled
./bin/manager --webui-addr=:8082
```

### 2. Create Test Resources

In another terminal, create some test resources:

```bash
# Create a test deployment and CronJobScaleDown
kubectl apply -f examples/webui-demo.yaml

# Or use the quick test example
kubectl apply -f examples/quick-test.yaml
```

### 3. Access the Web UI

Open your browser and navigate to: http://localhost:8082

You should see:
- A modern dashboard with glassmorphic design and gradients
- Company logo in the navigation bar
- Status information with color-coded badges (Ready/Not Ready/Scaled Down)
- Schedule information with monospace cron expressions
- Real-time updates every 30 seconds with smooth transitions
- Responsive design that works on all screen sizes

### 4. Test the API Directly

You can also test the REST API endpoints directly:

```bash
# Get all cron jobs
curl http://localhost:8082/api/v1/cronjobs

# Get specific cron job
curl http://localhost:8082/api/v1/cronjobs/default/your-cronjob-name
```

## Test with Helm

Alternatively, you can test using Helm:

```bash
# Install with Web UI enabled
helm install test-operator ./charts/cronjob-scale-down-operator \
  --set webui.enabled=true

# Port forward to access the web UI
kubectl port-forward deployment/test-operator-cronjob-scale-down-operator 8082:8082

# Access at http://localhost:8082
```

## Customizing the Web UI

### Change the Port

```bash
# Run on different port
./bin/manager --webui-addr=:8080

# Or in Helm
helm install test-operator ./charts/cronjob-scale-down-operator \
  --set webui.service.port=8080
```

### Enable Ingress (Production)

```yaml
# values.yaml
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

## Troubleshooting

### Web UI Not Loading
1. Check operator logs: `kubectl logs deployment/your-operator`
2. Verify port is not in use: `lsof -i :8082`
3. Check firewall settings

### No Data in Dashboard
1. Verify CronJobScaleDown resources exist: `kubectl get cronjobscaledown`
2. Check RBAC permissions
3. Look for API errors in browser console

### Auto-refresh Not Working
1. Check browser console for errors
2. Verify network connectivity
3. Clear browser cache

## Development Notes

- Web UI files are in `web/static/`
- Backend API is in `internal/webui/server.go`
- Uses Bootstrap 5 and Font Awesome for styling
- Auto-refresh interval is 30 seconds (configurable in JS)

For more detailed information, see the [Web UI Documentation](./docs/webui.md).
