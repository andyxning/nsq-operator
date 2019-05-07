NSQ_COMPONENTS := nsqd nsqlookupd nsqadmin
NSQ_OPERATOR_VERSION := 0.3.0
GOLANG_VERSION := 1.12.4

nsq-images := $(addprefix .image-, ${NSQ_COMPONENTS})
ldflags := $(shell ./hack/version.sh)

PKG_PREFIX := github.com/andyxning/nsq-operator

vet:
	go list ./... | grep -v "pkg/sdk/examples" | xargs go vet

fmt:
	find . -type f -name "*.go" | grep -v "./vendor/*" | xargs gofmt -s -w -l

test:
	go test -timeout=1m -v -race $(shell go list ./... | grep -v "pkg/sdk/examples")

build: clean
	go build -ldflags="${ldflags}" -o nsq-operator ${PKG_PREFIX}/cmd/nsq-operator

images: ${nsq-images}
.image-%:
	$(MAKE) --no-print-directory -C images $*

nsq-operator-image:
	docker run --rm -v "$(shell pwd)":/go/src/${PKG_PREFIX} -w /go/src/${PKG_PREFIX} golang:${GOLANG_VERSION} make build
	docker build --no-cache --build-arg GOLANG_VERSION=${GOLANG_VERSION} -t nsq-operator:${NSQ_OPERATOR_VERSION} -f Dockerfile .

gen-code:
	./hack/update-codegen.sh

update-gen-tool:
	./hack/update-codegen-tool.sh

verify-codegen:
	./hack/verify-codegen.sh

clean:
	rm -f nsq-operator

.PHONY: clean test fmt vet images build gen-code verify-codegen update-gen-tool nsq-operator-image
