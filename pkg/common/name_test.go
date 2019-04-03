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
	"testing"
)

func TestNsqAdminDeploymentName(t *testing.T) {
	cases := []struct {
		Desc   string
		Input  string
		Wanted string
	}{
		{
			Desc:   "normal",
			Input:  "test",
			Wanted: "test",
		},
		{
			Desc:   "empty",
			Input:  "",
			Wanted: "",
		},
	}

	for _, ut := range cases {
		ret := NsqAdminDeploymentName(ut.Input)
		if ret != ut.Wanted {
			t.Fatalf("Desc: %v. NsqAdmin Deployment name mismatches for %q. Return: %v. Wanted: %v",
				ut.Desc, ut.Input, ret, ut.Wanted)
		}
	}
}

func TestNsqAdminConfigMapName(t *testing.T) {
	cases := []struct {
		Desc   string
		Input  string
		Wanted string
	}{
		{
			Desc:   "normal",
			Input:  "test",
			Wanted: "test",
		},
		{
			Desc:   "empty",
			Input:  "",
			Wanted: "",
		},
	}

	for _, ut := range cases {
		ret := NsqAdminConfigMapName(ut.Input)
		if ret != ut.Wanted {
			t.Fatalf("Desc: %v. NsqAdmin ConfigMap name mismatches for %q. Return: %v. Wanted: %v",
				ut.Desc, ut.Input, ret, ut.Wanted)
		}
	}
}

func TestNsqLookupdStatefulSetName(t *testing.T) {
	cases := []struct {
		Desc   string
		Input  string
		Wanted string
	}{
		{
			Desc:   "normal",
			Input:  "test",
			Wanted: "test",
		},
		{
			Desc:   "empty",
			Input:  "",
			Wanted: "",
		},
	}

	for _, ut := range cases {
		ret := NsqLookupdStatefulSetName(ut.Input)
		if ret != ut.Wanted {
			t.Fatalf("Desc: %v. NsqLookupd StatefulSet name mismatches for %q. Return: %v. Wanted: %v",
				ut.Desc, ut.Input, ret, ut.Wanted)
		}
	}
}

func TestNsqLookupdConfigMapName(t *testing.T) {
	cases := []struct {
		Desc   string
		Input  string
		Wanted string
	}{
		{
			Desc:   "normal",
			Input:  "test",
			Wanted: "test",
		},
		{
			Desc:   "empty",
			Input:  "",
			Wanted: "",
		},
	}

	for _, ut := range cases {
		ret := NsqLookupdConfigMapName(ut.Input)
		if ret != ut.Wanted {
			t.Fatalf("Desc: %v. NsqLookupd ConfigMap name mismatches for %q. Return: %v. Wanted: %v",
				ut.Desc, ut.Input, ret, ut.Wanted)
		}
	}
}

func TestNsqdStatefulSetName(t *testing.T) {
	cases := []struct {
		Desc   string
		Input  string
		Wanted string
	}{
		{
			Desc:   "normal",
			Input:  "test",
			Wanted: "test",
		},
		{
			Desc:   "empty",
			Input:  "",
			Wanted: "",
		},
	}

	for _, ut := range cases {
		ret := NsqdStatefulSetName(ut.Input)
		if ret != ut.Wanted {
			t.Fatalf("Desc: %v. Nsqd StatefulSet name mismatches for %q. Return: %v. Wanted: %v",
				ut.Desc, ut.Input, ret, ut.Wanted)
		}
	}
}
