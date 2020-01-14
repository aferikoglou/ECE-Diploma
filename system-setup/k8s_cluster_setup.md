# Setting up our Kubernetes cluster
We use [Kubespray](https://github.com/kubernetes-sigs/kubespray) to install and make the initial setup of our Kubernetes cluster.
## Step 1: Clone the kubespray repository (kube-master)
```
mkdir projectName
cd projectName
git clone https://github.com/kubernetes-sigs/kubespray.git
```
Kubespray offers a variety of choices according to the Cluster configuration:
  - Number of nodes
  - Name of the nodes
  - Nodes classification
  - Different network plugins
  - Kubernetes release selection
## Step 2 (kube-master)
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
## Step 3 (kube-master, kube-cpu, kube-gpu)
Generate and exchange ssh keys
```
ssh-keygen
```
Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-cpu and kube-gpu for kube-master.
Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-master for kube-cpu.
Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-master for kube-gpu.

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
## Step 4 (kube-master)
Install dependencies from ``requirements.txt``
```
sudo pip install -r requirements.txt
```
Run as root in kubespray folder:
```
ansible-playbook  --private-key=/path/to/private/key --user=ubuntu -i inventory/local/hosts.ini --become --become-user=root cluster.yml
```
> MY COMMAND
ansible-playbook  --private-key=/root/.ssh/id_rsa --user=root -i /root/aferik_diploma/kubespray/inventory/local/hosts_changed.ini --become --become-user=root cluster.yml
## Step 5 (kube-master)
```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
## Step 6: Check if the Kubernetes cluster is OK (kube-master)
```
kubectl get pods -o wide -n kube-system
```
> NOTE 
If error "too many nameservers" occur just comment out some of the nameservers in /etc/resolv.conf file (maximum 3 nameservers).




