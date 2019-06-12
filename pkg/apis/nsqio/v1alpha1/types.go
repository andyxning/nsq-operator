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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Nsqd is a specification for a Nsqd resource
type Nsqd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NsqdSpec   `json:"spec"`
	Status NsqdStatus `json:"status"`
}

// NsqdSpec is the spec for a Nsqd resource
type NsqdSpec struct {
	Image                    string `json:"image"`
	Replicas                 int32  `json:"replicas"`
	StorageClassName         string `json:"storageClassName"`
	LogMappingDir            string `json:"logMappingDir"`
	MessageAvgSize           int32  `json:"messageAvgSize"`
	MemoryOverBookingPercent int32  `json:"memoryOverBookingPercent"`
	MemoryQueueSize          int32  `json:"memoryQueueSize"`
	ChannelCount             int32  `json:"channelCount"`
}

// NsqdStatus is the status for a Nsqd resource
type NsqdStatus struct {
	AvailableReplicas        int32 `json:"availableReplicas"`
	MessageAvgSize           int32 `json:"messageAvgSize"`
	MemoryOverBookingPercent int32 `json:"memoryOverBookingPercent"`
	MemoryQueueSize          int32 `json:"memoryQueueSize"`
	ChannelCount             int32 `json:"channelCount"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqdList is a list of Nsqd resources
type NsqdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Nsqd `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqdScale is a specification for a NsqdScale resource
type NsqdScale struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NsqdScaleSpec   `json:"spec"`
	Status NsqdScaleStatus `json:"status"`
}

// NsqdScaleSpec is the spec for a NsqdScale resource
type NsqdScaleSpec struct {
	QpsThreshold int32 `json:"qpsThreshold"`
	Minimum      int32 `json:"minimum"`
	Maximum      int32 `json:"maximum"`
	Enabled      bool  `json:"enabled"`
}

// NsqdScaleStatus is the status for a NsqdScale resource
type NsqdScaleStatus struct {
	LastScaleUpTime   metav1.Time       `json:"lastScaleUpTime"`
	LastScaleDownTime metav1.Time       `json:"lastScaleDownTime"`
	Metas             map[string][]Meta `json:"metas"`
}

type Meta struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	Qps            int64       `json:"qps"`
	Depth          int64       `json:"depth"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqdScaleList is a list of NsqdScale resources
type NsqdScaleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NsqdScale `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqLookupd is a specification for a NsqLookupd resource
type NsqLookupd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NsqLookupdSpec   `json:"spec"`
	Status NsqLookupdStatus `json:"status"`
}

// NsqLookupdSpec is the spec for a NsqLookupd resource
type NsqLookupdSpec struct {
	Image         string `json:"image"`
	Replicas      int32  `json:"replicas"`
	LogMappingDir string `json:"logMappingDir"`
}

// NsqLookupdStatus is the status for a NsqLookupd resource
type NsqLookupdStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqLookupdList is a list of NsqLookupd resources
type NsqLookupdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NsqLookupd `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqAdmin is a specification for a NsqAdmin resource
type NsqAdmin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NsqAdminSpec   `json:"spec"`
	Status NsqAdminStatus `json:"status"`
}

// NsqAdminSpec is the spec for a NsqAdmin resource
type NsqAdminSpec struct {
	Image         string `json:"image"`
	Replicas      int32  `json:"replicas"`
	LogMappingDir string `json:"logMappingDir"`
}

// NsqAdminStatus is the status for a NsqAdmin resource
type NsqAdminStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsqAdminList is a list of NsqAdmin resources
type NsqAdminList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NsqAdmin `json:"items"`
}
