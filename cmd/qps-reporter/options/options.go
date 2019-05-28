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
	"time"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

type Options struct {
	APIServerURL string
	KubeConfig   string

	NsqdApiAddress     string
	Topic              string
	Namespace          string
	InstanceName       string
	HttpRequestTimeout time.Duration
	UpdatePeriod       time.Duration
	PreservedQpsCount  int
	HttpApiAddress     string

	DryRun bool

	Version bool
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) MustRegisterFlags() {
	// register klog related flags
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.StringVar(&o.APIServerURL, "api-server-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")

	pflag.StringVar(&o.NsqdApiAddress, "nsqd-api-address", "0.0.0.0:4151", "Nsqd http api address")
	pflag.StringVar(&o.HttpApiAddress, "http-api-address", "0.0.0.0:4131", "qps-reporter http api address")
	pflag.StringVar(&o.Topic, "topic", os.Getenv("NSQ_CLUSTER"), "Topic name")
	pflag.StringVar(&o.InstanceName, "instance-name", os.Getenv("POD_NAME"), "Instance name")
	pflag.StringVar(&o.Namespace, "namespace", os.Getenv("POD_NAMESPACE"), "Pod namespace")
	pflag.DurationVar(&o.HttpRequestTimeout, "http-request-timeout", 2*time.Second, "Http request timeout")
	pflag.DurationVar(&o.UpdatePeriod, "update-period", 10*time.Second, "Message count check/qps update period")
	pflag.IntVar(&o.PreservedQpsCount, "preserved-qps-count", 3, "Preserved qps record count per nsqd instance")

	pflag.BoolVar(&o.DryRun, "dry-run", false, "Print qps update info to stdout instead of updating nsqdscale resource")

	pflag.BoolVar(&o.Version, "version", false, "Print version")
}

func (o *Options) Parse() {
	pflag.Parse()
}
