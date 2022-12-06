# opssc-api-server

![Version: 0.3.0](https://img.shields.io/badge/Version-0.3.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

OpsSC API Server for Hyperledger Fabric v2.x

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| adminMSPID | string | `"Org1MSP"` | MSP ID for the organization to be operated |
| affinity | object | `{}` | Affinity settings for pod assignment |
| api.channelProposalEnabled | bool | `true` | Whether to enable the Channel Proposal APIs |
| api.utilEnabled | bool | `true` | Whether to enable the Utility APIs |
| autoscaling.enabled | bool | `false` | Currently autoscaling is unsupported |
| discoverAsLocalhost | bool | `false` | Whether to discover as localhost |
| fullnameOverride | string | `""` | Override the full name of resources |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"fabric-opssc/opssc-api-server"` | opssc-api-server image repository |
| image.tag | string | `"latest"` | opssc-api-server image tag (Overrides the image tag whose default is the chart appVersion) |
| imagePullSecrets | list | `[]` | Image pull secrets |
| ingress.annotations | object | `{}` | Ingress annotations |
| ingress.className | string | `""` | Ingress class name |
| ingress.enabled | bool | `false` | If true, Ingress will be created |
| ingress.hosts | list | `[{"host":"opssc-api-server.local","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]` | Ingress hostnames |
| ingress.hosts[0] | object | `{"host":"opssc-api-server.local","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}` | Ingress host name |
| ingress.hosts[0].paths[0] | object | `{"path":"/","pathType":"ImplementationSpecific"}` | Ingress path |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` | Ingress path type |
| ingress.tls | list | `[]` | Ingress TLS configuration |
| livenessProbe.enabled | bool | `true` | Whether to enable livenessProbe |
| livenessProbe.failureThreshold | int | `3` | The livenessProbe failure threshold |
| livenessProbe.initialDelaySeconds | int | `5` | The livenessProbe initial delay (in seconds) |
| livenessProbe.periodSeconds | int | `30` | The livenessProbe period (in seconds) |
| livenessProbe.successThreshold | int | `1` | The livenessProbe success threshold |
| livenessProbe.timeoutSeconds | int | `5` | The livenessProbe timeout (in seconds) |
| logLevel | string | `"info"` | Log level |
| nameOverride | string | `""` | Override the name of resources |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| opsscChaincodeOpsCCName | string | `"chaincode-ops"` | Chaincode name of the chaincode OpsSC |
| opsscChannelName | string | `"ops-channel"` | Channel name for the OpsSC |
| opsscChannelOpsCCName | string | `"channel-ops"` | Chaincode name of the channel OpsSC |
| podAnnotations | object | `{}` | Pod annotations |
| podSecurityContext | object | `{}` | Pod security context |
| readinessProbe.enabled | bool | `true` | Whether to enable readinessProbe |
| readinessProbe.failureThreshold | int | `3` | The readinessProbe failure threshold |
| readinessProbe.initialDelaySeconds | int | `5` | The readinessProbe initial delay (in seconds) |
| readinessProbe.periodSeconds | int | `30` | The readinessProbe period (in seconds) |
| readinessProbe.successThreshold | int | `1` | The readinessProbe success threshold |
| readinessProbe.timeoutSeconds | int | `5` | The readinessProbe timeout (in seconds) |
| replicaCount | int | `1` | Replica count, currently only `1` is assumed |
| resources | object | `{}` | CPU/Memory resource requests/limits |
| secrets.adminMSPName | string | `"org1-admin-msp"` | Admin MSP config name, the value should be a tar file named `admin-msp.tar` (TODO: Should be improved) |
| secrets.connectionProfileItemKeyName | string | `"HLF_CONNECTION_PROFILE_ORG1"` | Connection profile item key name, the value should be a JSON or YAML-based connection profile file |
| secrets.connectionProfileName | string | `"fabric-rest-sample-config"` | Connection profile config name (TODO: Should be improved) |
| securityContext | object | `{}` | Security context |
| service.port | int | `5000` | TCP port |
| service.type | string | `"ClusterIP"` | k8s service type exposing ports (e.g., ClusterIP) |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` | Toleration labels for pod assignment |
| websocket.enabled | bool | `true` | Whether to enable WebSocket server to receive messages from agents or the API server itself |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
