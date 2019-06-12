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

package controller

import (
	"testing"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
)

func TestComputeNsqdMemoryResource(t *testing.T) {
	cases := []struct {
		Desc                     string
		MessageAvgSize           int32
		MemoryOverBookingPercent int32
		MemoryQueueSize          int32
		ChannelCount             int32
		WantedRequest            string
		WantedLimit              string
	}{
		{
			Desc:                     "topic 1, channel 0",
			MessageAvgSize:           123,
			MemoryQueueSize:          10000,
			MemoryOverBookingPercent: 50,
			ChannelCount:             0,
			WantedRequest:            "2Mi",
			WantedLimit:              "2Mi",
		},
		{
			Desc:                     "topic 1, channel 1",
			MessageAvgSize:           234,
			MemoryQueueSize:          567,
			MemoryOverBookingPercent: 30,
			ChannelCount:             1,
			WantedRequest:            "1Mi",
			WantedLimit:              "1Mi",
		},
		{
			Desc:                     "topic 1, channel 1024(Overflow int32)",
			MessageAvgSize:           1024,
			MemoryQueueSize:          10000,
			MemoryOverBookingPercent: 100,
			ChannelCount:             1024,
			WantedRequest:            "20020Mi",
			WantedLimit:              "20020Mi",
		},
	}

	ndc := NsqdController{}
	for _, ut := range cases {
		nd := &v1alpha1.Nsqd{
			Spec: v1alpha1.NsqdSpec{
				MessageAvgSize:           ut.MessageAvgSize,
				MemoryOverBookingPercent: ut.MemoryOverBookingPercent,
				MemoryQueueSize:          ut.MemoryQueueSize,
				ChannelCount:             ut.ChannelCount,
			},
		}
		memRequest, memLimit := ndc.computeNsqdMemoryResource(nd)
		if memRequest.String() != ut.WantedRequest || memLimit.String() != ut.WantedLimit {
			t.Errorf("Desc: %q. Wanted memory request: %q, got: %q, Wanted memory limit: %q, got: %q",
				ut.Desc, ut.WantedRequest, memRequest.String(), ut.WantedLimit, memLimit.String())
		}
	}
}
