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
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

type Options struct {
	APIServerURL string
	KubeConfig   string

	PrometheusAddress string

	LeaseID        string
	LeaseName      string
	LeaseNamespace string

	Version bool
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) MustRegisterFlags() {
	// register klog related flags
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	hostName, err := os.Hostname()
	if err != nil {
		panic("Can not extract hostname for holder identify")
	}

	pflag.StringVar(&o.APIServerURL, "api-server-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.PrometheusAddress, "prometheus-address", "0.0.0.0:3080", "Prometheus metrics api address")
	pflag.StringVar(&o.LeaseID, "lease-id", hostName, "The holder identify name for a nsq-operator instance in a HA environment")
	pflag.StringVar(&o.LeaseName, "lease-name", "nsq-operator", "The lease lock resource name")
	pflag.StringVar(&o.LeaseNamespace, "lease-namespace", "default", "The lease lock resource namespace")
	pflag.BoolVar(&o.Version, "version", false, "Print version")

}

func (o *Options) Parse() {
	pflag.Parse()
}
