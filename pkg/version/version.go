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

// Package version provides version information collected at build time
// to nsq-operator.
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	apimachinaryversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/kubernetes/pkg/version"
)

func Get() apimachinaryversion.Info {
	return version.Get()
}

func BuildUserAgent() string {
	return fmt.Sprintf(
		"%s/%s (%s/%s)", filepath.Base(os.Args[0]), Get().GitVersion, runtime.GOOS, runtime.GOARCH)
}
