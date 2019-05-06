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

package types

import "time"

var (
	nsqdMemoryOverSalePercent = 50

	topicMemoryQueueSize   = 10000
	channelMemoryQueueSize = 10000

	nsqdCommandDataPath               = "/data"
	nsqdCommandMaxBodySize            = 5242880
	nsqdCommandMaxChannelConsumers    = 1024
	nsqdCommandMaxMsgSize             = 1048576
	nsqdCommandMaxOutputBufferTimeout = time.Second
	nsqdCommandSyncEvery              = 2500
	nsqdCommandSyncTimeout            = 2 * time.Second
	nsqdCommandStatsdMemStats         = true
	nsqdCommandStatsdInterval         = 30 * time.Second
	nsqdCommandSnappy                 = true
	nsqdCommandMaxRequeueTimeout      = time.Hour

	waitTimeout = 900 * time.Second
)
