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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

type Options struct {
	APIServerURL string
	KubeConfig   string

	HttpAddress string

	LeaseID        string
	LeaseName      string
	LeaseNamespace string

	NsqAdminTerminationGracePeriodSeconds   int64
	NsqLookupdTerminationGracePeriodSeconds int64
	NsqdTerminationGracePeriodSeconds       int64

	NsqAdminControllerWorker   int
	NsqLookupdControllerWorker int
	NsqdControllerWorker       int
	NsqdScaleControllerWorker  int

	nsqAdminCPULimit      string
	nsqAdminCPURequest    string
	nsqAdminMemoryLimit   string
	nsqAdminMemoryRequest string

	nsqLookupdCPULimit      string
	nsqLookupdCPURequest    string
	nsqLookupdMemoryLimit   string
	nsqLookupdMemoryRequest string

	qpsReporterCPULimit      string
	qpsReporterCPURequest    string
	qpsReporterMemoryLimit   string
	qpsReporterMemoryRequest string

	nsqdCPULimit           string
	nsqdCPURequest         string
	nsqdPVCStorageResource string

	NsqAdminCPULimitResource      resource.Quantity
	NsqAdminCPURequestResource    resource.Quantity
	NsqAdminMemoryLimitResource   resource.Quantity
	NsqAdminMemoryRequestResource resource.Quantity

	NsqLookupdCPULimitResource      resource.Quantity
	NsqLookupdCPURequestResource    resource.Quantity
	NsqLookupdMemoryLimitResource   resource.Quantity
	NsqLookupdMemoryRequestResource resource.Quantity

	NsqdCPULimitResource   resource.Quantity
	NsqdCPURequestResource resource.Quantity
	NsqdPVCStorageResource resource.Quantity

	QpsReporterCPULimitResource      resource.Quantity
	QpsReporterCPURequestResource    resource.Quantity
	QpsReporterMemoryLimitResource   resource.Quantity
	QpsReporterMemoryRequestResource resource.Quantity

	NsqdScaleValidDuration       time.Duration
	NsqdScaleUpSilenceDuration   time.Duration
	NsqdScaleDownSilenceDuration time.Duration

	NsqdScaleUpdationLRUCacheSize int
	NsqdScaleUpdationLRUCacheTTL  time.Duration

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
		panic("can not extract hostname for lease lock identify")
	}

	pflag.StringVar(&o.APIServerURL, "api-server-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&o.HttpAddress, "http-address", "0.0.0.0:3080", "HTTP api address")
	pflag.StringVar(&o.LeaseID, "lease-id", hostName, "Lease lock identify name for a nsq-operator instance in a HA environment")
	pflag.StringVar(&o.LeaseName, "lease-name", "nsq-operator", "Lease lock resource name")
	pflag.StringVar(&o.LeaseNamespace, "lease-namespace", "default", "Lease lock resource namespace")
	pflag.BoolVar(&o.Version, "version", false, "Print version")

	pflag.IntVar(&o.NsqAdminControllerWorker, "nsqadmin-controller-worker", 8, "Worker number for nsqadmin controller")
	pflag.IntVar(&o.NsqLookupdControllerWorker, "nsqlookupd-controller-worker", 8, "Worker number for nsqlookupd controller")
	pflag.IntVar(&o.NsqdControllerWorker, "nsqd-controller-worker", 8, "Worker number for nsqd controller")
	pflag.IntVar(&o.NsqdScaleControllerWorker, "nsqdscale-controller-worker", 16, "Worker number for nsqdscale controller")

	pflag.Int64Var(&o.NsqAdminTerminationGracePeriodSeconds, "nsqadmin-termination-grace-period-seconds", 60, "Termination grace period seconds for nsqadmin resource object")
	pflag.Int64Var(&o.NsqLookupdTerminationGracePeriodSeconds, "nsqlookupd-termination-grace-period-seconds", 60, "Termination grace period seconds for nsqlookupd resource object")
	pflag.Int64Var(&o.NsqdTerminationGracePeriodSeconds, "nsqd-termination-grace-period-seconds", 300, "Termination grace period seconds for nsqd resource object")

	pflag.StringVar(&o.nsqAdminMemoryLimit, "nsqadmin-mem-limit", "200Mi", "Memory limit resource value for a nsqadmin instance")
	pflag.StringVar(&o.nsqAdminCPULimit, "nsqadmin-cpu-limit", "300m", "CPU limit resource value for a nsqadmin instance")
	pflag.StringVar(&o.nsqAdminMemoryRequest, "nsqadmin-mem-request", "150Mi", "Memory request resource value for a nsqadmin instance")
	pflag.StringVar(&o.nsqAdminCPURequest, "nsqadmin-cpu-request", "250m", "CPU request resource value for a nsqadmin instance")

	pflag.StringVar(&o.nsqLookupdMemoryLimit, "nsqlookupd-mem-limit", "200Mi", "Memory limit resource value for a nsqlookupd instance")
	pflag.StringVar(&o.nsqLookupdCPULimit, "nsqlookupd-cpu-limit", "300m", "CPU limit resource value for a nsqlookupd instance")
	pflag.StringVar(&o.nsqLookupdMemoryRequest, "nsqlookupd-mem-request", "150Mi", "Memory request resource value for a nsqlookupd instance")
	pflag.StringVar(&o.nsqLookupdCPURequest, "nsqlookupd-cpu-request", "250m", "CPU request resource value for a nsqlookupd instance")

	pflag.StringVar(&o.nsqdCPULimit, "nsqd-cpu-limit", "300m", "CPU limit resource value for a nsqd instance")
	pflag.StringVar(&o.nsqdCPURequest, "nsqd-cpu-request", "300m", "CPU request resource value for a nsqd instance")
	pflag.StringVar(&o.nsqdPVCStorageResource, "nsqd-pvc-storage-resource", "256Gi", "Storage resource value for a nsqd instance")

	pflag.StringVar(&o.qpsReporterMemoryLimit, "reporter-mem-limit", "100Mi", "Memory limit resource value for a reporter instance")
	pflag.StringVar(&o.qpsReporterCPULimit, "reporter-cpu-limit", "100m", "CPU limit resource value for a reporter instance")
	pflag.StringVar(&o.qpsReporterMemoryRequest, "reporter-mem-request", "100Mi", "Memory request resource value for a reporter instance")
	pflag.StringVar(&o.qpsReporterCPURequest, "reporter-cpu-request", "100m", "CPU request resource value for a reporter instance")

	pflag.DurationVar(&o.NsqdScaleValidDuration, "nsqd-scale-valid-duration", 80*time.Second, "Time duration during which qps is valid and counted")
	pflag.DurationVar(&o.NsqdScaleUpSilenceDuration, "nsqd-scale-up-silence-duration", 180*time.Second, "Time duration during which second scale up is not allowed")
	pflag.DurationVar(&o.NsqdScaleDownSilenceDuration, "nsqd-scale-down-silence-duration", 180*time.Second, "Time duration during which second scale down is ont allowed")

	pflag.IntVar(&o.NsqdScaleUpdationLRUCacheSize, "nsqd-scale-updation-lru-cache-size", 102400, "Updation LRU cache size when deduplicate nsqdscale updates")
	pflag.DurationVar(&o.NsqdScaleUpdationLRUCacheTTL, "nsqd-scale-updation-lru-cache-ttl", 30*time.Second, "Updation LRU cache ttl when deduplicate nsqdscale updates")
}

