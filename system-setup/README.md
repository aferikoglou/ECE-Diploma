# System Setup

The system I used consists of a *Kubernetes cluster* with three nodes (__k8s_cluster_setup.md__).

- *kube-master*: admin node
- *kube-cpu*: node without GPU access
- *kube-gpu*: node with GPU access (*NVIDIA TESLA V100*)

I enabled *Kubernetes GPU support* in order to use the GPU of *kube-gpu* node as an extendend resource (__k8s_gpu_setup.md__).

I also created a *Monitoring Mechanism* (*GPU Node Exporter* and *Prometheus TSDB*) in order to get GPU metrics and use them in my scheduler (__k8s_gpu_monitoring_setup.md__).

Finally, the results of my implementation are going to be compared with the *[Alibaba GPUshare scheduler extender](https://github.com/AliyunContainerService/gpushare-scheduler-extender)* (__k8s_gpushare_scheduler_extender_setup.md__).

The following image describes the system so far.

<img src="images/cluster-image.png" width="400" height="400">


