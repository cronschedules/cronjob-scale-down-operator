# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.1] - 2025-08-10

### Added
- **Prometheus Metrics**: Comprehensive metrics support for monitoring and observability
  - Scale operation metrics (up/down) with target resource details
  - Resource cleanup metrics by type and label
  - Reconciliation metrics with error tracking
  - Orphan resource cleanup tracking
  - Active resource counts and replica status
  - Schedule execution timestamps
  - Configuration validation error tracking
  - Metrics exposed on `:8080/metrics` endpoint
  - Grafana dashboard example and PromQL queries
  - Detailed metrics documentation

### Security
- **Dependency Updates**: Updated vulnerable dependencies to fix security issues
  - Updated `golang.org/x/oauth2` from v0.23.0 to v0.30.0 (fixes CVE-2025-22868 - HIGH)
  - Updated `golang.org/x/net` from v0.30.0 to v0.43.0 (fixes CVE-2025-22872, CVE-2025-22870 - MEDIUM)
  - Updated Go version from 1.23.0 to 1.23.12 (fixes CVE-2025-47907)
  - Updated various other golang.org/x packages to latest secure versions

### Fixed
- **Error Handling**: Improved graceful handling of missing target resources
  - Target resources not found errors now log as INFO instead of ERROR with stack traces
  - Cleaner logs when scaling/cleanup operations encounter missing resources
  - Better user experience in production environments

## [0.3.0] - 2025-07-22

### Added
- **Cleanup-Only Mode**: New pure cleanup functionality without requiring target scaling resources
  - Dedicated cleanup-only CronJobScaleDown resources
  - Perfect for CI/CD pipelines, test resource cleanup, and cost optimization
  - No `targetRef` required for cleanup-only operations
- **Enhanced Web UI**: 
  - Support for cleanup-only resources with special UI elements
  - "Cleanup Only" status badges and indicators
  - Cleanup schedule display and last cleanup timestamps
  - Graceful handling of missing target resources without crashes
- **Improved Error Handling**:
  - Controller gracefully handles missing target resources
  - Web UI continues to function when target deployments are not found
  - Enhanced logging with appropriate error levels
- **Enhanced RBAC**: Added permissions for cleanup operations on ConfigMaps, Secrets, and Services
- **Documentation**: 
  - New dedicated cleanup documentation (`docs/cleanup.md`)
  - Updated examples with cleanup-only configurations
  - Enhanced README with cleanup-only mode explanations

### Fixed
- **Web UI JavaScript Errors**: Fixed "targetStatus is undefined" errors when displaying cleanup-only resources
- **Null Pointer Handling**: Added proper null checks for optional fields in web UI
- **Missing Target Resources**: Controller and web UI now handle missing target resources gracefully

### Changed
- **API Response Format**: Added `isCleanupOnly` field to distinguish cleanup-only resources
- **Web UI Design**: Enhanced visual distinction between scaling and cleanup-only resources

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
