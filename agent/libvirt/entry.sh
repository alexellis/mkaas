#!/bin/bash

exit_script() {
    echo "Stopping cluster $CLUSTER_NAME"
    minikube delete --profile $CLUSTER_NAME
#    virsh destroy $CLUSTER_NAME
#    virsh undefine $CLUSTER_NAME

    trap - SIGINT SIGTERM # clear the trap
}

trap exit_script SIGINT SIGTERM

minikube start --bootstrapper=kubeadm --vm-driver=kvm2 --memory $CLUSTER_MEMORY --cpus $CLUSTER_CPUS --profile $CLUSTER_NAME

cd /root/
tar -czf /var/mkaas/$CLUSTER_NAME-bundle.tgz .minikube/*.crt .minikube/*.key .minikube/*.pem .kube/config
echo "/var/mkaas/$CLUSTER_NAME-bundle.tgz written."

while [ true ] ;
do
   sleep 5
done
