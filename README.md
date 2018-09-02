# mkaas

Minikube as a Service (mkaas)

mkaas provides a declarative way to create Kubernetes clusters using minikube within 1-2 minutes each.

* Create your named Minikube YAML file (custom resource) using details below
* Let mkaas do its thing
* Download the `.kube/config` from the designated node
* Set an environmental proxy setting to the node to reach the cluster on its private subnet
* Profit with `kubectl`, NodePorts, etc
* Tear down with `kubectl delete minikube/cluster_name` when you're done or add additional clusters

Status: this is a Proof-of-Concept Kubernetes Operator providing Minikube-as-a-Service or `mkaas`.

## PoC demo

[![asciicast](https://asciinema.org/a/O0pWQ6p7slnQ2X8Q5qtIjHqZG.png)](https://asciinema.org/a/O0pWQ6p7slnQ2X8Q5qtIjHqZG)

## How does it work?

This combines a custom resource for defining minikube clusters and an Operator for Kubernetes
to schedule Pods that create your minikube clusters.

Example cluster:

```yaml
apiVersion: "alexellis.io/v1alpha1"
kind: "Minikube"
metadata:
  name: "alex"
  namespace: clusters
spec:
  clusterName: "alex"
  cpuCount: 2
  memoryMB: 2048
```

* What do I need?

On your host you'll need virtualization support / KVM support.

* KVM package
* libvirtd
* Kubernetes (a single-node tainted master with kubeadm is fine)

Add the KVM packages for your distro (tested with Ubuntu)

Follow [these steps](https://gist.github.com/alexellis/eec21a96906726d08a071d58aee66ab9#create-a-cluster-with-kubeadm) on Ubuntu 16.04 up until you get to "Create a cluster with kubeadm".

You could use `kubeadm` for this. For Cloud turn on nested-virt with GCP or use Packet.net/Scaleway for a bare metal host.

> Note: if you use a public host, then I recommend you setup a firewall rule with `ufw` to block access to port 3128 on the host. You can still make use of the proxy using an SSH tunnel. The bundled squid proxy is set up to allow open access. `ssh -L 3128:3128 -N remote-host` then replace the HTTP proxy with `127.0.0.1:3128`.

* How does it work?

It uses a privileged Pod found in ./agent/. The container inside the Pod has
privileged access to the host and host networking which is required for the use
of minikube. The VMs are created using `minikube start`.

VMs are stored in `/var/mkaas/.minikube/` and this folder is mounted by the controller.

* How are machines deleted?

If you delete the custom resource i.e. `kubectl -n clusters delete minikube/alex` then the Pod will be reclaimed. It has a script listening for sigterm / sigint and will call `minikube destroy`.

* Are restarts supported.

Yes

* Is this production-ready?

Due to the privileges required to execute minikube commands this should not be run in a production environment or on clusters containing confidential data. In the future this may be able to be restricted to just a `libvirtd` socket.

The proxy container runs on the host network which means using this proxy you can reach any hosts reachable from the host node. In the future some limitations on the subnet could be applied - i.e. to only allow outgoing via the minikube subnet.

* Are multiple hosts supported?

Yes and if you use an NFS mount it may even allow for "motion" between hosts.

* Can nested virt be used?

Untested, but [perhaps](https://docs.fedoraproject.org/en-US/quick-docs/using-nested-virtualization-in-kvm/)?

* Has anyone had this idea before?

The [kube-virt](https://github.com/kubevirt/kubevirt) project has an open issue suggesting [this is in the works for them](https://github.com/kubevirt/kubevirt/issues/736).

* What else could this do?

It could be used to create `docker-machine` VMs instead of `minikube` clusters for temporary environments to perform isolated Docker builds.

## Usage:

* Install Kubernetes with `kubeadm`

Use one of my guides:

[Your instant Kubernetes cluster](https://blog.alexellis.io/your-instant-kubernetes-cluster/)

After running `kubeadm init` don't forget to deploy a network driver (included in guide) and to taint the master if you're using a single node.

```bash
kubectl taint nodes --all node-role.kubernetes.io/master-
```

* Install KVM and libvirtd

```bash
sudo apt update \
  && sudo apt-get install -qy \
    git qemu-kvm libvirt-bin virtinst bridge-utils cpu-checker --no-install-recommends \
  && sudo kvm-ok \
  && sudo usermod root -aG libvirtd
```

* Install `ufw` firewall

If using public cloud you can install the `ufw` firewall to block access to the squid proxy which will be deployed on each node on port `3128`.

```bash
sudo apt install ufw -qy

sudo -i
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw enable
ufw status verbose
```

This has now blocked any access incoming to your server apart from when using ssh on TCP port 22.

* (Optional) Install [operator-sdk](https://github.com/operator-framework/operator-sdk)

* Make a global settings folder:

```bash
sudo mkdir -p /var/mkaas
```

Create the following directories, or create a cluster using `minikube start` and delete it after which will do the same thing.

```bash
sudo mkdir -p /var/mkaas/.minikube
sudo mkdir -p /var/mkaas/.kube
```

* Clone this repo into the $GOPATH

```bash
mkdir -p /go/src/github.com/operator-framework/operator-sdk/
cd /go/src/github.com/operator-framework/operator-sdk/
git clone https://github.com/operator-framework/operator-sdk
cd operator-sdk
git clone https://github.com/alexellis/mkaas minikube
cd minikube
```

* (Optional) Build/push (optional to rebuild)

```bash
operator-sdk build alexellis2/mko:v0.0.6 \
  && docker push alexellis2/mko:v0.0.6
```

* Deploy on a host:

Setup the Operator, RBAC and CRD:

```bash
kubectl create ns clusters
cd deploy
kubectl apply -f crd.yaml,operator.yaml,rbac.yaml
```

Now create the first cluster:

```bash
kubectl apply -f cr.yaml
```

This will create your first cluster and place a helper Pod into the `clusters` namespace.

Check the logs:

```bash
kubectl logs -n clusters pod/alex-minikube -f
Starting local Kubernetes v1.10.0 cluster...
Starting VM...
Getting VM IP address...
Moving files into cluster...
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
```

Wait until you see the bundle created.

Now you can access the cluster from the host using `kubectl` by retrieving the IP of the
cluster and the IP of the node.

Squid will now be run on the host node as part of a daemonset exposing port `3128`. It requires host networking to be able to reach the minikube network.

Now:

Get your Minikube IP either when we copy the .kube/config file down later on, or on the host with this command:

```bash
sudo -i MINIKUBE_HOME=/var/mkaas/.minikube/ minikube ip --profile alex
192.168.39.125
```

Note down your minikube IP, for example: `192.168.39.125`.

Run these commands on your remote client/laptop:

For HTTP access on local network:

```bash
export http_proxy=http://node_ip:3128
faas-cli list --gateway $MINIKUBE_IP
```

If using a remote host with `ufw` enabled then open an SSH tunnel to give you access to the remote squid proxy instance:

```
ssh -L 3128:3128 -N node_ip
```

With the SSH tunnel the config for the `http_proxy` environmental variable would become as follows:

```bash
export http_proxy=http://127.0.0.1:3128
faas-cli list --gateway $MINIKUBE_IP
```

For access via `kubectl`:

Copy the bundle to your client/laptop and untar using (sftp/scp):

```bash
mkdir -p mkaas
cd mkaas

scp node:/var/mkaas/alex-bundle.tgz .
tar -xvf alex-bundle.tgz
```

If your home directory is `/home/alex/` then do the following:

```bash
sed -ie 's#/root/#/home/alex/mkaas/#g' .kube/config
```

This changes the absolute paths used for the root user to match the point you copied to.

Now:

```bash
export http_proxy=http://node_ip:3128
export KUBECONFIG=.kube/config

kubectl get node
NAME       STATUS    ROLES     AGE       VERSION
minikube   Ready     master    1m        v1.10.0
```

* Deploy a test workload and access over the proxy

Add the CLI if not present:

```bash
curl -sLSf https://cli.openfaas.com | sudo sh
```

Deploy OpenFaaS:

```bash
git clone https://github.com/openfaas/faas-netes
kubectl apply -f ./faas-netes/namespaces.yml,./faas-netes/yaml
rm -rf faas-netes

export minikube_ip=

curl $minikube_ip:31112/system/info; echo
{"provider":{"provider":"faas-netes","version":{"sha":"5539cf43c15a28e9af998cdc25b5da06252b62e1","release":"0.6.0"},"orchestration":"kubernetes"},"version":{"commit_message":"Attach X-Call-Id to asynchronous calls","sha":"c86de503c7a20a46645239b9b081e029b15bf69b","release":"0.8.11"}}

export OPENFAAS_URL=$minikube_ip:31112

faas-cli store deploy figlet

echo "Sleeping for 5 seconds while figlet is downloaded"
sleep 5
echo "MKAAS!" | faas-cli invoke figlet
```

## Development / troubleshooting

Run these commands on your host node.

Operator logs:

```bash
kubectl logs -n clusters deploy/minikube -f
```

Events:

```bash
kubectl get events --sort-by='{.firstTimestamp}' -n clusters
```

Resources:

```bash
kubectl get all -n clusters
```

## License

MIT License

Copyright Alex Ellis 2018
