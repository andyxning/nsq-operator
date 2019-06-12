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

package main

import (
	"github.com/andyxning/nsq-operator/pkg/sdk/examples/common"
	"k8s.io/klog"

	sdkv1alpha1 "github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1"
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/types"
	"github.com/spf13/pflag"
)

func main() {
	var name string
	var namespace string
	var messageAvgSize int32
	var memoryQueueSize int32
	var memoryOverBookingPercent int32
	var channelCount int32 = 2

	common.RegisterFlags()

	pflag.StringVar(&name, "name", "solo", "Cluster name")
	pflag.StringVar(&namespace, "namespace", "default", "Cluster namespace")
	pflag.Int32Var(&messageAvgSize, "message_avg_size", 1204, "Average message size")
	pflag.Int32Var(&memoryQueueSize, "memory_queue_size", 10000, "Memory queue size")
	pflag.Int32Var(&memoryOverBookingPercent, "memory_overbooking_percent", 50, "Memory overbooking percent")
	pflag.Int32Var(&channelCount, "channel_count", 1, "Channel count")

	common.Parse()

	_, nsqClient, err := common.InitClients()
	if err != nil {
		klog.Fatalf("Init clients error: %v", err)
	}

	ndcr := types.NewNsqdConfigRequest(name, namespace, messageAvgSize, memoryQueueSize, memoryOverBookingPercent, channelCount)
	ndcr.ApplyDefaults()

	// Customize wait timeout
	//wt := 180 * time.Second
	//ndcr.SetWaitTimeout(wt)

	err = sdkv1alpha1.AdjustNsqdMemoryResources(nsqClient, ndcr)
	if err != nil {
		klog.Fatalf("Update nsqd %s/%s config to %s error: %v", ndcr.Namespace, ndcr.Name, ndcr.String(), err)
	}

	klog.Infof("Update nsqd %s/%s config to %s success", ndcr.Namespace, ndcr.Name, ndcr.String())
}
