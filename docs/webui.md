# Web UI Documentation

The CronJob Scale Down Operator now includes a built-in web UI that provides a dashboard to monitor and view the status of all CronJobScaleDown resources and their target deployments/statefulsets.

## Features

The web UI provides the following features:

- **Real-time Dashboard**: View all CronJobScaleDown resources in a clean, organized interface
- **Status Monitoring**: Monitor the current state of target deployments and statefulsets
- **Schedule Information**: View scale-up and scale-down schedules with timezone information
- **Replica Status**: Visual indicators showing ready vs desired replicas
- **Action History**: Last scale-up and scale-down timestamps
- **Auto-refresh**: Automatically updates data every 30 seconds
- **Responsive Design**: Works on desktop, tablet, and mobile devices

## Access

The web UI is accessible at `http://localhost:8082` by default. You can customize the port using the `--webui-addr` flag when starting the operator.

### Starting with Custom Web UI Address

```bash
./bin/manager --webui-addr=:8080
```

## UI Components

### Navigation Bar
- Shows the operator name and branding
- Displays the last update timestamp
- Indicates auto-refresh interval

### Dashboard Cards
Each CronJobScaleDown resource is displayed as a card containing:

1. **Resource Information**
   - Resource name and namespace
   - Target resource (kind, name, namespace)

2. **Current Status**
   - Ready/Not Ready indicator with visual badges
   - Replica count with progress bar visualization
   - Available vs desired replicas

3. **Schedule Configuration**
   - Scale-down cron schedule
   - Scale-up cron schedule  
   - Configured timezone

4. **Action History**
   - Timestamp of last scale-down action
   - Timestamp of last scale-up action

### Interactive Elements
- **Refresh Button**: Manual refresh of data
- **Auto-refresh**: Automatically updates every 30 seconds
- **Responsive Layout**: Adapts to different screen sizes

## API Endpoints

The web UI is backed by REST API endpoints:

### GET /api/v1/cronjobs
Returns a list of all CronJobScaleDown resources with their current status.

**Response Format:**
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
    "scaleDownSchedule": "0 22 * * *",
    "scaleUpSchedule": "0 6 * * *",
    "timeZone": "UTC",
    "lastScaleDownTime": "2025-01-20T22:00:00Z",
    "lastScaleUpTime": "2025-01-21T06:00:00Z",
    "currentReplicas": 3,
    "targetStatus": {
      "ready": true,
      "desiredReplicas": 3,
      "availableReplicas": 3,
      "readyReplicas": 3,
      "lastUpdateTime": "2025-01-21T10:30:00Z"
    }
  }
]
```

### GET /api/v1/cronjobs/{namespace}/{name}
Returns details for a specific CronJobScaleDown resource.

## Security Considerations

- The web UI runs on a separate port from the metrics endpoint
- No authentication is currently implemented - consider using a reverse proxy with authentication in production
- The UI only provides read-only access to resource information
- CORS is not explicitly configured - restrict access appropriately in production environments

## Troubleshooting

### Web UI Not Accessible
1. Check that the operator is running: `kubectl get pods -n cronjob-scale-down-operator-system`
2. Verify the web UI port is not blocked by firewall
3. Check operator logs for any web UI startup errors

### Empty Dashboard
1. Verify CronJobScaleDown resources exist: `kubectl get cronjobscaledown --all-namespaces`
2. Check operator logs for any API errors
3. Ensure the operator has proper RBAC permissions

### Auto-refresh Not Working
1. Check browser console for JavaScript errors
2. Verify network connectivity to the operator
3. Check if the browser is blocking the requests

## Development

The web UI consists of:
- **Backend**: Go HTTP server in `internal/webui/server.go`
- **Frontend**: HTML/CSS/JavaScript in `web/static/`
- **Styling**: Bootstrap 5 with custom CSS
- **Icons**: Font Awesome 6

### Adding New Features
1. Add API endpoints in `server.go`
2. Update the JavaScript in `dashboard.js`
3. Modify styles in `styles.css` as needed
4. Test the changes by rebuilding and running the operator

### API Extensions
To add new API endpoints:
1. Add route in `setupRoutes()` method
2. Implement handler function
3. Update frontend JavaScript to consume new endpoints

## UI Screenshots

### Dashboard Overview
The new dashboard features a modern, glassmorphic design with:

- **Modern Design**: Glass morphism effects with gradients and blur effects
- **Logo Integration**: Company logo prominently displayed in the navigation
- **Enhanced Status Display**: Clear visual indicators for ready/not ready/scaled down states
- **Improved Replica Visualization**: Color-coded progress bars showing replica health
- **Better Typography**: Inter font for improved readability
- **Responsive Layout**: Works beautifully on all device sizes

### Key Improvements Made

1. **Visual Design**:
   - Modern glassmorphic design with backdrop blur effects
   - Gradient backgrounds and enhanced color scheme
   - Logo integration in navigation bar
   - Custom favicon with operator branding

2. **Status Indicators**:
   - Green badges for ready resources
   - Red badges for not ready resources  
   - Orange badges for scaled down resources
   - Color-coded replica progress bars

3. **Enhanced UX**:
   - Smooth animations and transitions
   - Loading states with opacity transitions
   - Staggered card animations on load
   - Improved hover effects

4. **Information Architecture**:
   - Cleaner section dividers
   - Better typography hierarchy
   - Improved spacing and layout
   - Monospace font for cron expressions
