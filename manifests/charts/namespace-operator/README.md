# namespace-operator

![Version: 1.3.0](https://img.shields.io/badge/Version-1.3.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.3.0](https://img.shields.io/badge/AppVersion-1.3.0-informational?style=flat-square)

A Helm chart for Kubernetes

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules |
| autoscaling | object | `{"enabled":false,"maxReplicas":100,"minReplicas":1,"targetCPUUtilizationPercentage":80}` | ---------------------------------------------------------------------------- |
| autoscaling.enabled | bool | `false` | Enable HorizontalPodAutoscaler |
| autoscaling.maxReplicas | int | `100` | Maximum replicas |
| autoscaling.minReplicas | int | `1` | Minimum replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization percentage |
| fullnameOverride | string | `""` | Override full resource name |
| httpRoute | object | `{"annotations":{},"enabled":false,"hostnames":["chart-example.local"],"parentRefs":[{"name":"gateway","sectionName":"http"}],"rules":[{"matches":[{"path":{"type":"PathPrefix","value":"/headers"}}]}]}` | ---------------------------------------------------------------------------- |
| httpRoute.annotations | object | `{}` | HTTPRoute annotations |
| httpRoute.enabled | bool | `false` | Enable HTTPRoute resource |
| httpRoute.hostnames | list | `["chart-example.local"]` | Hostnames |
| httpRoute.parentRefs | list | `[{"name":"gateway","sectionName":"http"}]` | Gateway parent references |
| httpRoute.rules | list | `[{"matches":[{"path":{"type":"PathPrefix","value":"/headers"}}]}]` | Routing rules |
| image | object | `{"pullPolicy":"Always","repository":"baabdoul/namespace-operator","tag":"1.3.0"}` | ---------------------------------------------------------------------------- |
| image.pullPolicy | string | `"Always"` | Image pull policy (Always, IfNotPresent, Never) |
| image.repository | string | `"baabdoul/namespace-operator"` | Container image repository |
| image.tag | string | `"1.3.0"` | Image tag (defaults to Chart.appVersion if empty) |
| imagePullSecrets | list | `[]` | Secrets for private registry authentication |
| ingress | object | `{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example.local","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}],"tls":[]}` | ---------------------------------------------------------------------------- |
| ingress.annotations | object | `{}` | Ingress annotations |
| ingress.className | string | `""` | Ingress class name |
| ingress.enabled | bool | `false` | Enable Ingress resource |
| ingress.tls | list | `[]` | TLS configuration |
| leaderElection | bool | `true` | Enable leader election (recommended in HA mode) |
| livenessProbe | object | `{"httpGet":{"path":"/healthz","port":"health"},"initialDelaySeconds":15,"periodSeconds":20}` | ---------------------------------------------------------------------------- |
| livenessProbe.httpGet | object | `{"path":"/healthz","port":"health"}` | Liveness probe configuration |
| manager | object | `{"health":{"bindAddress":":8081","enabled":true},"metrics":{"bindAddress":":8080","enabled":true}}` | ---------------------------------------------------------------------------- |
| manager.health.bindAddress | string | `":8081"` | Health probe bind address |
| manager.health.enabled | bool | `true` | Enable health endpoint |
| manager.metrics.bindAddress | string | `":8080"` | Metrics bind address |
| manager.metrics.enabled | bool | `true` | Enable metrics endpoint |
| nameOverride | string | `""` | Override chart name |
| nodeSelector | object | `{}` | Node selector constraints |
| podAnnotations | object | `{}` | Additional pod annotations |
| podLabels | object | `{}` | Additional pod labels |
| podSecurityContext.runAsNonRoot | bool | `true` | Enforce running as non-root user |
| ports | list | `[{"containerPort":8080,"name":"metrics","protocol":"TCP"},{"containerPort":8081,"name":"health","protocol":"TCP"}]` | Container ports exposed by the controller |
| readinessProbe.httpGet | object | `{"path":"/readyz","port":"health"}` | Readiness probe configuration |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.periodSeconds | int | `10` |  |
| replicaCount | int | `1` | Number of controller replicas |
| resources | object | `{}` | CPU/Memory resource requests & limits |
| securityContext.allowPrivilegeEscalation | bool | `false` | Prevent privilege escalation |
| securityContext.capabilities.drop | list | `["ALL"]` | Drop Linux capabilities |
| securityContext.readOnlyRootFilesystem | bool | `true` | Mount root filesystem as read-only |
| service | object | `{"port":8081,"type":"ClusterIP"}` | ---------------------------------------------------------------------------- |
| service.port | int | `8081` | Service port |
| service.type | string | `"ClusterIP"` | Service type (ClusterIP, NodePort, LoadBalancer) |
| serviceAccount | object | `{"annotations":{},"automount":true,"create":true,"name":"namespace-operator-controller-manager"}` | ---------------------------------------------------------------------------- |
| serviceAccount.annotations | object | `{}` | ServiceAccount annotations |
| serviceAccount.automount | bool | `true` | Automatically mount ServiceAccount token |
| serviceAccount.create | bool | `true` | Create a ServiceAccount |
| serviceAccount.name | string | `"namespace-operator-controller-manager"` | ServiceAccount name (auto-generated if empty and create=true) |
| tolerations | list | `[]` | Tolerations |
| volumeMounts | list | `[]` | Additional volume mounts |
| volumes | list | `[]` | Additional volumes |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
