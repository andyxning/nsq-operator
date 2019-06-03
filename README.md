# nsq-operator
**nsq-operator** targets to run NSQ-As-A-Service, i.e., NAAS, based on [Kubernetes](https://kubernetes.io/).

# Status
**nsq-operator** is currently in **Alpha** stage. Until **GA** status is clearly stated, it is not recommended to run 
**nsq-operator** in production. 

But, we encourage end users to give **nsq-operator** a try, i.e., testing and benching **nsq-operator**, and give us 
some feedback.

# Feature
* High availability. NSQ-Operator supports HA by utilizing Kubernetes 
[leaderelection](https://github.com/kubernetes/client-go/tree/master/tools/leaderelection) package and 
[coordination.k8s.io/lease](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#lease-v1-coordination-k8s-io) resource.
* Create/Delete a nsq cluster
* Scale nsqd/nsqlookupd/nsqadmin components separately
* Update nsqd/nsqlookupd/nsqadmin image separately
* Adjust nsqd command line arguments
* Adjust nsqd pods memory resource according to average message size, memory queue size, memory oversale percent and channel count
* Log Management. 
  * Rotate log by [logrotate](https://linux.die.net/man/8/logrotate) hourly
  * Mount log directory to a dedicated host machine directory in nsqd/nsqlookupd/nsqadmin spec
* QPS based horizontal pod autoscale. No memory/cpu/network/disk resources based autoscale support. QPS autoscale does 
not rely on any other systems. It is implemented in Kubernetes way. [More details](docs/auto_scale.md)

# Runtime Requirement
* Kubernetes >= 1.14.0
* Golang >= 1.12.0

# SDK
There is a [sdk](pkg/sdk/v1alpha1) which can be used to CRUD nsq related resource. Files under 
[examples](pkg/sdk/examples) show sdk use examples.

|   File   |   Description
|----------|---------------------------
| [create_cluster.go](pkg/sdk/examples/create_cluster.go) | Create a NSQ cluster
| [delete_cluster.go](pkg/sdk/examples/delete_cluster.go) | Delete a NSQ cluster
| [update_nsqadmin_replica.go](pkg/sdk/examples/update_nsqadmin_replica.go) | Update nsqadmin resource object replica
| [update_nsqlookupd_replica.go](pkg/sdk/examples/update_nsqlookupd_replica.go) | Update nsqlookupd resource object replica
| [update_nsqd_replica.go](pkg/sdk/examples/update_nsqd_replica.go) | Update nsqd resource object replica
| [adjust_nsqd_config.go](pkg/sdk/examples/adjust_nsqd_config.go) | Adjust nsqd command line arguments
| [adjust_nsqd_memory_resource.go](pkg/sdk/examples/adjust_nsqd_memory_resource.go) | Adjust nsqd resource object's memory resource limits/requests
| [adjust_nsqdscale.go](pkg/sdk/examples/adjust_nsqdscale.go) | Adjust nsqdscale resource object
| [bump_nsqadmin_image.go](pkg/sdk/examples/bump_nsqadmin_image.go) | Update a NSQ cluster's nsqadmin image
| [bump_nsqlookupd_image.go](pkg/sdk/examples/bump_nsqlookupd_image.go) | Update a NSQ cluster's nsqlookupd image
| [bump_nsqd_image.go](pkg/sdk/examples/bump_nsqd_image.go) | Update a NSQ cluster's nsqd image
 
# FAQ
See [FAQ](FAQ.md)

# Users
See [Users](USERS.md)  

