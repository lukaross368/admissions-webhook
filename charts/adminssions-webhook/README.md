# Helm Chart Installation

This chart deploys the admission webhook server, its Kubernetes webhook configurations, a Prometheus `ServiceMonitor`, and a Grafana dashboard.

## Prerequisites

- Kubernetes cluster with webhook admission controllers enabled
- [Helm](https://helm.sh/docs/intro/install/) v3+
- [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) (or equivalent) installed in the cluster for metrics scraping and dashboards

## 1. Install the Prometheus stack

If you do not already have Prometheus and Grafana running in your cluster, install the community stack:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack
```

## 2. Generate TLS certificates

The webhook server requires a TLS certificate. The chart includes a helper script to generate a self-signed certificate and extract the CA bundle needed by the webhook configuration.

From the repository root:

```bash
bash charts/adminssions-webhook/scripts/certs.sh
```

The script outputs the base64-encoded certificate and key. Copy these values into `values.yaml` (or pass them via `--set`):

```yaml
tls:
  crt: <base64-encoded certificate>
  key: <base64-encoded private key>
```

> **Note**: Self-signed certificates are suitable for development only. For production, use a certificate authority such as [cert-manager](https://cert-manager.io/) to issue and rotate certificates automatically.

## 3. Install the chart

```bash
helm install admissions-webhook ./charts/adminssions-webhook \
  --set tls.crt=<base64-cert> \
  --set tls.key=<base64-key>
```

Or with a values file:

```bash
helm install admissions-webhook ./charts/adminssions-webhook -f my-values.yaml
```

## 4. Verify the deployment

```bash
kubectl get pods -l app=admissions-webhook
kubectl get validatingwebhookconfiguration
kubectl get mutatingwebhookconfiguration
```

## 5. View metrics in Grafana

Once the `ServiceMonitor` is picked up by Prometheus, the provisioned Grafana dashboard will show request counts and latency histograms for both webhook endpoints.

To access Grafana locally:

```bash
kubectl port-forward svc/prometheus-grafana 3000:80
```

Then open [http://localhost:3000](http://localhost:3000) and search for the **admissions-webhook** dashboard.

## Configuration reference

| Key | Default | Description |
|-----|---------|-------------|
| `tls.crt` | `""` | Base64-encoded TLS certificate |
| `tls.key` | `""` | Base64-encoded TLS private key |
| `webhookServer.replicas` | `1` | Number of webhook server replicas |
| `webhookServer.port` | `8443` | HTTPS port for webhook traffic |
| `webhookServer.metricsPort` | `2112` | Port exposing Prometheus metrics |
| `webhookServer.image.repository` | `lukaross123/admissions-webhook` | Container image repository |
| `webhookServer.image.tag` | `v0.4.0` | Container image tag |
| `webhookServer.resources.cpu` | `50m` | CPU request/limit |
| `webhookServer.resources.memory` | `64Mi` | Memory request/limit |
| `prometheus.enabled` | `true` | Deploy a `ServiceMonitor` for Prometheus scraping |
| `prometheus.releaseName` | `prometheus` | Release name of your Prometheus installation (used to match `ServiceMonitor` labels) |
| `grafana.enabled` | `true` | Provision a Grafana dashboard ConfigMap |

## Uninstall

```bash
helm uninstall admissions-webhook
```
