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

package types

import (
	"time"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
)

type NsqCreateRequest struct {
	NsqAdminSpec   v1alpha1.NsqAdminSpec
	NsqLookupdSpec v1alpha1.NsqLookupdSpec
	NsqdSpec       v1alpha1.NsqdSpec

	Name                      string
	Namespace                 string
	MessageAvgSize            int
	NsqdMemoryOverSalePercent *int
	TopicMemoryQueueSize      *int
	ChannelMemoryQueueSize    *int

	NsqdCommandDataPath               *string
	NsqdCommandMaxBodySize            *int
	NsqdCommandMaxChannelConsumers    *int
	NsqdCommandMaxMsgSize             *int
	NsqdCommandMaxOutputBufferTimeout *time.Duration
	NsqdCommandSyncEvery              *int
	NsqdCommandSyncTimeout            *time.Duration
	NsqdCommandStatsdMemStats         *bool
	NsqdCommandStatsdInterval         *time.Duration
	NsqdCommandSnappy                 *bool
	NsqdCommandMaxRequeueTimeout      *time.Duration
	NsqdCommandMsgTimeout             *time.Duration
	NsqdCommandMaxHeartbeatInterval   *time.Duration

	WaitTimeout *time.Duration
}

func NewNsqCreateRequest(name string, namespace string, messageAvgSize int,
	nds v1alpha1.NsqdSpec, nls v1alpha1.NsqLookupdSpec, nas v1alpha1.NsqAdminSpec) *NsqCreateRequest {
	return &NsqCreateRequest{
		NsqdSpec:       nds,
		NsqLookupdSpec: nls,
		NsqAdminSpec:   nas,

		Name:           name,
		Namespace:      namespace,
		MessageAvgSize: messageAvgSize,
	}
}

func (nr *NsqCreateRequest) ApplyDefaults() {
	if nr.TopicMemoryQueueSize == nil {
		nr.TopicMemoryQueueSize = &topicMemoryQueueSize
	}

	if nr.ChannelMemoryQueueSize == nil {
		nr.ChannelMemoryQueueSize = &channelMemoryQueueSize
	}

	if nr.NsqdCommandDataPath == nil {
		nr.NsqdCommandDataPath = &nsqdCommandDataPath
	}

	if nr.NsqdCommandMaxBodySize == nil {
		nr.NsqdCommandMaxBodySize = &nsqdCommandMaxBodySize
	}

	if nr.NsqdCommandMaxChannelConsumers == nil {
		nr.NsqdCommandMaxChannelConsumers = &nsqdCommandMaxChannelConsumers
	}

	if nr.NsqdCommandMaxMsgSize == nil {
		nr.NsqdCommandMaxMsgSize = &nsqdCommandMaxMsgSize
	}

	if nr.NsqdCommandMaxOutputBufferTimeout == nil {
		nr.NsqdCommandMaxOutputBufferTimeout = &nsqdCommandMaxOutputBufferTimeout
	}

	if nr.NsqdCommandSyncEvery == nil {
		nr.NsqdCommandSyncEvery = &nsqdCommandSyncEvery
	}

	if nr.NsqdCommandSyncTimeout == nil {
		nr.NsqdCommandSyncTimeout = &nsqdCommandSyncTimeout
	}

	if nr.NsqdCommandStatsdInterval == nil {
		nr.NsqdCommandStatsdInterval = &nsqdCommandStatsdInterval
	}

	if nr.NsqdCommandStatsdMemStats == nil {
		nr.NsqdCommandStatsdMemStats = &nsqdCommandStatsdMemStats
	}

	if nr.NsqdCommandSnappy == nil {
		nr.NsqdCommandSnappy = &nsqdCommandSnappy
	}

	if nr.NsqdCommandMaxRequeueTimeout == nil {
		nr.NsqdCommandMaxRequeueTimeout = &nsqdCommandMaxRequeueTimeout
	}

	if nr.NsqdCommandMsgTimeout == nil {
		nr.NsqdCommandMsgTimeout = &nsqdCommandMsgTimeout
	}

	if nr.NsqdCommandMaxHeartbeatInterval == nil {
		nr.NsqdCommandMaxHeartbeatInterval = &nsqdCommandMaxHeartbeatInterval
	}

	if nr.NsqdMemoryOverSalePercent == nil {
		nr.NsqdMemoryOverSalePercent = &nsqdMemoryOverSalePercent
	}

	if nr.WaitTimeout == nil {
		nr.WaitTimeout = &waitTimeout
	}
}

type NsqDeleteRequest struct {
	Name      string
	Namespace string

	WaitTimeout *time.Duration
}

