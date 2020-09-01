# Resource Aware GPU Scheduling in Kubernetes Infrastructure

In this project we tried to create a custom Kubernetes GPU scheduler based on [Kube-Knots](docs/papers) paper and compare it with the [official NVIDIA GPU device plugin](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/) and the [Alibaba GPU sharing scheduler extender](https://www.alibabacloud.com/blog/gpu-sharing-scheduler-extender-now-supports-fine-grained-kubernetes-clusters_594926).

Our Kubernetes cluster consists of three nodes kube-master, kube-cpu and kube-gpu. The kube-master node is the admin node of the cluster while kube-cpu and kube-gpu are the worker nodes. Their main difference is that kube-gpu node has GPU access while kube-cpu has not. As a GPU we used a NVIDIA TESLA V100 card. 

The proposed experimental infrastructure contains a GPU monitoring system. Our monitoring system consists of DCGM node exporter (exports GPU metrics in timeseries format) and [Prometheus TSDB](https://prometheus.io/) (stores GPU metrics and provides PromQL for quering the data).

The used workloads where created using image classification and object detection tasks from the [MLPerf Inference benchmark](https://mlperf.org/inference-overview/) suite.

Our proposed system creates a workload and feeds it to all our available schedulers. Each workload is fed multiple times to our For each workload execution
a set of GPU metrics are exported and plotted. The execution statistics data for each Job are also exported. These data are used for the comparison of our schedulers.

> The container images that are used can be found in this [Dockerhub account](https://hub.docker.com/search?q=aferikoglou&type=image).

More information for this project can be found in the detailed report in __docs__ folder. 

## Getting Started

These instructions will get you a copy of the project on your cluster.

## Prerequisites

* Ubuntu
* Python
* Go

## Setup

### Experimental Infrastructure

In __system-setup__ folder the instructions of setting up the experimental infrastructure can be found.

## Usage

Execute a full experiment.

```bash
./runFinalExperiments.sh
```

## Results

Sample output results can be found in monitoring files. The comparative results can be found in the detailed report in __docs__ folder.

## Author

* **Ferikoglou Aggelos**

This project was created through my *Diploma* in *[Microlab](https://microlab.ntua.gr/)*.

