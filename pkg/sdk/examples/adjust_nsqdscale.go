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
	var minimum int32
	var maximum int32
	var qpsThreshold int32
	var enabled bool

	common.RegisterFlags()

	pflag.StringVar(&name, "name", "solo", "Cluster name")
	pflag.StringVar(&namespace, "namespace", "default", "Cluster namespace")
	pflag.Int32Var(&qpsThreshold, "qps-threshold", 15000, "Metas threshold before autoscaling")
	pflag.Int32Var(&minimum, "minimum", 2, "Minimum nsqd instances")
	pflag.Int32Var(&maximum, "maximum", 4, "Maximum nsqd instances")
	pflag.BoolVar(&enabled, "enabled", false, "Whether nsqdscale is enabled")

	common.Parse()

	_, nsqClient, err := common.InitClients()
	if err != nil {
		klog.Fatalf("Init clients error: %v", err)
	}

	ndsur := types.NewNsqdScaleUpdateRequest(name, namespace, qpsThreshold, minimum, maximum, enabled)

	// Customize wait timeout
	//wt := 180 * time.Second
	//ndsur.SetWaitTimeout(wt)

	err = sdkv1alpha1.AdjustNsqdScale(nsqClient, ndsur)
	if err != nil {
		klog.Fatalf("Update nsqdscale %s/%s to %#v error: %v", ndsur.Namespace, ndsur.Name, ndsur, err)
	}

	klog.Infof("Update nsqdscale %s/%s to %#v success", ndsur.Namespace, ndsur.Name, ndsur)
}
