#!/usr/bin/env bash

# Copyright 2019 The NSQ-Operator Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# https://people.seas.harvard.edu/~apw/stress/
STRESS_VERSION=1.0.4

apk update && apk add --no-cache bash bash-doc bash-completion g++ make curl && \
 curl -o /tmp/stress-${STRESS_VERSION}.tar.gz https://people.seas.harvard.edu/~apw/stress/stress-${STRESS_VERSION}.tar.gz && \
 cd /tmp && tar zxvf stress-${STRESS_VERSION}.tar.gz && cd /tmp/stress-${STRESS_VERSION} && \
 ./configure && make && make install