#!/usr/bin/env bash

# Copyright 2018 The NSQ-Operator Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

main_pid=$$

_term() {
    echo "receive SIGTERM signal"

    echo "killing nsqlookupd"
    nsqlookupd_pid=`pgrep -P ${main_pid} "^nsqlookupd"`
    /bin/kill -s SIGTERM ${nsqlookupd_pid}
    while true; do
        if [[ ! -e /proc/${nsqlookupd_pid} ]]; then
            break
        fi
        sleep 1
    done
    echo "nsqlookupd killed"
}

_int() {
    echo "receive SIGINT signal"

    echo "killing nsqlookupd"
    nsqlookupd_pid=`pgrep -P ${main_pid} "^nsqlookupd"`
    /bin/kill -s SIGINT ${nsqlookupd_pid}
    while true; do
        if [[ ! -e /proc/${nsqlookupd_pid} ]]; then
            break
        fi
        sleep 1
    done
    echo "nsqlookupd killed"
}

trap _term SIGTERM
trap _int SIGINT

LOG_DIR=${LOG_DIR:-"/var/log/nsqlookupd"}

mkdir -p ${LOG_DIR}

CONF_FILE="/etc/nsq/nsqlookupd"

if [[ -f ${CONF_FILE} ]]; then
  source ${CONF_FILE}
else
  echo "${CONF_FILE} does not exist"
  exit 1
fi

nsqlookupd ${NSQLOOKUPD_COMMAND_ARGUMENTS} 2>&1 | /usr/local/bin/cronolog_alpine ${LOG_DIR}/log.%Y-%m-%d_%H &

wait