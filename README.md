# mkaas
Minikube as a Service

## PoC demo

[![asciicast](https://asciinema.org/a/s1UWfywtfpOp9be2r7igbbnBB.png)](https://asciinema.org/a/s1UWfywtfpOp9be2r7igbbnBB)

## How does it work?

This combines a custom resource for defining minikube clusters and an Operator for Kubernetes
to schedule Pods that create your minikube clusters.


Example cluster:

```yaml
apiVersion: "alexellis.io/v1alpha1"
kind: "Minikube"
metadata:
  name: "alex"
spec:
  clusterName: "alex"
  cpuCount: 2
  memoryMB: 2048
```

* What do I need?

You need KVM and libvirtd installed on your host machine and Kubernetes installed too.

You could use `kubeadm` for this. For Cloud turn on nested-virt with GCP or use Packet.net/Scaleway for a bare metal host.

* How does it work?

It uses a privileged Pod found in ./agent/. The container inside the Pod has
privileged access to the host and host networking which is required for the use
of minikube. The VMs are created using `minikube start`.

VMs are stored in /root/.minikube and this folder is mounted by the controller.

* Are restarts supported.

Yes

* Are multiple hosts supported?

Yes and if you use an NFS mount it may even allow for "motion" between hosts.

## Usage:

* Install [operator-sdk](https://github.com/operator-framework/operator-sdk)

* Clone this repo into the $GOPATH

```
mkdir -p /go/src/github.com/operator-framework/operator-sdk/
cd /go/src/github.com/operator-framework/operator-sdk/
git clone https://github.com/operator-framework/operator-sdk
```

* Build/push

```
operator-sdk build alexellis2/mko:v0.0.5 && docker push alexellis2/mko:v0.0.5
```

* Deploy on a host:

```
kubectl create ns clusters
cd deploy
kubectl apply -f .
```

This will create your first cluster and place a helper Pod into the `clusters` namespace.

Check the logs:

```
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

Now you can access the cluster from the host using `kubectl` by retrieving the IP of the
cluster and the IP of the node.

Run squid on the host with host-networking (in the future this will be automated)

This will be automated later, but for now:

```
docker run -d --name proxy --net=host alexellis2/squid-proxy:0.1
```

Now:

On the host:

```
sudo -i
minikube ip --profile alex
```

On your client:

For HTTP access:

```
export http_proxy=http://node_ip:3128
faas-cli list --gateway $MINIKUBE_IP
```

For access via `kubectl`:

Use SFTP/SCP to copy the certificates and the kubeconfig from `/root/.minikube` and place them in a new folder:

* `*.crt`
* `*.key`

Now copy `config` from `/root/.kube`.

Now:

```
export http_proxy=http://node_ip:3128
export KUBECONFIG=./config

kubectl get node
```

## Development

Operator logs:

```
kubectl logs -n clusters deploy/minikube -f
```
