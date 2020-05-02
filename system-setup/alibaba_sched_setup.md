# Alibaba scheduler extension setup

- [Alibaba scheduler extension installation guide](https://github.com/AliyunContainerService/gpushare-scheduler-extender/blob/master/docs/install.md?spm=a2c65.11461447.0.0.5d8b51batZVMzf&file=install.md)

## 1\. Deploy GPU share scheduler extender

```bash
cp PATH-TO-alibaba-sched-setup-DIR/alibaba-sched-setup/scheduler-policy-config.json /etc/kubernetes/
kubectl create -f PATH-TO-alibaba-sched-setup-DIR/alibaba-sched-setup/gpushare-schd-extender.yaml
```

## 2\. Modify scheduler configuration

The goal is to include `/etc/kubernetes/scheduler-policy-config.json` into the scheduler configuration (`/etc/kubernetes/manifests/kube-scheduler.yaml`).

> Notice: If your Kubernetes default scheduler is deployed as static pod, don't edit the yaml file inside /etc/kubernetes/manifest. You need to edit the yaml file outside the `/etc/kubernetes/manifest` directory. and copy the yaml file you edited to the '/etc/kubernetes/manifest/' directory, and then kubernetes will update the default static pod with the yaml file automatically.

```bash
cp PATH-TO-alibaba-sched-setup-DIR/alibaba-sched-setup/modified-kube-scheduler.yaml /etc/kubernetes/manifests/kube-scheduler.yaml
```

## 3\. Deploy Device Plugin

```bash
kubectl create -f PATH-TO-alibaba-sched-setup-DIR/alibaba-sched-setup/device-plugin-rbac.yaml

kubectl create -f PATH-TO-alibaba-sched-setup-DIR/alibaba-sched-setup/device-plugin-ds.yaml
```

> Notice: Remove default GPU device plugin, for example, if you are using [nvidia-device-plugin](https://github.com/NVIDIA/k8s-device-plugin/blob/v1.11/nvidia-device-plugin.yml), you can run `kubectl delete ds -n kube-system nvidia-device-plugin-daemonset` to delete.

## 4\. Add gpushare node labels to the nodes requiring GPU sharing

You need to add a label "gpushare=true" to all node where you want to install device plugin because the device plugin is deamonset.

```bash
kubectl label node GPU-NODE-NAME gpushare=true
```

## 5\. Install kubectl extension

### 5.1 Download and install the kubectl extension

```bash
cp PATH-TO-alibaba-sched-setup-DIR/alibaba-sched-setup/kubectl-inspect-gpushare /usr/bin/
chmod u+x /usr/bin/kubectl-inspect-gpushare
```

