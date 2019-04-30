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

package common

import (
	"flag"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	nsqclientset "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	"github.com/spf13/pflag"
)

var apiServerURL string
var kubeConfig string

func RegisterFlags() {
	// register klog related flags
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.StringVar(&apiServerURL, "api-server-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster")
	pflag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
}

func Parse() {
	pflag.Parse()
}

func InitClients() (kubeClient *kubernetes.Clientset, clientset *nsqclientset.Clientset, err error) {
	var cfg *rest.Config
	// creates in-cluster config
	if apiServerURL == "" && kubeConfig == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(apiServerURL, kubeConfig)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("building kubeconfig: %v", err)
	}

	kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("building kubernetes clientset: %v", err)
	}

	nsqClient, err := nsqclientset.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("building nsq clientset: %v", err)
	}

	return kubeClient, nsqClient, err
}
