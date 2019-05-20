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
	// DeploymentResourceNotOwnedByNsqAdmin is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	DeploymentResourceNotOwnedByNsqAdmin = "Deployment %q already exists and is not managed by NsqAdmin"
	// DeploymentResourceNotOwnedByNsqLookupd is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	DeploymentResourceNotOwnedByNsqLookupd = "Deployment %q already exists and is not managed by NsqLookupd"
	// ConfigMapResourceNotOwnedByNsqLookupd is the message used for Events when a resource
	// fails to sync due to a ConfigMap already existing
	ConfigMapResourceNotOwnedByNsqLookupd = "ConfigMap %q already exists and is not managed by NsqLookupd"
	// StatefulSetResourceNotOwnedByNsqd is the message used for Events when a resource
	// fails to sync due to a StatefulSet already existing
	StatefulSetResourceNotOwnedByNsqd = "StatefulSet %q already exists and is not managed by Nsqd"
)
