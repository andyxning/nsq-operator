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

    echo "killing nsqd"
    nsqd_pid=`pgrep -P ${main_pid} "^nsqd"`
    /bin/kill -s SIGTERM ${nsqd_pid}
    while true; do
        if [[ ! -e /proc/${nsqd_pid} ]]; then
            break
        fi
        sleep 1
    done
    echo "nsqd killed"
}

_int() {
    echo "receive SIGINT signal"

    echo "killing nsqd"
    nsqd_pid=`pgrep -P ${main_pid} "^nsqd"`
    /bin/kill -s SIGINT ${nsqd_pid}
    while true; do
        if [[ ! -e /proc/${nsqd_pid} ]]; then
            break
        fi
        sleep 1
    done
    echo "nsqd killed"
}

trap _term SIGTERM
trap _int SIGINT

LOG_DIR=${LOG_DIR:-"/var/log/nsqd"}

mkdir -p ${LOG_DIR}

CONF_FILE="/etc/nsq/nsqd"

if [[ -f ${CONF_FILE} ]]; then
  source ${CONF_FILE}
else
  echo "${CONF_FILE} does not exist"
  exit 1
fi

nsqd ${NSQD_COMMAND_ARGUMENTS} --lookupd-tcp-address=${NSQD_LOOKUPD_TCP_ADDRESS} -statsd-address=${NODE_IP}:8125 2>&1 | /usr/local/bin/cronolog_alpine ${LOG_DIR}/log.%Y-%m-%d_%H &

wait