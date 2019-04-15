NSQ_COMPONENTS := nsqd nsqlookupd nsqadmin

nsq-images := $(addprefix .image-, ${NSQ_COMPONENTS})
ldflags := $(shell ./hack/version.sh)

PKG_PREFIX := github.com/andyxning/nsq-operator

build:
	go build -ldflags="${ldflags}" -o nsq-operator ${PKG_PREFIX}/cmd/nsq-operator


images: ${nsq-images}
.image-%:
	$(MAKE) --no-print-directory -C images $*

gen-code:
	./hack/update-codegen.sh

update-gen-tool:
	./hack/update-codegen-tool.sh

verify-codegen:
	./hack/verify-codegen.sh

.PHONY: images build gen-code verify-codegen update-gen-tool
