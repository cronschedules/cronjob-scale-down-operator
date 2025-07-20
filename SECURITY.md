# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in this project, please report it responsibly:

1. **Do not open a public issue** for security vulnerabilities
2. Send details to the maintainers via private communication channels
3. Include details about the vulnerability and steps to reproduce
4. Allow reasonable time for the maintainers to respond and fix the issue

## Security Measures Implemented

### Container Security

- **Non-root user**: Container runs as non-root user (UID 65532)
- **Read-only root filesystem**: Container filesystem is read-only to prevent tampering
- **No privileged escalation**: `allowPrivilegeEscalation: false`
- **Dropped capabilities**: All Linux capabilities are dropped
- **Distroless base image**: Uses Google's distroless image for minimal attack surface

### Network Security

- **NetworkPolicy**: Restricts network traffic to only necessary communications
- **TLS encryption**: HTTPS endpoints for metrics and health checks
- **Principle of least privilege**: Only allows required network access

### Input Validation

- **Cron schedule validation**: Validates cron expressions to prevent injection attacks
- **Timezone validation**: Restricts timezone inputs to valid IANA timezone names
- **Resource reference validation**: Validates target resource references
- **Length limits**: Enforces maximum length limits on user inputs

### RBAC Security

- **Least privilege**: Operator only has permissions required for its function
- **Resource-specific access**: Limited to deployments and statefulsets
- **Namespace-scoped when possible**: Minimizes cluster-wide permissions

### Supply Chain Security

- **Pinned GitHub Actions**: All CI/CD actions are pinned to specific SHA hashes
- **Signed container images**: Images are built and signed in CI/CD
- **Dependency scanning**: Regular security scans of dependencies

### Runtime Security

- **Resource limits**: CPU and memory limits to prevent resource exhaustion
- **Health checks**: Liveness and readiness probes for container health
- **Graceful shutdown**: Proper signal handling for clean shutdowns

## Security Configuration Best Practices

### Deployment Security

1. **Use dedicated namespace**: Deploy the operator in its own namespace
2. **Enable Pod Security Standards**: Use restricted Pod Security Standards
3. **Network segmentation**: Implement network policies for traffic isolation
4. **Monitor logs**: Enable audit logging and monitor for suspicious activities

### Operational Security

1. **Regular updates**: Keep the operator and its dependencies updated
2. **Backup configurations**: Backup CRD definitions and configurations
3. **Access control**: Limit who can create/modify CronJobScaleDown resources
4. **Resource monitoring**: Monitor resource usage and scaling activities

### Configuration Security

1. **Validate inputs**: Always validate cron schedules and timezone inputs
2. **Limit permissions**: Use least privilege for target resources
3. **Secure storage**: Protect sensitive configuration data
4. **Regular audits**: Periodically review operator permissions and activities

## Compliance and Standards

This project aims to comply with:

- Kubernetes Pod Security Standards (Restricted)
- NIST Cybersecurity Framework
- CIS Kubernetes Benchmark
- OWASP Container Security Guidelines

## Security Updates

Security updates will be:

1. Released as quickly as possible after discovery
2. Documented in the CHANGELOG.md
3. Tagged with security advisory information
4. Communicated through GitHub security advisories

## Contact

For security-related questions or concerns, please contact the maintainers through the project's communication channels.