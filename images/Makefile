NSQ_VERSION = 0.3.8
GO_VERSION = 1.6.2

%:
	docker build --no-cache --build-arg nsq_version=${NSQ_VERSION} --build-arg go_version=${GO_VERSION} -t $*:${NSQ_VERSION} -f $*.Dockerfile .
