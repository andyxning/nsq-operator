NSQ_COMPONENTS := nsqd nsqlookupd nsqadmin

nsq-images := $(addprefix .image-, ${NSQ_COMPONENTS})

images: ${nsq-images}
.image-%:
	$(MAKE) --no-print-directory -C images $*

.PHONY: images
