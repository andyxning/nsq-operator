# nsq-operator
**nsq-operator** targets to run NSQ-As-A-Service, i.e., NAAS, based on [Kubernetes](https://kubernetes.io/).

# Status
**nsq-operator** is currently in **Alpha** stage. Until **GA** status is clearly stated, it is not recommended to run 
**nsq-operator** in production. 

But, we encourage end users to give **nsq-operator** a try, i.e., testing and benching **nsq-operator**, and give us 
some feedback.

# Runtime Requirements
* Kubernetes >= 1.14.0
* Golang >= 1.12.0

# SDK
There is a [sdk](pkg/sdk/v1alpha1) which can be used to CRUD nsq related resource. Files under 
[examples](pkg/sdk/examples) show sdk use examples.

|   File   |   Description
|----------|---------------------------
| [create_cluster.go](pkg/sdk/examples/create_cluster.go) | Create a NSQ cluster
| [delete_cluster.go](pkg/sdk/examples/delete_cluster.go) | Delete a NSQ cluster
| [scale_nsqadmin.go](pkg/sdk/examples/scale_nsqadmin.go) | Scale a NSQ cluster's nsqadmin resource object
| [scale_nsqlookupd.go](pkg/sdk/examples/scale_nsqlookupd.go) | Scale a NSQ cluster's nsqlookupd resource object
| [scale_nsqd.go](pkg/sdk/examples/scale_nsqd.go) | Scale a NSQ cluster's nsqd resource object
| [add_channel.go](pkg/sdk/examples/add_channel.go) | Add a channel, i.e., adjust nsqd resource object's memory resource limits/requests
| [delete_channel.go](pkg/sdk/examples/delete_channel.go) | Delete a channel, i.e., adjust nsqd resource object's memory resource limits/requests
 
 
# FAQ
See [FAQ](FAQ.md)

# Users
See [Users](USERS.md)  

