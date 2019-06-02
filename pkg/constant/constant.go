/*
Copyright 2019 The NSQ-Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constant

const (
	// Controller name
	NsqdControllerName       = "nsqd-controller"
	NsqdScaleControllerName  = "nsqdscale-controller"
	NsqLookupdControllerName = "nsqlookupd-controller"
	NsqAdminControllerName   = "nsqadmin-controller"

	// Cluster environment Name
	ClusterNameEnv = "NSQ_CLUSTER"

	LogVolumeName = "log"

	NsqConfigMapMountPath = "/etc/nsq"
	NsqdDataMountPath     = "/data"

	NsqConfigMapAnnotationKey = "configmap/signature"
)

type NsqdScaleOperation int

const (
	Unknown NsqdScaleOperation = iota
	ScaleUp
	ScaleDown
)
