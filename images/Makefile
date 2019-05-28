NSQ_VERSION = 1.1.0
GO_VERSION = 1.10.3

nsqd:
	$(MAKE) --no-print-directory -C ../ qps-reporter
	cp ../qps-reporter .
	docker build --no-cache --build-arg nsq_version=${NSQ_VERSION} --build-arg go_version=${GO_VERSION} -t $@:${NSQ_VERSION} -f $@.Dockerfile .

nsqlookupd:
	docker build --no-cache --build-arg nsq_version=${NSQ_VERSION} --build-arg go_version=${GO_VERSION} -t $@:${NSQ_VERSION} -f $@.Dockerfile .

nsqadmin:
	docker build --no-cache --build-arg nsq_version=${NSQ_VERSION} --build-arg go_version=${GO_VERSION} -t $@:${NSQ_VERSION} -f $@.Dockerfile .

.PHONY: nsqd nsqlookupd nsqadmin
