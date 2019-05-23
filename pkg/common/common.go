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
	"fmt"
	"strings"
)

func AssembleNsqLookupdAddresses(nsqLookupdAddressed []string) string {
	return strings.Join(nsqLookupdAddressed, ",")
}

func NsqdLogMountPath(cluster string) string {
	return fmt.Sprintf("/var/log/%s/", cluster)
}

func NsqLookupdLogMountPath(cluster string) string {
	return fmt.Sprintf("/var/log/%s/", cluster)
}

func NsqAdminLogMountPath(cluster string) string {
	return fmt.Sprintf("/var/log/%s/", cluster)
}

func QpsReporterLogMountPath(cluster string) string {
	return fmt.Sprintf("/var/log/%s/", cluster)
}
