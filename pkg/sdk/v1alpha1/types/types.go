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
	"fmt"
	"time"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	"github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/constant"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NsqdConfigRequest struct {
	Name                  string
	Namespace             string
	MessageAvgSize        int32
	MemoryQueueSize       int32
	MemoryOverSalePercent int32
	ChannelCount          int32

	DataPath               *string
	MaxBodySize            *int32
	MaxChannelConsumers    *int32
	MaxMsgSize             *int32
	MaxOutputBufferTimeout *time.Duration
	SyncEvery              *int32
	SyncTimeout            *time.Duration
	StatsdMemStats         *bool
	StatsdInterval         *time.Duration
	Snappy                 *bool
	MaxRequeueTimeout      *time.Duration
	MsgTimeout             *time.Duration
	MaxHeartbeatInterval   *time.Duration

	WaitTimeout *time.Duration
}

func NewNsqdConfigRequest(name string, namespace string, messageAvgSize int32, memoryQueueSize int32,
	memoryOverSalePercent int32, channelCount int32) *NsqdConfigRequest {
	return &NsqdConfigRequest{
		Name:                  name,
		Namespace:             namespace,
		MessageAvgSize:        messageAvgSize,
		MemoryQueueSize:       memoryQueueSize,
		MemoryOverSalePercent: memoryOverSalePercent,
		ChannelCount:          channelCount,
	}
}

func (ndcr *NsqdConfigRequest) String() string {
	return fmt.Sprintf("Name: %v, Namespace: %v, MessageAvgSize: %v, MemoryQueueSize: %v, MemoryOverSalePercent: %v, "+
		"ChannelCount: %v", ndcr.Name, ndcr.Namespace, ndcr.MessageAvgSize, ndcr.MemoryQueueSize, ndcr.MemoryOverSalePercent,
		ndcr.ChannelCount)
}

func (ndcr *NsqdConfigRequest) SetDataPath(dataPath string) {
	ndcr.DataPath = &dataPath
}

func (ndcr *NsqdConfigRequest) SetMaxBodySize(maxBodySize int32) {
	ndcr.MaxBodySize = &maxBodySize
}

func (ndcr *NsqdConfigRequest) SetMaxChannelConsumers(maxChannelConsumers int32) {
	ndcr.MaxChannelConsumers = &maxChannelConsumers
}

func (ndcr *NsqdConfigRequest) SetMaxMsgSize(maxMsgSize int32) {
	ndcr.MaxMsgSize = &maxMsgSize
}

func (ndcr *NsqdConfigRequest) SetMaxOutputBufferTimeout(maxOutputBufferTimeout time.Duration) {
	ndcr.MaxOutputBufferTimeout = &maxOutputBufferTimeout
}

func (ndcr *NsqdConfigRequest) SetSyncEvery(syncEvery int32) {
	ndcr.SyncEvery = &syncEvery
}

func (ndcr *NsqdConfigRequest) SetSyncTimeout(syncTimeout time.Duration) {
	ndcr.SyncTimeout = &syncTimeout
}

func (ndcr *NsqdConfigRequest) SetStatsdMemStats(statsdMemStats bool) {
	ndcr.StatsdMemStats = &statsdMemStats
}

func (ndcr *NsqdConfigRequest) SetStatsdInterval(statsdInterval time.Duration) {
	ndcr.StatsdInterval = &statsdInterval
}

func (ndcr *NsqdConfigRequest) SetSnappy(snappy bool) {
	ndcr.Snappy = &snappy
}

func (ndcr *NsqdConfigRequest) SetMaxRequeueTimeout(maxRequeueTimeout time.Duration) {
	ndcr.MaxRequeueTimeout = &maxRequeueTimeout
}

func (ndcr *NsqdConfigRequest) SetMsgTimeout(msgTimeout time.Duration) {
	ndcr.MsgTimeout = &msgTimeout
}

func (ndcr *NsqdConfigRequest) SetMaxHeartbeatInterval(maxHeartbeatInterval time.Duration) {
	ndcr.MaxHeartbeatInterval = &maxHeartbeatInterval
}

