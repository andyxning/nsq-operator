/*
Copyright 2018 The NSQ-Operator Authors.

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
	"os"
	"time"

	"fmt"

	"github.com/andyxning/nsq-operator/cmd/nsq-operator/options"
	"github.com/andyxning/nsq-operator/pkg/controller"
	"github.com/andyxning/nsq-operator/pkg/signal"
	"github.com/andyxning/nsq-operator/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	nsqclientset "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	nsqinformers "github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
)

func main() {
	opts := options.NewOptions()
	opts.RegisterFlags()
	opts.Parse()

	if opts.Version {
		fmt.Printf("%#v\n", version.Get())
		os.Exit(0)
	}

	registerHttpHandler()
	go startHttpServer(opts.PrometheusAddress)

	stopCh := signal.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(opts.APIServerURL, opts.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}

	cfg.UserAgent = version.BuildUserAgent()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	nsqClient, err := nsqclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building nsq clientset: %v", err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*10)
	nsqInformerFactory := nsqinformers.NewSharedInformerFactory(nsqClient, time.Second*10)

	nsqAdminController := controller.NewNsqAdminController(kubeClient, nsqClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().ConfigMaps(),
		nsqInformerFactory.Nsq().V1alpha1().NsqAdmins())

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	nsqInformerFactory.Start(stopCh)

	if err = nsqAdminController.Run(8, stopCh); err != nil {
		klog.Fatalf("Error running nsqadmin controller: %v", err)
	}

	<-stopCh
}
