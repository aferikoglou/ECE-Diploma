# Setting up our Kubernetes cluster
We use [Kubespray](https://github.com/kubernetes-sigs/kubespray) to install and make the initial setup of our Kubernetes cluster.
## 1\. Clone the kubespray repository (kube-master)
```
mkdir projectName
cd projectName
git clone https://github.com/kubernetes-sigs/kubespray.git
```
## 2\. (kube-master)
```
cd kubespray
```
Configure the /inventory/local/hosts.ini file using your own IPs and name for each node. My ``host.ini`` file:
  ```
    [all]
    kube-master-1 ansible_host=192.168.1.145 
    kube-cpu-1 ansible_host=192.168.1.146
    kube-gpu-1 ansible_host=192.168.1.147
    
    [kube-master]
    kube-master-1
    
    [etcd]
    kube-master-1
    
    [kube-node]
    kube-cpu-1
    kube-gpu-1
    
    [k8s-cluster:children]
    kube-master
    kube-node
  ```
Configure the /inventory/sample/group_vars/k8s_cluster/k8s_cluster.yml:
*kube_network_plugin: flannel*
## 3\. (kube-master, kube-cpu, kube-gpu)
Generate and exchange ssh keys
```
ssh-keygen
```
- Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-cpu and kube-gpu for kube-master.
- Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-master for kube-cpu.
- Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-master for kube-gpu.

Disable firewall 
```
sudo ufw disable
```
Enable ip forwarding
```
sudo sysctl -w net.ipv4.ip_forward=1
```
Disable swap
```
swapoff -a
```
## 4\. (kube-master)
Install dependencies from ``requirements.txt``
```
sudo pip install -r requirements.txt
```
Run as root in kubespray folder:
```
ansible-playbook  --private-key=/path/to/private/key --user=ubuntu -i inventory/local/hosts.ini --become --become-user=root cluster.yml
```
## 5\. (kube-master)
```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
## 6\. Check if the Kubernetes cluster is OK (kube-master)
```
kubectl get pods -o wide -n kube-system
```


