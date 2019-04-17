FROM alpine:3.9

LABEL maintainer="Andy Xie <andy.xning@gmail.com>"

RUN apk update && apk add --no-cache bash bash-doc bash-completion

COPY ./nsq-operator /usr/local/bin/

ENTRYPOINT ["nsq-operator", "-v=4"]