# Web UI

Built-in web dashboard to monitor CronJobScaleDown resources.

## Features

- View all CronJobScaleDown resources
- Monitor target deployment/statefulset status
- View schedules and timezones
- Track replica counts and scaling history
- Auto-refresh every 30 seconds
- Responsive design

## Access

Default listen address: `127.0.0.1:8082` (loopback only). The dashboard has no authentication and shows
resources from every namespace, so reach it with `kubectl port-forward ... 8082:8082` and open
http://localhost:8082.

### Custom Address

```bash
./bin/manager --webui-addr=:8082   # expose on all interfaces (put behind an authenticating proxy / NetworkPolicy)
./bin/manager --webui-addr=0       # disable the web UI
```

### Security

- Responses carry a strict `Content-Security-Policy`, `X-Frame-Options: DENY` and `X-Content-Type-Options: nosniff`.
- All resource fields are HTML-escaped before rendering.
- Target workload status is only shown for targets in the resource's own namespace unless the operator runs with `--allow-cross-namespace`.

### Helm Configuration

```yaml
webui:
  enabled: true
  service:
    port: 8080
```

## API Endpoints

### GET /api/v1/cronjobs

Returns all CronJobScaleDown resources with status.

**Response:**
```json
[
  {
    "name": "example-cronjob",
    "namespace": "default",
    "targetRef": {
      "name": "my-deployment",
      "namespace": "default", 
      "kind": "Deployment",
      "apiVersion": "apps/v1"
    },
    "scaleDownSchedule": "0 0 22 * * *",
    "scaleUpSchedule": "0 0 8 * * *",
    "timezone": "UTC",
    "currentReplicas": 3,
    "desiredReplicas": 3,
    "lastScaleDown": "2025-01-15T22:00:00Z",
    "lastScaleUp": "2025-01-16T08:00:00Z"
  }
]
```

### GET /api/v1/cronjobs/{namespace}/{name}

Returns specific CronJobScaleDown resource.

## Development

- UI files: `web/static/`
- Backend: `internal/webui/server.go`
- Framework: Bootstrap 5 + Font Awesome
