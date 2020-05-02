# Monitoring system setup

- [Node Exporter installation guide](https://github.com/NVIDIA/gpu-monitoring-tools/tree/master/exporters/prometheus-dcgm) 
- [Prometheus installation guide](https://devopscube.com/setup-prometheus-monitoring-on-kubernetes/)

## 1\. Node Exporter setup

### 1.1\. Label GPU node

Add label to GPU node.

```bash
kubectl label nodes GPU-NODE-NAME hardware-type=NVIDIAGPU
```

Check if the label was added.

```bash
kubectl get nodes --show-labels
```

### 1.2\. Deploy node exporter daemonset

```bash
kubectl create -f PATH-TO-node-exporter-setup-DIR/node-exporter-setup/gpu-node-exporter-daemonset.yaml
```

### 1.3\. Testing

```bash
curl 192.168.1.147:9100/metrics
```

## 2\. Prometheus TSDB setup

### 2.1\. Create monitoring namespace

```bash
kubectl create namespace monitoring
kubectl create -f PATH-TO-prometheus-setup-DIR/prometheus-setup/clusterRole.yaml
```

### 2.2\. Create config map

```bash
kubectl create -f PATH-TO-prometheus-setup-DIR/prometheus-setup/config-map.yaml
```

### 2.3\. Create Prometheus deployment

```bash
kubectl create -f PATH-TO-prometheus-setup-DIR/prometheus-setup/prometheus-deployment.yaml
```

### 2.4\. Exposing Prometheus as a service

```bash
kubectl create -f PATH-TO-prometheus-setup-DIR/prometheus-setup/prometheus-service.yaml --namespace=monitoring
```

### 2.5\. Testing

```bash
curl '192.168.1.145:30000/api/v1/query?query=dcgm_gpu_temp' | python -m json.tool
```

