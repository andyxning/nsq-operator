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
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fmt"

	"github.com/andyxning/nsq-operator/cmd/qps-reporter/options"
	"github.com/andyxning/nsq-operator/cmd/qps-reporter/types"
	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	"github.com/andyxning/nsq-operator/pkg/signal"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	nsqclientset "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	"k8s.io/klog"
)

var (
	topicStatsApi = "%s/stats?format=json&topic=%s"

	preMessageCount uint64 = 0
	preUpdateTime          = time.Now()
)

const version string = "1.0.0"

func main() {
	opts := options.NewOptions()
	opts.MustRegisterFlags()
	opts.Parse()

	if opts.Version {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	stopCh := signal.SetupSignalHandler()
	exiting := false

	var cfg *rest.Config
	var err error
	// creates in-cluster config
	if opts.APIServerURL == "" && opts.KubeConfig == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(opts.APIServerURL, opts.KubeConfig)
	}
	if err != nil {
		klog.Fatalf("Building kubeconfig error: %v", err)
	}

	cfg.UserAgent = fmt.Sprintf(
		"%s (%s/%s)", filepath.Base(os.Args[0]), runtime.GOOS, runtime.GOARCH)

	nsqClient, err := nsqclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Building nsq clientset error: %v", err)
	}

	http.Handle("/healthz", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_, err := resp.Write([]byte("OK"))
		if err != nil {
			klog.Errorf("Writer response error: %v", err)
			resp.WriteHeader(http.StatusInternalServerError)
		}
	}))
	go func() {
		klog.Infof("Start http api server on %s", opts.HttpApiAddress)
		err = http.ListenAndServe(opts.HttpApiAddress, nil)
		if err != nil {
			klog.Warningf("Start http api server error: %v", err)
			os.Exit(1)
		}
	}()

	periodTicker := time.NewTicker(opts.UpdatePeriod)
	isFirstStart := true

	for {
		if exiting {
			break
		}

		select {
		case <-periodTicker.C:
			now := time.Now()
			messageCount, err := queryNsqdMessageCount(opts)
			if err != nil {
				klog.Errorf("Ignore updating qps for topic %q", opts.Topic)
				break
			}

			if !isFirstStart {
				timeDiff := now.Sub(preUpdateTime)
				klog.Infof("Topic %q, now: %v, previous update time: %v, time shift: %v", opts.Topic, now, preUpdateTime, timeDiff)
				countDiff := int64(messageCount - preMessageCount)
				klog.Infof("Topic %q, message count: %v, previous message count: %v, message count shift: %v", opts.Topic, messageCount, preMessageCount, countDiff)

				qps := int64(math.Ceil(float64(countDiff) / float64(timeDiff/time.Second)))
				klog.Infof("Topic %q, counted qps: %v", opts.Topic, qps)
				if qps < 0 {
					klog.Warningf("Ignore update qps for topic %q. Message count diff is negative. Maybe nsqd has been restarted",
						opts.Topic)
					goto update
				}

				nsqdQPS := v1alpha1.Qps{
					LastUpdateTime: metav1.NewTime(now),
					Qps:            qps,
				}
				if err := updateNsqdScaleStatus(nsqClient, nsqdQPS, opts); err != nil {
					klog.Errorf("Updating nsqd scale status for %s/%s error. qps: %+v. Error: %v",
						opts.Namespace, opts.InstanceName, nsqdQPS, err)
				} else {
					klog.Infof("Updating nsqd scale status for %s/%s. qps: %+v",
						opts.Namespace, opts.InstanceName, nsqdQPS)
				}
			} else {
				isFirstStart = false
			}

		update:
			klog.V(2).Infof("Update preMessageCount from %d to %d, preUpdateTime(UTC) from %d to %d",
				preMessageCount, messageCount, preUpdateTime.Unix(), now.Unix())
			preMessageCount = messageCount
			preUpdateTime = now
		case <-stopCh:
			exiting = true
			break
		}
	}

	klog.Infof("Shut down qps-reporter")
}

func queryNsqdMessageCount(opts *options.Options) (messageCount uint64, err error) {
	http.DefaultClient.Timeout = opts.HttpRequestTimeout

	var formatter string
	if strings.HasPrefix(opts.NsqdApiAddress, "http://") {
		formatter = topicStatsApi
	} else {
		formatter = fmt.Sprintf("http://%s", topicStatsApi)
	}

	url := fmt.Sprintf(formatter, opts.NsqdApiAddress, opts.Topic)
	klog.V(2).Infof("Nsqd stats api for topic %q: %v", opts.Topic, url)

	resp, err := http.Get(url)
	if err != nil {
		klog.Errorf("Get nsqd stats api for topic %q error: %v", opts.Topic, err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		klog.Errorf("Response from nsqd stats api for topic %q returns status code: %v", opts.Topic, resp.StatusCode)
		return 0, fmt.Errorf("response status code is %v", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("Read nsqd stats api for topic %q error: %v", opts.Topic, err)
		return 0, err
	}

	topicStats := types.TopicStats{}
	err = json.Unmarshal(content, &topicStats)
	if err != nil {
		klog.Errorf("Unmarshal nsqd stats api for topic %q error: %v", opts.Topic, err)
		return 0, err
	}

	if len(topicStats.Topics) < 1 {
		klog.Errorf("Nsqd stats api for topic %q does not exist", opts.Topic)
		return 0, fmt.Errorf("nsqd stats api for topic %q does not exist", opts.Topic)
	}

	return topicStats.Topics[0].MessageCount, nil
}

func updateNsqdScaleStatus(nsqclientset *nsqclientset.Clientset, qps v1alpha1.Qps, opts *options.Options) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if opts.DryRun {
			klog.Infof("Qps for nsqdscale %s/%s is: %+v", opts.Namespace, opts.Topic, qps)
			return nil
		} else {
			nds, err := nsqclientset.NsqV1alpha1().NsqdScales(opts.Namespace).Get(opts.Topic, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					klog.Infof("Can not find nsqdscale for %s/%s. Ignore updates", opts.Namespace, opts.Topic)
					return nil
				} else {
					return fmt.Errorf("failed to get latest version of nsqdscale %s/%s: %v", opts.Namespace, opts.Topic, err)
				}
			}

			newNDS := nds.DeepCopy()
			if newNDS.Status.Qpses == nil {
				newNDS.Status.Qpses = make(map[string][]v1alpha1.Qps)
			}

			newNDS.Status.Qpses[opts.InstanceName] = append(newNDS.Status.Qpses[opts.InstanceName], qps)
			if len(newNDS.Status.Qpses[opts.InstanceName]) > opts.PreservedQpsCount {
				newNDS.Status.Qpses[opts.InstanceName] = newNDS.Status.Qpses[opts.InstanceName][1 : opts.PreservedQpsCount+1]
			}

			_, err = nsqclientset.NsqV1alpha1().NsqdScales(opts.Namespace).Update(newNDS)
			return err
		}
	})

	return err
}