func (ndcr *NsqdConfigRequest) SetWaitTimeout(waitTimeout time.Duration) {
	ndcr.WaitTimeout = &waitTimeout
}

func (ndcr *NsqdConfigRequest) GetMessageAvgSize() int32 {
	return ndcr.MessageAvgSize
}

func (ndcr *NsqdConfigRequest) GetChannelCount() int32 {
	return ndcr.ChannelCount
}

func (ndcr *NsqdConfigRequest) GetMemoryOverSalePercent() int32 {
	return ndcr.MemoryOverSalePercent
}

func (ndcr *NsqdConfigRequest) GetMemoryQueueSize() int32 {
	return ndcr.MemoryQueueSize
}

func (ndcr *NsqdConfigRequest) GetDataPath() string {
	return *ndcr.DataPath
}

func (ndcr *NsqdConfigRequest) GetMaxBodySize() int32 {
	return *ndcr.MaxBodySize
}

func (ndcr *NsqdConfigRequest) GetMaxChannelConsumers() int32 {
	return *ndcr.MaxChannelConsumers
}

func (ndcr *NsqdConfigRequest) GetMaxMsgSize() int32 {
	return *ndcr.MaxMsgSize
}

func (ndcr *NsqdConfigRequest) GetMaxOutputBufferTimeout() time.Duration {
	return *ndcr.MaxOutputBufferTimeout
}

func (ndcr *NsqdConfigRequest) GetSyncEvery() int32 {
	return *ndcr.SyncEvery
}

func (ndcr *NsqdConfigRequest) GetSyncTimeout() time.Duration {
	return *ndcr.SyncTimeout
}

func (ndcr *NsqdConfigRequest) GetStatsdMemStats() bool {
	return *ndcr.StatsdMemStats
}

func (ndcr *NsqdConfigRequest) GetStatsdInterval() time.Duration {
	return *ndcr.StatsdInterval
}

func (ndcr *NsqdConfigRequest) GetSnappy() bool {
	return *ndcr.Snappy
}

func (ndcr *NsqdConfigRequest) GetMaxRequeueTimeout() time.Duration {
	return *ndcr.MaxRequeueTimeout
}

func (ndcr *NsqdConfigRequest) GetMsgTimeout() time.Duration {
	return *ndcr.MsgTimeout
}

func (ndcr *NsqdConfigRequest) GetMaxHeartbeatInterval() time.Duration {
	return *ndcr.MaxHeartbeatInterval
}

func (ndcr *NsqdConfigRequest) GetWaitTimeout() time.Duration {
	return *ndcr.WaitTimeout
}

func (ndcr *NsqdConfigRequest) ApplyDefaults() {
	if ndcr.DataPath == nil {
		ndcr.DataPath = &nsqdCommandDataPath
	}

	if ndcr.MaxBodySize == nil {
		ndcr.MaxBodySize = &nsqdCommandMaxBodySize
	}

	if ndcr.MaxChannelConsumers == nil {
		ndcr.MaxChannelConsumers = &nsqdCommandMaxChannelConsumers
	}

	if ndcr.MaxMsgSize == nil {
		ndcr.MaxMsgSize = &nsqdCommandMaxMsgSize
	}

	if ndcr.MaxOutputBufferTimeout == nil {
		ndcr.MaxOutputBufferTimeout = &nsqdCommandMaxOutputBufferTimeout
	}

	if ndcr.SyncEvery == nil {
		ndcr.SyncEvery = &nsqdCommandSyncEvery
	}

	if ndcr.SyncTimeout == nil {
		ndcr.SyncTimeout = &nsqdCommandSyncTimeout
	}

	if ndcr.StatsdInterval == nil {
		ndcr.StatsdInterval = &nsqdCommandStatsdInterval
	}

	if ndcr.StatsdMemStats == nil {
		ndcr.StatsdMemStats = &nsqdCommandStatsdMemStats
	}

	if ndcr.Snappy == nil {
		ndcr.Snappy = &nsqdCommandSnappy
	}

	if ndcr.MaxRequeueTimeout == nil {
		ndcr.MaxRequeueTimeout = &nsqdCommandMaxRequeueTimeout
	}

	if ndcr.MsgTimeout == nil {
		ndcr.MsgTimeout = &nsqdCommandMsgTimeout
	}

	if ndcr.MaxHeartbeatInterval == nil {
		ndcr.MaxHeartbeatInterval = &nsqdCommandMaxHeartbeatInterval
	}

	if ndcr.WaitTimeout == nil {
		ndcr.WaitTimeout = &waitTimeout
	}
}

