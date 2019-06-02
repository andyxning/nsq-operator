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

package error

import "errors"

// ErrResourceNotOwnedByNsqAdmin is used as part of the Event 'reason' when nsqadmin resource fails
// to sync due to a Deployment/ConfigMap of the same name already existing.
var ErrResourceNotOwnedByNsqAdmin = "ErrResourceNotOwnedByNsqAdmin"

// ErrResourceNotOwnedByNsqLookupd is used as part of the Event 'reason' when nsqlookupd resource fails
// to sync due to a Deployment/ConfigMap of the same name already existing.
var ErrResourceNotOwnedByNsqLookupd = "ErrResourceNotOwnedByNsqLookupd"

// ErrResourceNotOwnedByNsqd is used as part of the Event 'reason' when nsqd resource fails
// to sync due to a StatefulSet/ConfigMap of the same name already existing.
var ErrResourceNotOwnedByNsqd = "ErrResourceNotOwnedByNsqd"

// ErrNoNeedToUpdateNsqdReplica means that no replica update needed for nsqd against nsqdscale
var ErrNoNeedToUpdateNsqdReplica = errors.New("ErrNoNeedToUpdateNsqdReplica")
