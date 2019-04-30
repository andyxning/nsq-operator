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
	var replicas int32

	common.RegisterFlags()

	pflag.StringVar(&name, "name", "solo", "Cluster name")
	pflag.StringVar(&namespace, "namespace", "default", "Cluster namespace")
	pflag.Int32Var(&replicas, "replicas", 2, "Replicas")

	common.Parse()

	_, nsqClient, err := common.InitClients()
	if err != nil {
		klog.Fatalf("Init clients error: %v", err)
	}

	ndsr := types.NewNsqdScaleRequest(name, namespace, replicas)

	// Customize wait timeout
	//wt := 180 * time.Second
	//ndsr.SetWaitTimeout(&wt)

	err = sdkv1alpha1.ScaleNsqd(nsqClient, ndsr)
	if err != nil {
		klog.Fatalf("Scale nsqd %s/%s to replicas %d error: %v", ndsr.Namespace, ndsr.Name, ndsr.Replicas, err)
	}

	klog.Infof("Scale nsqd %s/%s to replicas %d success", ndsr.Namespace, ndsr.Name, ndsr.Replicas)
}