func (ndcr *NsqdConfigRequest) AssembleNsqdConfigMap(addresses []string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqdConfigMapName(ndcr.Name),
			Namespace: ndcr.Namespace,
		},
		Data: map[string]string{
			"nsqd": fmt.Sprintf("%s=%q\n%s=%q\n",
				constant.NsqdCommandArguments, ndcr.assembleNsqdCommandArguments(),
				constant.NsqdLookupdTcpAddress, common.AssembleNsqLookupdAddresses(addresses)),
		},
	}
}

func (ndcr *NsqdConfigRequest) assembleNsqdCommandArguments() string {
	return fmt.Sprintf("-statsd-interval=%v "+
		"-statsd-mem-stats=%v "+
		"-statsd-prefix=nsq_cluster_%s.%s "+
		"-http-address=0.0.0.0:%v "+
		"-tcp-address=0.0.0.0:%v "+
		"-max-req-timeout=%v "+
		"-mem-queue-size=%v "+
		"-max-msg-size=%v "+
		"-max-body-size=%v "+
		"-max-heartbeat-interval=%v "+
		"-msg-timeout=%v "+
		"-snappy=%v "+
		"-sync-every=%v "+
		"-sync-timeout=%v "+
		"-data-path=%v", ndcr.GetStatsdInterval(), ndcr.GetStatsdMemStats(),
		ndcr.Name, ndcr.Name, constant.NsqdHttpPort, constant.NsqdTcpPort,
		ndcr.GetMaxRequeueTimeout(), ndcr.GetMemoryQueueSize(), ndcr.GetMaxMsgSize(),
		ndcr.GetMaxBodySize(), ndcr.GetMaxHeartbeatInterval(), ndcr.GetMsgTimeout(), ndcr.GetSnappy(),
		ndcr.GetSyncEvery(), ndcr.GetSyncTimeout(), ndcr.GetDataPath())
}

type NsqCreateRequest struct {
	NsqAdminSpec   v1alpha1.NsqAdminSpec
	NsqLookupdSpec v1alpha1.NsqLookupdSpec
	NsqdSpec       v1alpha1.NsqdSpec
	NsqdScaleSpec  v1alpha1.NsqdScaleSpec

	NsqdConfig *NsqdConfigRequest
}

func NewNsqCreateRequest(nsqdConfig *NsqdConfigRequest, nds v1alpha1.NsqdSpec, nls v1alpha1.NsqLookupdSpec,
	ndss v1alpha1.NsqdScaleSpec, nas v1alpha1.NsqAdminSpec) *NsqCreateRequest {
	return &NsqCreateRequest{
		NsqdSpec:       nds,
		NsqLookupdSpec: nls,
		NsqAdminSpec:   nas,
		NsqdScaleSpec:  ndss,
		NsqdConfig:     nsqdConfig,
	}
}

func (ncr *NsqCreateRequest) AssembleNsqd() *v1alpha1.Nsqd {
	return &v1alpha1.Nsqd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ncr.NsqdConfig.Name,
			Namespace: ncr.NsqdConfig.Namespace,
		},
		Spec: ncr.NsqdSpec,
	}
}

func (ncr *NsqCreateRequest) AssembleNsqdScale() *v1alpha1.NsqdScale {
	return &v1alpha1.NsqdScale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ncr.NsqdConfig.Name,
			Namespace: ncr.NsqdConfig.Namespace,
		},
		Spec: ncr.NsqdScaleSpec,
	}
}

