# Data Replicator

## Description
Kubernetes Operator to duplicate ConfigMaps and Secrets across namespaces.

# Install
```
kubectl apply -f https://raw.githubusercontent.com/maniraja1122/datareplicator/master/dist/install.yaml
```

### Tested on
[x] KinD Cluster
[x] Minikube Cluster
[x] Killercoda Playground (https://killercoda.com/playgrounds/scenario/kubernetes)
[x] EKS Cluster

Sample Test:
```
k create cm userdata --from-literal=name=zizu
k annotate cm userdata datareplicator/replicate-to="namespace1,namespace2"
k annotate cm userdata datareplicator/createnamespace=true
k get cm -A
```

Sample Output:
```
NAMESPACE               NAME                                                   DATA   AGE
datareplicator-system   kube-root-ca.crt                                       1      100s
default                 kube-root-ca.crt                                       1      21d
default                 userdata                                               1      4m36s
kube-node-lease         kube-root-ca.crt                                       1      21d
kube-public             cluster-info                                           2      21d
kube-public             kube-root-ca.crt                                       1      21d
kube-system             canal-config                                           6      21d
kube-system             coredns                                                1      21d
kube-system             extension-apiserver-authentication                     6      21d
kube-system             kube-apiserver-legacy-service-account-token-tracking   1      21d
kube-system             kube-proxy                                             2      21d
kube-system             kube-root-ca.crt                                       1      21d
kube-system             kubeadm-config                                         1      21d
kube-system             kubelet-config                                         1      21d
local-path-storage      kube-root-ca.crt                                       1      21d
local-path-storage      local-path-config                                      4      21d
namespace1              kube-root-ca.crt                                       1      98s
namespace1              userdata                                               1      98s
namespace2              kube-root-ca.crt                                       1      98s
namespace2              userdata                                               1      98s
```