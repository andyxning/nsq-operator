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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andyxning/nsq-operator/cmd/reporter/options"
	"github.com/andyxning/nsq-operator/cmd/reporter/types"
	"github.com/nsqio/nsq/nsqd"
)

func TestQueryNsqdMessageCount(t *testing.T) {
	cases := []struct {
		Desc               string
		Handler            http.Handler
		WantedError        bool
		WantedMessageCount uint64
		WantedDepth        int64
	}{
		{
			Desc: "timeout",
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				time.Sleep(2 * time.Second)
			}),
			WantedError:        true,
			WantedMessageCount: 0,
			WantedDepth:        0,
		},
		{
			Desc: "non 200 response status code",
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				resp.WriteHeader(http.StatusInternalServerError)
			}),
			WantedError:        true,
			WantedMessageCount: 0,
			WantedDepth:        0,
		},
		{
			Desc: "invalid json",
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				_, err := resp.Write([]byte("invalid json"))
				if err != nil {
					t.Errorf("Write response error: %v", err)
				}
			}),
			WantedError:        true,
			WantedMessageCount: 0,
			WantedDepth:        0,
		},
		{
			Desc: "topic not found",
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				topicStats := types.TopicStats{
					Version:   "1.0.0",
					Health:    "OK",
					StartTime: 123,
					Topics:    []nsqd.TopicStats{},
				}
				content, err := json.Marshal(topicStats)
				if err != nil {
					t.Errorf("Marshal response error: %v", err)
				}
				_, err = resp.Write(content)
				if err != nil {
					t.Errorf("Write response error: %v", err)
				}
			}),
			WantedError:        true,
			WantedMessageCount: 0,
			WantedDepth:        0,
		},
		{
			Desc: "one topic, two channels",
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				topicStats := types.TopicStats{
					Version:   "1.0.0",
					Health:    "OK",
					StartTime: 123,
					Topics: []nsqd.TopicStats{
						{
							Channels: []nsqd.ChannelStats{
								{
									ChannelName:  "ch1",
									Depth:        1,
									BackendDepth: 1,
								},
								{
									ChannelName:  "ch2",
									Depth:        2,
									BackendDepth: 2,
								},
							},
							TopicName:    "test",
							MessageCount: 100,
							Depth:        100,
							BackendDepth: 100,
						},
					},
				}
				content, err := json.Marshal(topicStats)
				if err != nil {
					t.Errorf("Marshal response error: %v", err)
				}
				_, err = resp.Write(content)
				if err != nil {
					t.Errorf("Write response error: %v", err)
				}
			}),
			WantedError:        false,
			WantedMessageCount: 100,
			WantedDepth:        100 + 100 + 1 + 1 + 2 + 2,
		},
	}

	for _, ut := range cases {
		func() {
			server := httptest.NewServer(ut.Handler)
			defer server.Close()

			opts := options.Options{}
			opts.Topic = "test"
			opts.HttpRequestTimeout = time.Second
			opts.NsqdApiAddress = server.URL

			messageCount, depth, err := queryNsqdStats(&opts)
			if (ut.WantedError && err == nil) || (!ut.WantedError && err != nil) {
				t.Errorf("Desc: %q. Wanted error: %v, got error: %v", ut.Desc, ut.WantedError, err)
			}
			if ut.WantedMessageCount != messageCount {
				t.Errorf("Desc: %q. Wanted messageCount: %v, got: %v", ut.Desc, ut.WantedMessageCount, messageCount)
			}
			if ut.WantedDepth != depth {
				t.Errorf("Desc: %q. Wanted depth: %v, got: %v", ut.Desc, ut.WantedDepth, depth)
			}
		}()
	}
}
