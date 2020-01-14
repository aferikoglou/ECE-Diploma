# Installation guide
[Node Exporter Setup](https://github.com/NVIDIA/gpu-monitoring-tools/tree/master/exporters/prometheus-dcgm) 
[Prometheus Setup](https://devopscube.com/setup-prometheus-monitoring-on-kubernetes/)
## 1\. GPU Node Exporter Setup
### 1.1\. Clone repository for gpu node exporter installation
```bash
	git clone https://github.com/NVIDIA/gpu-monitoring-tools.git
```
### 1.2\. Label GPU node
Add label to GPU node
```bash
	cd ./gpu-monitoring-tools/exporters/prometheus-dcgm/k8s/node-exporter/
	kubectl label nodes kube-gpu-1 hardware-type=NVIDIAGPU
```
Check if the label was added
```bash
	kubectl get nodes --show-labels
```
### 1.3\. Deploy node exporter daemonset
```bash
	kubectl create -f node-exporter/gpu-node-exporter-daemonset.yaml
```
Check if everything works
```bash
	curl 192.168.1.147:9100/metrics
```
## 2\. Prometheus Setup
### 2.1\. Clone repository for prometheus installation
```bash
	git clone https://github.com/bibinwilson/kubernetes-prometheus
```
### 2.2\. Create Monitoring Namespace
```bash
	kubectl create namespace monitoring
	cd ./kubernetes-prometheus
	kubectl create -f clusterRole.yaml
```
### 2.3\. Create Config Map
Change config-map.yaml
Change scrape_interval to 1s in prometheus.yml (data will be taken every 1 second)
```yaml
  global:
    scrape_interval: 1s
  ...
```
Add a job to scrape_configs section in order to get the data provided by node exporter
```yaml
  - job_name: 'node-exporter'
        static_configs:
        - targets: ['192.168.1.147:9100']
```
Create config-map.yaml
```bash
	kubectl create -f config-map.yaml
```
### 2.4\. Create Prometheus Deployment
```bash
	kubectl create  -f prometheus-deployment.yaml
```
Check the created deployment using the following command
```bash
	kubectl get deployments --namespace=monitoring
```
### 2.5\. Exposing Prometheus as a Service
```bash
	kubectl create -f prometheus-service.yaml --namespace=monitoring
```
### 2.6\. Testing
```bash
	curl '192.168.1.145:30000/api/v1/query?query=dcgm_gpu_temp' | python -m json.tool
	curl '192.168.1.145:30000/api/v1/query_range?query=dcgm_gpu_temp&start=Start_Timestamp&end=End_Timestamp&step=1s' | python -m json.tool
```
