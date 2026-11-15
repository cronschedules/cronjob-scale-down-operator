# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.1] - 2026-11-14

### Security
- Fixed all 89 vulnerabilities reported by the Artifact Hub (Trivy) scan of image 0.5.0
  (1 critical, 37 high, 43 medium, 6 low, 2 unknown), all in the Go binary:
  - Go toolchain 1.25.0 → 1.26.11 (standard library, including CVE-2025-68121 in `crypto/tls`).
  - `golang.org/x/net` 0.52.0 → 0.62.0, `golang.org/x/text` 0.35.0 → 0.44.0, `golang.org/x/sys` 0.42.0 → 0.50.0.
  - `google.golang.org/grpc` 1.80.0 → 1.85.0, OpenTelemetry 1.43.0 → 1.47.0.
  - `github.com/google/cel-go` 0.22.0 → 0.29.2, which required upgrading controller-runtime 0.20.0 → 0.25.2
    and the Kubernetes libraries 0.32 → 0.37.

### Changed
- Minimum Go version is now 1.26 (`go 1.26.0`, `toolchain go1.26.11`); builder image `golang:1.26.11`.
- controller-gen v0.17.1 → v0.23.0 and golangci-lint v2.1.6 → v2.15.0; envtest now runs against Kubernetes 1.37.
- API scheme registration uses `k8s.io/apimachinery` `runtime.SchemeBuilder` instead of the deprecated
  controller-runtime `scheme.Builder`.

### Fixed
- **Release pipeline**: releases are created with `GITHUB_TOKEN`, which never triggers a `release: published` run, so
  since 0.4.3 the multi-arch build and the `latest` tag were never updated (`latest` was stuck at 0.4.3) and
  versioned tags were single-arch images from `build.yaml`. The release workflow now builds and pushes the
  multi-arch (`linux/amd64`, `linux/arm64`) image and moves `latest` in the same run that creates the release.
  Runs are serialized, images are always built from the release's git tag, a run resumes publishing when a
  release's images are missing or not multi-arch, and `latest` only ever points at the newest stable release.
  Manual dispatches must name an existing release version (`latest` is no longer accepted).
- `build.yaml` now pushes a rolling `main` image instead of overwriting the released version tag on every commit.
- `make docker-buildx*` no longer ignores `docker buildx build` failures.

## [0.5.0] - 2026-11-12

### Security
- **Namespace confinement (breaking)**: A `CronJobScaleDown` can now only scale targets (`spec.targetRef.namespace`)
  and clean up resources (`spec.cleanupConfig.namespaces`) in its own namespace. Previously any user able to create a
  `CronJobScaleDown` could use the operator's cluster-wide service account to scale or delete resources in every
  namespace. Set the new `--allow-cross-namespace` flag to restore the previous behavior on trusted, single-tenant clusters.
- **Cluster-scoped cleanup (breaking)**: Cleanup of `ClusterRole` and `ClusterRoleBinding` now requires the new
  `--allow-cluster-scoped-cleanup` flag.
- The policy is enforced both during spec validation and again right before scaling/deleting.
- **Web UI**: listens on `127.0.0.1:8082` by default (still reachable with `kubectl port-forward`) because it has no
  authentication; use `--webui-addr=:8082` to expose it or `--webui-addr=0` to disable it.
- **Web UI**: fixed stored XSS by HTML-escaping all resource fields in the dashboard; added `Content-Security-Policy`,
  `X-Frame-Options`, `X-Content-Type-Options` and `Referrer-Policy` headers; added Subresource Integrity to CDN assets;
  API errors no longer echo raw Kubernetes API messages; added HTTP server timeouts; target status is only read
  for targets in the resource's own namespace unless `--allow-cross-namespace` is set.
- **CRD validation**: `targetRef.name`/`namespace` and `cleanupConfig.namespaces` must be valid DNS names,
  `cleanupConfig.resourceTypes` is an enum, and schedules/timezone have length limits.
- **CI**: the release workflow no longer interpolates the `workflow_dispatch` tag input into shell and validates it.

### Fixed
- `dist/install.yaml` was regenerated; its CRD was missing the cleanup fields.
- Web UI no longer panics when a target Deployment/StatefulSet has no `spec.replicas`.

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
