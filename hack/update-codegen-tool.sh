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

CUR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TOOL_PKG_GIT_ADDRESS="git@github.com:kubernetes/code-generator.git"
TOOL_PKG="k8s.io/code-generator"
VENDOR_TOOL_PKG_PATH="${CUR_DIR}/../vendor/${TOOL_PKG}"
TMP_DIR_PATTERN="code-generator-XXXXX"


# Preliminary
if [[ ! -d "vendor" ]]; then
  echo "vendor directory loses. Please first gen the vendor directory."
  exit 1
fi

tmp_dir=$(mktemp -d ${TMPDIR:-/tmp}/${TMP_DIR_PATTERN})
pushd ${tmp_dir}
git clone "${TOOL_PKG_GIT_ADDRESS}"
if [[ ! -d "${VENDOR_TOOL_PKG_PATH}" ]]; then
  mkdir -p "${VENDOR_TOOL_PKG_PATH}"
fi
cp -r code-generator/* "${VENDOR_TOOL_PKG_PATH}"
popd
rm -rf ${tmp_dir}