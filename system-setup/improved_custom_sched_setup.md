# Improved custom GPU scheduler setup

## 1\. Create improved custom scheduler rbac YAML file (kube-master)

```bash
kubectl create -f PATH-TO-improved-custom-sched-setup-DIR/improved-custom-sched-setup/improved-custom-scheduler-rbac.yaml
```

## 2\. Create improved custom scheduler deployment YAML file (kube-master)

```bash
kubectl create -f PATH-TO-improved-custom-sched-setup-DIR/improved-custom-sched-setup/improved-custom-scheduler-dep.yaml
```
