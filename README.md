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

* Make a global settings folder:

```
sudo mkdir /var/mkaas
```

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

Get your Minikube IP either when we copy the .kube/config file down later on, or on the host with this command:

```
sudo -i minikube ip --profile alex
192.168.39.125
```

Note down your minikube IP, for example: `192.168.39.125`.

On your client:

For HTTP access:

```
export http_proxy=http://node_ip:3128
faas-cli list --gateway $MINIKUBE_IP
```

For access via `kubectl`:

Copy the bundle to your client/laptop and untar using (sftp/scp):

```
mkdir -p mkaas
cd mkaas

scp node:/var/mkaas/alex-bundle.tgz .
tar -xvf alex-bundle.tgz
```

If your home directory is `/home/alex/` then do the following:

``` 
sed -ie 's#/root/#/home/alex/mkaas/#g' .kube/config
```

This changes the absolute paths used for the root user to match the point you copied to.

Now:

```
export http_proxy=http://node_ip:3128
export KUBECONFIG=.kube/config

kubectl get node
NAME       STATUS    ROLES     AGE       VERSION
minikube   Ready     master    24m       v1.10.0
```

* Deploy a test workload and access over the proxy

Add the CLI if not present:

```
curl -sLSf https://cli.openfaas.com | sudo sh
```

Deploy OpenFaaS:

```
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

## Development

Operator logs:

```
kubectl logs -n clusters deploy/minikube -f
```
