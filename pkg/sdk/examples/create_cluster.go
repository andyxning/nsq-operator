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

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	sdkv1alpha1 "github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1"
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/types"
	"github.com/spf13/pflag"
)

func main() {
	var name string
	var namespace string

	common.RegisterFlags()

	pflag.StringVar(&name, "name", "solo", "Cluster name")
	pflag.StringVar(&namespace, "namespace", "default", "Cluster namespace")

	common.Parse()

	kubeClient, nsqClient, err := common.InitClients()
	if err != nil {
		klog.Fatalf("Init clients error: %v", err)
	}

	var nsqdReplicas int32 = 2
	var nsqLookupdReplicas int32 = 2
	var nsqAdminReplicas int32 = 2
	messageAvgSize := 1024

	nds := v1alpha1.NsqdSpec{
		Image:            "dockerops123/nsqd:1.1.0",
		Replicas:         &nsqdReplicas,
		StorageClassName: "standard",
		LogMappingDir:    "/var/log/",
	}

	nls := v1alpha1.NsqLookupdSpec{
		Image:         "dockerops123/nsqlookupd:1.1.0",
		Replicas:      &nsqLookupdReplicas,
		LogMappingDir: "/var/log/",
	}

	nas := v1alpha1.NsqAdminSpec{
		Image:         "dockerops123/nsqadmin:1.1.0",
		Replicas:      &nsqAdminReplicas,
		LogMappingDir: "/var/log/",
	}

	nr := types.NewNsqCreateRequest(name, namespace, messageAvgSize, nds, nls, nas)
	// Do whatever other fields setting here
	//nr.NsqdCommandMaxMsgSize = 2 * 1024 * 1024
	nr.ApplyDefaults()

	err = sdkv1alpha1.CreateCluster(kubeClient, nsqClient, nr)
	if err != nil {
		klog.Fatalf("Create nsq cluster %s/%s error: %v", namespace, name, err)
	}

	klog.Infof("Create nsq cluster %s/%s success", namespace, name)
}
