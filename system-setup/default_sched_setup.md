# Kubernetes cluster - GPU Support

We follow the instructions shown in [NVIDIA/k8s-device-plugin](https://github.com/NVIDIA/k8s-device-plugin#preparing-your-gpu-nodes) github repository.

We assume that the NVIDIA driver is installed in the GPU node.

## 1\. Install [nvidia-docker2](https://github.com/NVIDIA/nvidia-docker) (kube-gpu)

Add the package repositories

```
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list
sudo apt-get update && sudo apt-get install -y nvidia-docker2
```

Restart docker

```
sudo systemctl restart docker
```

## 2\. Install [nvidia-container-runtime](https://github.com/NVIDIA/nvidia-container-runtime) (kube-gpu)

```
sudo apt-get install nvidia-container-runtime
```

## 3\. Enable nvidia runtime as default runtime on node (kube-gpu)

Edit the docker daemon config file

```
vim /etc/docker/daemon.json
---
{
    "default-runtime": "nvidia",
    "runtimes": {
        "nvidia": {
            "path": "/usr/bin/nvidia-container-runtime",
            "runtimeArgs": []
        }
    }
}
---
```

Restart kube-gpu kubelet

```
systemctl restart kubelet
```

## 4\. Enable GPU support in Kubernetes (kube-master)

Deploy the following daemonset

```
kubectl create -f .../default-sched-setup/nvidia-device-plugin.yml
```

## 5\. Check if GPU support is enabled in Kubernetes cluster (kube-master)

Check if GPU appears in the allocated resources section

```
kubectl describe nodes kube-gpu
```
Run GPU pod

```
vim gpu-pod.yaml

---
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
spec:
  containers:
    - name: digits-container
      image: nvidia/digits:6.0
      resources:
        limits:
          nvidia.com/gpu: 1
---

kubectl apply -f gpu-pod.yaml

kubectl get pods
```

In kube-gpu execute

```
nvidia-smi
```