func (o *Options) MustParse() {
	pflag.Parse()

	o.NsqAdminCPULimitResource = resource.MustParse(o.nsqAdminCPULimit)
	o.NsqAdminMemoryLimitResource = resource.MustParse(o.nsqAdminMemoryLimit)
	o.NsqAdminCPURequestResource = resource.MustParse(o.nsqAdminCPURequest)
	o.NsqAdminMemoryRequestResource = resource.MustParse(o.nsqAdminMemoryRequest)

	o.NsqLookupdCPULimitResource = resource.MustParse(o.nsqLookupdCPULimit)
	o.NsqLookupdMemoryLimitResource = resource.MustParse(o.nsqLookupdMemoryLimit)
	o.NsqLookupdCPURequestResource = resource.MustParse(o.nsqLookupdCPURequest)
	o.NsqLookupdMemoryRequestResource = resource.MustParse(o.nsqLookupdMemoryRequest)

	o.NsqdCPULimitResource = resource.MustParse(o.nsqdCPULimit)
	o.NsqdCPURequestResource = resource.MustParse(o.nsqdCPURequest)
	o.NsqdPVCStorageResource = resource.MustParse(o.nsqdPVCStorageResource)

	o.QpsReporterCPULimitResource = resource.MustParse(o.qpsReporterCPULimit)
	o.QpsReporterMemoryLimitResource = resource.MustParse(o.qpsReporterMemoryLimit)
	o.QpsReporterCPURequestResource = resource.MustParse(o.qpsReporterCPURequest)
	o.QpsReporterMemoryRequestResource = resource.MustParse(o.qpsReporterMemoryRequest)
}
