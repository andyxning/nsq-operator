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
	var snappy bool

	common.RegisterFlags()

	pflag.StringVar(&name, "name", "solo", "Cluster name")
	pflag.StringVar(&namespace, "namespace", "default", "Cluster namespace")
	pflag.BoolVar(&snappy, "snappy", true, "Consumer snappy compression enabled")

	messageAvgSize := 1024

	common.Parse()

	kubeClient, _, err := common.InitClients()
	if err != nil {
		klog.Fatalf("Init clients error: %v", err)
	}

	ndcr := types.NewNsqdConfigRequest(name, namespace, messageAvgSize)
	ndcr.ApplyDefaults()
	ndcr.SetSnappy(snappy)

	// Customize wait timeout
	//wt := 180 * time.Second
	//ndcr.SetWaitTimeout(wt)

	err = sdkv1alpha1.AdjustNsqdConfig(kubeClient, ndcr)
	if err != nil {
		klog.Fatalf("Update nsqd %s/%s config to %#v error: %v", ndcr.Namespace, ndcr.Name, ndcr.AssembleNsqdConfigMap([]string{}).Data, err)
	}

	klog.Infof("Update nsqd %s/%s config to %#v success", ndcr.Namespace, ndcr.Name, ndcr.AssembleNsqdConfigMap([]string{}).Data)
}
