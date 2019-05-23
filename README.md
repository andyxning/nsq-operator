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
| [adjust_nsqd_memory_resource.go](pkg/sdk/examples/adjust_nsqd_memory_resource.go) | Adjust nsqd resource object's memory resource limits/requests
| [bump_nsqadmin_image.go](pkg/sdk/examples/bump_nsqadmin_image.go) | Update a NSQ cluster's nsqadmin image
| [bump_nsqlookupd_image.go](pkg/sdk/examples/bump_nsqlookupd_image.go) | Update a NSQ cluster's nsqlookupd image
| [bump_nsqd_image.go](pkg/sdk/examples/bump_nsqd_image.go) | Update a NSQ cluster's nsqd image
 
# FAQ
See [FAQ](FAQ.md)

# Users
See [Users](USERS.md)  

