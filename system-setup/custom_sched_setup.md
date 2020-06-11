# Custom GPU scheduler setup

## 1\. Create custom scheduler rbac YAML file (kube-master)

```bash
kubectl create -f PATH-TO-custom-sched-setup-DIR/custom-sched-setup/custom-scheduler-rbac.yaml
```

## 2\. Create custom scheduler deployment YAML file (kube-master)

```bash
kubectl create -f PATH-TO-custom-sched-setup-DIR/custom-sched-setup/custom-scheduler-dep.yaml
```
