# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

.PHONY: all
all: bin/serve-package

bin/serve-package:
	$(Q)go build -o $@ cmd/serve-package/main.go

.PHONY: clean
clean:
	$(Q)rm -rf bin

.PHONY: vendor
vendor:
	$(Q)glide update --strip-vendor
	$(Q)glide-vc --use-lock-file --no-tests --only-code