func (ncr *NsqCreateRequest) AssembleNsqLookupdConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqLookupdConfigMapName(ncr.NsqdConfig.Name),
			Namespace: ncr.NsqdConfig.Namespace,
		},
		Data: map[string]string{
			"nsqlookupd": fmt.Sprintf("%s=\"--http-address=0.0.0.0:%v --tcp-address=0.0.0.0:%v\n\"",
				constant.NsqLookupdCommandArguments, constant.NsqLookupdHttpPort, constant.NsqLookupdTcpPort),
		},
	}
}

func (ncr *NsqCreateRequest) AssembleNsqLookupd() *v1alpha1.NsqLookupd {
	return &v1alpha1.NsqLookupd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ncr.NsqdConfig.Name,
			Namespace: ncr.NsqdConfig.Namespace,
		},
		Spec: ncr.NsqLookupdSpec,
	}
}

func (ncr *NsqCreateRequest) AssembleNsqAdmin() *v1alpha1.NsqAdmin {
	return &v1alpha1.NsqAdmin{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ncr.NsqdConfig.Name,
			Namespace: ncr.NsqdConfig.Namespace,
		},
		Spec: ncr.NsqAdminSpec,
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

func (ndr *NsqDeleteRequest) SetWaitTimeout(wt time.Duration) {
	ndr.WaitTimeout = &wt
}

type NsqAdminReplicaUpdateRequest struct {
	Name      string
	Namespace string
	Replicas  int32

	WaitTimeout *time.Duration
}

func NewNsqAdminReplicaUpdateRequest(name string, namespace string, replicas int32) *NsqAdminReplicaUpdateRequest {
	return &NsqAdminReplicaUpdateRequest{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,

		WaitTimeout: &waitTimeout,
	}
}

func (nasr *NsqAdminReplicaUpdateRequest) SetWaitTimeout(wt time.Duration) {
	nasr.WaitTimeout = &wt
}

type NsqLookupdReplicaUpdateRequest struct {
	Name      string
	Namespace string
	Replicas  int32

	WaitTimeout *time.Duration
}

func NewNsqLookupdReplicaUpdateRequest(name string, namespace string, replicas int32) *NsqLookupdReplicaUpdateRequest {
	return &NsqLookupdReplicaUpdateRequest{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,

		WaitTimeout: &waitTimeout,
	}
}

func (nlsr *NsqLookupdReplicaUpdateRequest) SetWaitTimeout(wt time.Duration) {
	nlsr.WaitTimeout = &wt
}

type NsqdReplicaUpdateRequest struct {
	Name      string
	Namespace string
	Replicas  int32

	WaitTimeout *time.Duration
}

func NewNsqdReplicaUpdateRequest(name string, namespace string, replicas int32) *NsqdReplicaUpdateRequest {
	return &NsqdReplicaUpdateRequest{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,

		WaitTimeout: &waitTimeout,
	}
}

func (ndsr *NsqdReplicaUpdateRequest) SetWaitTimeout(wt time.Duration) {
	ndsr.WaitTimeout = &wt
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

func (nauir *NsqAdminUpdateImageRequest) SetWaitTimeout(wt time.Duration) {
	nauir.WaitTimeout = &wt
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

func (nluir *NsqLookupdUpdateImageRequest) SetWaitTimeout(wt time.Duration) {
	nluir.WaitTimeout = &wt
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

func (nduir *NsqdUpdateImageRequest) SetWaitTimeout(wt time.Duration) {
	nduir.WaitTimeout = &wt
}

type NsqdScaleUpdateRequest struct {
	Name         string
	Namespace    string
	QpsThreshold int32
	Minimum      int32
	Maximum      int32

	WaitTimeout *time.Duration
}

func NewNsqdScaleUpdateRequest(name string, namespace string, qpsThreshold int32,
	minimum int32, maximum int32) *NsqdScaleUpdateRequest {
	return &NsqdScaleUpdateRequest{
		Name:         name,
		Namespace:    namespace,
		QpsThreshold: qpsThreshold,
		Minimum:      minimum,
		Maximum:      maximum,

		WaitTimeout: &waitTimeout,
	}
}

func (ndsur *NsqdScaleUpdateRequest) SetWaitTimeout(wt time.Duration) {
	ndsur.WaitTimeout = &wt
}
