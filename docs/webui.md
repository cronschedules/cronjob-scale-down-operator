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

Default: http://localhost:8082

### Custom Port

```bash
./bin/manager --webui-addr=:8080
```

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
