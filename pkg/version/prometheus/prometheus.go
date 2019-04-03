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

// Package prometheus registers nsq-operator version information as
// Prometheus metrics.
package prometheus

import (
	"github.com/andyxning/nsq-operator/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	versionInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nsq_operator_version_info",
			Help: "A metric with a constant '1' value labeled by major, minor, git version, git commit, git tree state, build date, Go version, compiler from which NSQ-Operator was built, and platform on which it is running.",
		},
		[]string{"major", "minor", "gitVersion", "gitCommit", "gitTreeState", "buildDate", "goVersion", "compiler", "platform"},
	)

	versionInfo := version.Get()
	versionInfoMetric.WithLabelValues(
		versionInfo.Major, versionInfo.Minor, versionInfo.GitVersion,
		versionInfo.GitCommit, versionInfo.GitTreeState, versionInfo.BuildDate,
		versionInfo.GoVersion, versionInfo.Compiler, versionInfo.Platform).Set(1)

	prometheus.MustRegister(versionInfoMetric)
}