func NewNsqDeleteRequest(name string, namespace string) *NsqDeleteRequest {
	return &NsqDeleteRequest{
		Name:      name,
		Namespace: namespace,

		WaitTimeout: &waitTimeout,
	}
}

func (ndr *NsqDeleteRequest) SetWaitTimeout(wt *time.Duration) {
	ndr.WaitTimeout = wt
}

type NsqAdminScaleRequest struct {
	Name      string
	Namespace string
	Replicas  int32

	WaitTimeout *time.Duration
}

func NewNsqAdminScaleRequest(name string, namespace string, replicas int32) *NsqAdminScaleRequest {
	return &NsqAdminScaleRequest{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,

		WaitTimeout: &waitTimeout,
	}
}

func (nasr *NsqAdminScaleRequest) SetWaitTimeout(wt *time.Duration) {
	nasr.WaitTimeout = wt
}

type NsqLookupdScaleRequest struct {
	Name      string
	Namespace string
	Replicas  int32

	WaitTimeout *time.Duration
}

func NewNsqLookupdScaleRequest(name string, namespace string, replicas int32) *NsqLookupdScaleRequest {
	return &NsqLookupdScaleRequest{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,

		WaitTimeout: &waitTimeout,
	}
}

func (nlsr *NsqLookupdScaleRequest) SetWaitTimeout(wt *time.Duration) {
	nlsr.WaitTimeout = wt
}

type NsqdScaleRequest struct {
	Name      string
	Namespace string
	Replicas  int32

	WaitTimeout *time.Duration
}

func NewNsqdScaleRequest(name string, namespace string, replicas int32) *NsqdScaleRequest {
	return &NsqdScaleRequest{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,

		WaitTimeout: &waitTimeout,
	}
}

func (ndsr *NsqdScaleRequest) SetWaitTimeout(wt *time.Duration) {
	ndsr.WaitTimeout = wt
}

type NsqdAddChannelRequest struct {
	Name      string
	Namespace string

	WaitTimeout *time.Duration
}

func NewNsqdAddChannelRequest(name string, namespace string) *NsqdAddChannelRequest {
	return &NsqdAddChannelRequest{
		Name:      name,
		Namespace: namespace,

		WaitTimeout: &waitTimeout,
	}
}

func (ndac *NsqdAddChannelRequest) SetWaitTimeout(wt *time.Duration) {
	ndac.WaitTimeout = wt
}

type NsqdDeleteChannelRequest struct {
	Name      string
	Namespace string

	WaitTimeout *time.Duration
}

func NewNsqdDeleteChannelRequest(name string, namespace string) *NsqdDeleteChannelRequest {
	return &NsqdDeleteChannelRequest{
		Name:      name,
		Namespace: namespace,

		WaitTimeout: &waitTimeout,
	}
}

func (nddc *NsqdDeleteChannelRequest) SetWaitTimeout(wt *time.Duration) {
	nddc.WaitTimeout = wt
}

type NsqAdminUpdateImageRequest struct {
	Name      string
	Namespace string
	Image     string

	WaitTimeout *time.Duration
}

func NewNsqAdminUpdateImageRequest(name string, namespace string, image string) *NsqAdminUpdateImageRequest {
	return &NsqAdminUpdateImageRequest{
		Name:      name,
		Namespace: namespace,
		Image:     image,

		WaitTimeout: &waitTimeout,
	}
}

func (nauir *NsqAdminUpdateImageRequest) SetWaitTimeout(wt *time.Duration) {
	nauir.WaitTimeout = wt
}

type NsqLookupdUpdateImageRequest struct {
	Name      string
	Namespace string
	Image     string

	WaitTimeout *time.Duration
}

func NewNsqLookupdUpdateImageRequest(name string, namespace string, image string) *NsqLookupdUpdateImageRequest {
	return &NsqLookupdUpdateImageRequest{
		Name:      name,
		Namespace: namespace,
		Image:     image,

		WaitTimeout: &waitTimeout,
	}
}

func (nluir *NsqLookupdUpdateImageRequest) SetWaitTimeout(wt *time.Duration) {
	nluir.WaitTimeout = wt
}

type NsqdUpdateImageRequest struct {
	Name      string
	Namespace string
	Image     string

	WaitTimeout *time.Duration
}

func NewNsqdUpdateImageRequest(name string, namespace string, image string) *NsqdUpdateImageRequest {
	return &NsqdUpdateImageRequest{
		Name:      name,
		Namespace: namespace,
		Image:     image,

		WaitTimeout: &waitTimeout,
	}
}

func (nduir *NsqdUpdateImageRequest) SetWaitTimeout(wt *time.Duration) {
	nduir.WaitTimeout = wt
}
