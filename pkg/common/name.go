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

import "fmt"

func NsqAdminDeploymentName(cluster string) string {
	return fmt.Sprintf("%s-nsqadmin", cluster)
}

func NsqAdminConfigMapName(cluster string) string {
	return fmt.Sprintf("%s-nsqadmin", cluster)
}

func NsqLookupdDeploymentName(cluster string) string {
	return fmt.Sprintf("%s-nsqlookupd", cluster)
}

func NsqLookupdConfigMapName(cluster string) string {
	return fmt.Sprintf("%s-nsqlookupd", cluster)
}

func NsqdStatefulSetName(cluster string) string {
	return fmt.Sprintf("%s-nsqd", cluster)
}

func NsqdConfigMapName(cluster string) string {
	return fmt.Sprintf("%s-nsqd", cluster)
}

func NsqdVolumeClaimTemplatesName(cluster string) string {
	return fmt.Sprintf("%s-pvc", cluster)
}
