ARG GOLANG_VERSION=latest
FROM golang:${GOLANG_VERSION}

LABEL maintainer="Andy Xie <andy.xning@gmail.com>"

COPY ./nsq-operator /usr/local/bin/

ENTRYPOINT ["nsq-operator", "-v=4"]