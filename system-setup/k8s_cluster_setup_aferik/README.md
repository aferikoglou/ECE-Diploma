## Setting up our Kubernetes cluster

We use [Kubespray](https://github.com/kubernetes-sigs/kubespray) to install and make the initial setup of our Kubernetes cluster.

## Step 1: Clone the kubespray repository (kube-master-1)

$ mkdir projectName
$ cd projectName
$ git clone https://github.com/kubernetes-sigs/kubespray.git

Kubespray offers a variety of choices according to the Cluster configuration:
  - Number of nodes
  - Name of the nodes
  - Nodes classification
  - Different network plugins
  - Kubernetes release selection

## Step 2: (kube-master-1)

$ cd kubespray

- Configure the /inventory/local/hosts.ini file using your own IPs and name for each node.
  NOTE : Deletion of the first line ( kube-master-1 ansible_connection=local local_release_dir={{ansible_env.HOME}}/releases ) requires ssh connection from kube-master-1 to kube-master-1

- Configure the /inventory/sample/group_vars/k8s_cluster/k8s_cluster.yml:
  'kube_network_plugin: flannel'

## Step 3: (kube-master-1, kube-cpu-1, kube-gpu-1)

$ ssh-keygen
- Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-cpu-1 and kube-gpu-1 for kube-master-1.
  Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-master-1 for kube-cpu-1.
  Add id_rsa.pub (public key) to ~/.ssh/authorized_keys of kube-master-1 for kube-gpu-1.
- Disable firewall ($ sudo ufw disable)
- Enable ip forwarding ($ sudo sysctl -w net.ipv4.ip_forward=1)
- Disable swap ($ swapoff -a)

## Step 4: (kube-master-1)

- Install dependencies from ``requirements.txt``.
$ sudo pip install -r requirements.txt

- Run as root in kubespray folder:
$ ansible-playbook  --private-key=/path/to/private/key --user=ubuntu -i inventory/local/hosts.ini --become --become-user=root cluster.yml

(MY COMMAND : ansible-playbook  --private-key=/root/.ssh/id_rsa --user=root -i /root/aferik_diploma/kubespray/inventory/local/hosts_changed.ini --become --become-user=root cluster.yml)

## Step 5: (kube-master-1)

$ mkdir -p $HOME/.kube OK
$ sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
$ sudo chown $(id -u):$(id -g) $HOME/.kube/config

## Step 6: Check if the Kubernetes cluster is OK (kube-master-1)

$ kubectl get pods -o wide -n kube-system


## NOTE

If error "too many nameservers" occur just comment out some of the nameservers in /etc/resolv.conf file (maximum 3 nameservers).



