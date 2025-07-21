# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-01-21

### Added
- **Web UI Dashboard**: New built-in web interface for monitoring CronJobScaleDown resources
  - Real-time dashboard showing all cron jobs and their status
  - Visual status indicators for target deployments and statefulsets  
  - Schedule information with timezone display
  - Replica status with progress bars
  - Action history showing last scale operations
  - Auto-refresh every 30 seconds
  - Responsive design for desktop, tablet, and mobile
  - REST API endpoints at `/api/v1/cronjobs`
- **Helm Chart Updates**: 
  - Added web UI service configuration
  - Optional ingress support for web UI
  - Configurable web UI port (default: 8082)
- **Documentation**: 
  - Comprehensive web UI documentation (`docs/webui.md`)
  - Updated main README with web UI information
  - Added web UI demo example (`examples/webui-demo.yaml`)

### Changed
- **Command Line Arguments**: Added `--webui-addr` flag to configure web UI port
- **Container Ports**: Exposed additional port 8082 for web UI
- **Version**: Bumped to 0.2.0 to reflect new major feature

### Dependencies
- Added `github.com/gorilla/mux` for HTTP routing

## [0.1.2] - Previous Release

### Features
- Cron-based scheduling with second precision
- Timezone support using IANA timezone names
- Support for Deployments and StatefulSets
- Flexible scale-up and scale-down schedules
- Status tracking and monitoring
- Kubernetes operator with controller-runtime
- Helm chart for easy deployment
- Comprehensive examples and documentation
