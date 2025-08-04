# CronJob Scale Down Operator - Copilot Instructions

This is a Go-based Kubernetes operator that automatically scales down Deployments and StatefulSets during specific time windows to save resources and costs. The operator is built using the Operator SDK and Kubebuilder framework.

## Project Overview

The CronJob Scale Down Operator provides:
- Cron-based scheduling with timezone support
- Flexible scaling for Deployments and StatefulSets
- Resource cleanup functionality
- Web UI dashboard for monitoring
- Multi-architecture Docker image support

## Code Standards

### Required Before Each Commit
- Run `make fmt` to format Go code using gofmt
- Run `make vet` to check for Go code issues
- Run `make lint` to run golangci-lint for code quality
- Run `make test` to ensure all tests pass

### Development Flow
- **Build**: `make build` - Build the manager binary
- **Test**: `make test` - Run unit tests with coverage
- **E2E Tests**: `make test-e2e` - Run end-to-end tests (requires Kind cluster)
- **Lint**: `make lint` - Run golangci-lint
- **Full CI**: `make manifests generate fmt vet test` - Complete CI pipeline
- **Docker Build**: `make docker-build` - Build container image
- **Docker Push**: `make docker-push-all` - Push to both GitHub Container Registry and Docker Hub

## Repository Structure

### Core Directories
- `api/v1/`: Kubernetes API definitions and CRD types
- `cmd/`: Main application entry point (`main.go`)
- `internal/controller/`: Controller logic and reconciliation
- `internal/utils/`: Utility functions and helper code
- `config/`: Kubernetes manifests and Kustomize configurations
  - `config/crd/`: Custom Resource Definitions
  - `config/rbac/`: Role-based access control
  - `config/manager/`: Operator deployment configuration
  - `config/samples/`: Example CronJobScaleDown resources

### Documentation & Examples
- `docs/`: Documentation website (Docsify-based)
- `examples/`: Example YAML files for different use cases
- `test/`: Test utilities and end-to-end tests

### Build & CI
- `Makefile`: Build automation and development commands
- `.github/workflows/`: GitHub Actions CI/CD pipelines
- `Dockerfile`: Multi-stage container build
- `hack/`: Development scripts and boilerplate

## Key Guidelines

### Go Development
1. **Follow Kubebuilder patterns**: Use controller-runtime best practices
2. **API Versioning**: All APIs are in `api/v1/` with proper versioning
3. **Controller Logic**: Keep reconciliation logic in `internal/controller/`
4. **Error Handling**: Use controller-runtime's Result and error patterns
5. **Logging**: Use logr structured logging throughout
6. **Testing**: Write table-driven tests, use envtest for controller testing

### Kubernetes Operator Patterns
1. **Reconciliation**: Implement idempotent reconcile loops
2. **Status Updates**: Always update resource status with current state
3. **Finalizers**: Use finalizers for cleanup operations
4. **Watches**: Set up proper controller watches for dependent resources
5. **RBAC**: Maintain minimal required permissions in `config/rbac/`

### Container Images
- **Multi-Registry Support**: Push to both `ghcr.io/cronschedules/cronjob-scale-down-operator` and `cronschedules/cronjob-scale-down-operator`
- **Multi-Architecture**: Support `linux/amd64` and `linux/arm64`
- **Semantic Versioning**: Use version from `VERSION` file
- **Security**: Use distroless base images for minimal attack surface

### Documentation
1. **Docsify Site**: Update `docs/` for user-facing documentation
2. **Code Comments**: Document public APIs and complex logic
3. **Examples**: Provide working examples in `examples/` directory
4. **Helm Charts**: Charts are maintained in separate repository `cronschedules/charts`

## Testing Guidelines

### Unit Tests
- Use `internal/controller/suite_test.go` for controller test suites
- Test files should end with `_test.go`
- Use Ginkgo/Gomega for BDD-style tests (following Kubebuilder convention)
- Mock external dependencies appropriately

### End-to-End Tests
- Located in `test/e2e/`
- Requires Kind cluster for testing
- Tests should verify complete operator functionality
- Use `make test-e2e` to run (ensures Kind cluster exists)

### Coverage
- Maintain test coverage above 80%
- Coverage reports generated in `cover.out`
- Run `make test` to generate coverage

## Configuration Management

### Environment Variables
- Follow 12-factor app principles
- Use sensible defaults in code
- Document all configurable options

### Kubernetes Resources
- **CRDs**: Defined in `api/v1/cronjobscaledown_types.go`
- **RBAC**: Minimal permissions in `config/rbac/`
- **Samples**: Working examples in `config/samples/`

### Helm Chart Integration
- Charts maintained in separate repository: `cronschedules/charts`
- Default image repository: `ghcr.io/cronschedules/cronjob-scale-down-operator`
- Support Docker Hub alternative: `cronschedules/cronjob-scale-down-operator`

## Build and Release Process

### Local Development
```bash
# Setup development environment
make manifests generate fmt vet

# Build and test
make build test

# Run locally (requires kubeconfig)
make run
```

### Container Images
```bash
# Build image
make docker-build IMG=myregistry/myimage:tag

# Push to registries
make docker-push-all IMG=myregistry/myimage:tag

# Multi-arch build
make docker-buildx-all PLATFORMS=linux/amd64,linux/arm64
```

### Release Process
1. Update `VERSION` file with semantic version
2. Push to main branch triggers CI/CD
3. Create GitHub release triggers multi-arch builds
4. Images pushed to both registries automatically

## Web UI Dashboard

The operator includes a built-in web UI dashboard:
- Located in controller code (embedded assets)
- Provides real-time monitoring of CronJobScaleDown resources
- Enable/disable via Helm chart values
- Default port: 8081

## Common Development Tasks

### Adding New Features
1. Update CRD types in `api/v1/cronjobscaledown_types.go`
2. Run `make manifests generate` to update generated code
3. Implement controller logic in `internal/controller/`
4. Add tests and examples
5. Update documentation

### Debugging
- Use `kubectl logs` to view operator logs
- Enable debug logging via environment variables
- Use `kubectl describe` for resource status details

### Performance Considerations
- Controllers should be efficient and not over-reconcile
- Use proper controller watches and predicates
- Consider rate limiting for external API calls
- Monitor resource usage in production

## Dependencies

### Go Modules
- Controller-runtime for Kubernetes controller framework
- Kubebuilder for scaffolding and code generation
- Ginkgo/Gomega for testing
- Prometheus client for metrics

### Development Tools
- `controller-gen`: Code and manifest generation
- `kustomize`: Kubernetes configuration management
- `golangci-lint`: Go code linting
- `setup-envtest`: Test environment setup

All dependencies are managed in `go.mod` and installed automatically during builds.
