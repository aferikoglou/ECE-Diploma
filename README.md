# Kubernetes GPU Scheduling

In this project we try to create a GPU Kubernetes scheduler and compare it with the Alibaba GPU sharing scheduler extender.

In __system-setup__ folder the instructions of setting up the system I have used can be found.

We have created a GPU workload monitoring system. The *script.py* submits workloads to the Kubernetes cluster. While a workload is running, the script uses our *Go program* to get GPU metrics from Prometheus TSDB. When the workload is finished a *Python script* is used to plot the output metrics. Finally, the workload is deleted from the cluster.

We have created a Kubernetes scheduler that is based on [Kube-Knots paper](docs/papers).

> The container images that are used can be found in my [Dockerhub account](https://hub.docker.com/search?q=aferikoglou&type=image).


## Getting Started

These instructions will get you a copy of the project on your local machine.

## Prerequisites


## Setup

### Workload Generator / Monitoring System


### Custom Scheduler


## Usage

### Workload Generator / Monitoring System


### Custom Scheduler


## Results

### Workload Generator / Monitoring System


### Custom Scheduler


## Author

* **Ferikoglou Aggelos**

This project was created through my *Diploma* in *[Microlab](https://microlab.ntua.gr/)*.

