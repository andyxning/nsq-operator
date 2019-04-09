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
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"fmt"

	"github.com/andyxning/nsq-operator/cmd/nsq-operator/options"
	"github.com/andyxning/nsq-operator/pkg/controller"
	"github.com/andyxning/nsq-operator/pkg/signal"
	"github.com/andyxning/nsq-operator/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/transport"

	nsqclientset "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	nsqinformers "github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
)

func main() {
	opts := options.NewOptions()
	opts.MustRegisterFlags()
	opts.Parse()

	if opts.Version {
		fmt.Printf("%#v\n", version.Get())
		os.Exit(0)
	}

	registerHttpHandler()
	go startHttpServer(opts.PrometheusAddress)

	stopCh := signal.SetupSignalHandler()

	var cfg *rest.Config
	var err error
	// creates in-cluster config
	if opts.APIServerURL == "" && opts.KubeConfig == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(opts.APIServerURL, opts.KubeConfig)
	}
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// use a client that will stop allowing new requests once the context ends
	cfg.Wrap(transport.ContextCanceller(ctx, fmt.Errorf("the leader is shutting down")))

	cfg.UserAgent = version.BuildUserAgent()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      opts.LeaseName,
			Namespace: opts.LeaseNamespace,
		},
		Client: kubeClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: opts.LeaseID,
		},
	}

	// start the leader election code loop
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   30 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.Infof("%s: leading", opts.LeaseID)

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

				nsqLookupdController := controller.NewNsqLookupdController(kubeClient, nsqClient,
					kubeInformerFactory.Apps().V1().StatefulSets(),
					kubeInformerFactory.Core().V1().ConfigMaps(),
					nsqInformerFactory.Nsq().V1alpha1().NsqLookupds())

				// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
				// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
				kubeInformerFactory.Start(stopCh)
				nsqInformerFactory.Start(stopCh)

				wg := sync.WaitGroup{}
				wg.Add(2)

				go func() {
					defer wg.Done()

					if err = nsqAdminController.Run(8, stopCh); err != nil {
						klog.Fatalf("Error running nsqadmin controller: %v", err)
					}
				}()

				go func() {
					defer wg.Done()

					if err = nsqLookupdController.Run(8, stopCh); err != nil {
						klog.Fatalf("Error running nsqlookupd controller: %v", err)
					}
				}()

				wg.Wait()
				cancel()

				klog.Infof("Shut down nsq-operator")
			},
			OnStoppedLeading: func() {
				// we can do cleanup here, or after the RunOrDie method
				// returns
				klog.Infof("%s: lost", opts.LeaseID)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == opts.LeaseID {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %v", identity)
			},
		},
	})

	// because the context is closed, the client should report errors
	_, err = kubeClient.CoordinationV1().Leases(opts.LeaseNamespace).Get(opts.LeaseName, metav1.GetOptions{})
	if err == nil || !strings.Contains(err.Error(), "the leader is shutting down") {
		klog.Fatalf("%s: expected to get an error when trying to make a client call: %v", opts.LeaseID, err)
	}

	// we no longer hold the lease, so perform any cleanup and then
	// exit
	klog.Infof("%s: done", opts.LeaseID)
}
