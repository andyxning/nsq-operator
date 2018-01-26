FROM alpine:3.7

LABEL maintainer="Andy Xie <andy.xning@gmail.com>"

ARG nsq_version=0.3.8
ARG go_version=1.6.2

ENV NSQ_VERSION=${nsq_version}
ENV GO_VERSION=${go_version}

RUN wget https://github.com/nsqio/nsq/releases/download/v${NSQ_VERSION}/nsq-${NSQ_VERSION}.linux-amd64.go${GO_VERSION}.tar.gz -O - \
    | tar -C /usr/bin --strip-components=2 -zxv -f -

CMD sh