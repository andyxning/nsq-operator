NSQ_COMPONENTS := nsqd nsqlookupd nsqadmin

nsq-images := $(addprefix .image-, ${NSQ_COMPONENTS})
ldflags := $(shell ./hack/version.sh)

build:
	go build -ldflags="${ldflags}" -o nsq-operator cmd/nsq-operator/nsq-operator.go


images: ${nsq-images}
.image-%:
	$(MAKE) --no-print-directory -C images $*

.PHONY: images build
