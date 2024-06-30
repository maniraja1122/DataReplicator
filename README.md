# Data Replicator

## Description
Kubernetes Operator to duplicate ConfigMaps and Secrets across namespaces.

# Install
```
kubectl apply -f https://raw.githubusercontent.com/maniraja1122/datareplicator/master/dist/install.yaml
```
# Working 
### Special annotations for operations
Use this in the original Configmap or Secret that you want to duplicate in the given namespace or namespaces.
```
datareplicator/replicate-to: "namespace1,namespace2" [default: empty]
```
By default, creation of duplicate will be skipped if the destination namespace is not present. This annotation can overwrite this.
```
datareplicator/createnamespace: true [default: false]
```
Depicts that the operator will consider this object (Secret or ConfigMap) as already operated on.
```
datareplicator/replicated: true [default: false]
```
This annotation will be added in the duplicated object for reference.
```
datareplicator/sourcenamespace: “namespaceX”
```


### Tested on
- KinD Cluster
- Minikube Cluster
- [Killercoda Playground](https://killercoda.com/playgrounds/scenario/kubernetes)
- EKS Cluster

### Demo Run
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

### Special Features
- It observes the state of original object and modifies(data only)/deletes the duplicates if the original is modified(data only)/deleted.
- It has the capability of maintaining tree structure like if object in namespace A makes copy in namespace B and namespace B makes copy in namespace C, so the changes on object in A, will be reflected on object in C.
- As this is a replicator, so only make changes in the original object and duplicates should not be manually changed.