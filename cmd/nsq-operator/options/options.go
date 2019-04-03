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

package options

import (
	"flag"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

type options struct {
	APIServerURL string
	KubeConfig   string

	PrometheusAddress string

	Version bool
}

func NewOptions() *options {
	return &options{}
}

func (o *options) RegisterFlags() {
	// register klog related flags
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.StringVar(&o.APIServerURL, "api_server_url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.PrometheusAddress, "prometheus_address", "0.0.0.0:3080", "Prometheus metrics api address")
	pflag.BoolVar(&o.Version, "version", false, "Print version")

}

func (o *options) Parse() {
	pflag.Parse()
}
