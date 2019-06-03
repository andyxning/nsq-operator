FROM alpine:3.9

LABEL maintainer="Andy Xie <andy.xning@gmail.com>"

ARG nsq_version
ARG go_version

ENV NSQ_VERSION=${nsq_version}
ENV GO_VERSION=${go_version}

RUN wget -q https://github.com/nsqio/nsq/releases/download/v${NSQ_VERSION}/nsq-${NSQ_VERSION}.linux-amd64.go${GO_VERSION}.tar.gz -O - \
| tar -C /usr/local/bin --strip-components=2 -zxv -f -
RUN apk update && apk add --no-cache bash bash-doc bash-completion curl

COPY ./scripts /usr/local/bin/
COPY ./bin /usr/local/bin/
COPY ./qps-reporter /usr/local/bin/

ENTRYPOINT ["start_nsqd.sh"]