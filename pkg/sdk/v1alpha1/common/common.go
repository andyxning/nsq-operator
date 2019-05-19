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

package common

import (
	"fmt"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	"github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/constant"
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AssembleNsqLookupdConfigMap(nr *types.NsqCreateRequest) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqLookupdConfigMapName(nr.Name),
			Namespace: nr.Namespace,
		},
		Data: map[string]string{
			"nsqlookupd": fmt.Sprintf("%s=\"--http-address=0.0.0.0:%v --tcp-address=0.0.0.0:%v\n\"",
				constant.NsqLookupdCommandArguments, constant.NsqLookupdHttpPort, constant.NsqLookupdTcpPort),
		},
	}
}

func AssembleNsqdConfigMap(nr *types.NsqCreateRequest, addresses []string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqdConfigMapName(nr.Name),
			Namespace: nr.Namespace,
		},
		Data: map[string]string{
			"nsqd": fmt.Sprintf("%s=%q\n%s=%q\n",
				constant.NsqdCommandArguments, assembleNsqdCommandArguments(nr),
				constant.NsqdLookupdTcpAddress, common.AssembleNsqLookupdAddresses(addresses)),
		},
	}
}

func assembleNsqdCommandArguments(nr *types.NsqCreateRequest) string {
	return fmt.Sprintf("-statsd-interval=%v "+
		"-statsd-mem-stats=%v "+
		"-statsd-prefix=nsq_cluster_%s.%s "+
		"-http-address=0.0.0.0:%v "+
		"-tcp-address=0.0.0.0:%v "+
		"-max-req-timeout=%v "+
		"-mem-queue-size=%v "+
		"-max-msg-size=%v "+
		"-max-body-size=%v "+
		"-sync-every=%v "+
		"-sync-timeout=%v "+
		"-data-path=%v", *nr.NsqdCommandStatsdInterval, *nr.NsqdCommandStatsdMemStats, nr.Name, nr.Name,
		constant.NsqdHttpPort, constant.NsqdTcpPort,
		*nr.NsqdCommandMaxRequeueTimeout, *nr.TopicMemoryQueueSize, *nr.NsqdCommandMaxMsgSize,
		*nr.NsqdCommandMaxBodySize, *nr.NsqdCommandSyncEvery, *nr.NsqdCommandSyncTimeout, *nr.NsqdCommandDataPath)
}

func AssembleNsqLookupd(nr *types.NsqCreateRequest) *v1alpha1.NsqLookupd {
	return &v1alpha1.NsqLookupd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nr.Name,
			Namespace: nr.Namespace,
		},
		Spec: nr.NsqLookupdSpec,
	}
}

func AssembleNsqAdmin(nr *types.NsqCreateRequest) *v1alpha1.NsqAdmin {
	return &v1alpha1.NsqAdmin{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nr.Name,
			Namespace: nr.Namespace,
		},
		Spec: nr.NsqAdminSpec,
	}
}

func AssembleNsqd(nr *types.NsqCreateRequest) *v1alpha1.Nsqd {
	return &v1alpha1.Nsqd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nr.Name,
			Namespace: nr.Namespace,
		},
		Spec: nr.NsqdSpec,
	}
}
