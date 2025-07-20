# CronJob Scale Down Operator Helm Chart

This chart installs the CronJob Scale Down Operator on a Kubernetes cluster using the Helm package manager.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.0+

## Installation

### Add Helm Repository (if available)

```bash
helm repo add cronjob-scale-down-operator https://z4ck404.github.io/cronjob-scale-down-operator
helm repo update
```

### Install from Local Chart

```bash
# Clone the repository
git clone https://github.com/z4ck404/cronjob-scale-down-operator.git
cd cronjob-scale-down-operator

# Install the chart
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator
```

### Install with Custom Values

```bash
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator \
  --set image.tag=0.1.2 \
  --set resources.requests.memory=128Mi
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of operator replicas | `1` |
| `image.repository` | Container image repository | `ghcr.io/z4ck404/cronjob-scale-down-operator` |
| `image.tag` | Container image tag | `0.1.2` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `serviceAccount.create` | Create service account | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `metrics.enabled` | Enable metrics service | `true` |
| `leaderElection.enabled` | Enable leader election | `true` |

## Usage

After installation, create a CronJobScaleDown resource:

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: my-scaler
  namespace: default
spec:
  targetRef:
    name: my-deployment
    namespace: default
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "0 0 22 * * *"  # 10 PM daily
  scaleUpSchedule: "0 0 6 * * *"     # 6 AM daily
  timeZone: "UTC"
```

## Uninstallation

```bash
helm uninstall cronjob-scale-down-operator
```

## Development

To render templates locally:

```bash
helm template cronjob-scale-down-operator ./charts/cronjob-scale-down-operator
```

To validate the chart:

```bash
helm lint ./charts/cronjob-scale-down-operator
```
