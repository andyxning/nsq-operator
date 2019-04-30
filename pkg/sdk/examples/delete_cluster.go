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
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/types"
	"k8s.io/klog"

	sdkv1alpha1 "github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1"
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

	ndr := types.NewNsqDeleteRequest(name, namespace)

	err = sdkv1alpha1.DeleteCluster(kubeClient, nsqClient, ndr)
	if err != nil {
		klog.Fatalf("Delete nsq cluster %s/%s error: %v", namespace, name, err)
	}

	klog.Infof("Delete nsq cluster %s/%s success", namespace, name)
}
