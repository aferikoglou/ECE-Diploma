## Setting up our Kubernetes cluster

We use [Kubespray](https://github.com/kubernetes-sigs/kubespray) to install and make the initial setup of our Kubernetes cluster.

#### Step 1 : Clone the kubespray repository  

```sh
$ mkdir projectName
$ cd projectName
$ git clone https://github.com/kubernetes-sigs/kubespray.git OK
```
Kubespray offers a variety of choices according to the Cluster configuration:
  - Number of nodes
  - Name of the nodes
  - Nodes classification
  - Different network plugins
  - Kubernetes release selection

#### Step 2: 

- Configure the inventory/sample/hosts.ini file using your own IPs and name for each node. --- inventory/sample/inventory.ini OK
- Configure the /inventory/sample/group_vars/k8s_cluster/k8s_cluster.yml:
    ```sh
    kube_network_plugin: flannel
    ```
  OK

### Step 3:
- Exchange RSA keys both private and public between VMs so as to communicate each other --- OK for kube-master and kube-cpu needs to be done for kube-master and kube-gpu
- Diasble firewall (sudo ufw disable) OK
- Enable ip forwarding (sudo sysctl -w net.ipv4.ip_forward=1) OK
- Disable swap (swapoff -a) OK

### Step 4:
Install dependencies from ``requirements.txt``
sudo pip install -r requirements.txt OK

Run as root
```sh
$ ansible-playbook  --private-key=/path/to/private/key --user=ubuntu -i inventory/mycluster/hosts.ini --become --become-user=root cluster.yml
```
ansible-playbook  --private-key=/root/.ssh/id_rsa --user=root -i /root/aferik_diploma/kubespray/inventory/inventory.ini --become --become-user=root cluster.yml ?

ansible-playbook --flush-cache -u root -i /root/aferik_diploma/kubespray/inventory/inventory.cfg -b -v cluster.yml

ansible-playbook -i /root/aferik_diploma/kubespray/inventory/inventory.ini cluster.yml -b -v --private-key=/root/.ssh/id_rsa

### Step 5
In master node
 ```sh
 $ mkdir -p $HOME/.kube OK
 $ sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
 $ sudo chown $(id -u):$(id -g) $HOME/.kube/config
```


