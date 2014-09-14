TAGS := 
SYSOTAGS := $(TAGS) syso
LIB := libdecnumber
SYSO := $(LIB).syso
TARGET := $(GOPATH)/pkg/linux_amd64/github.com/wildservices/go-decnumber.a

all: build-standalone

.PHONY: all build-standalone build test install clean

build-standalone:
	-@[ ! -e "$(SYSO)" ] || rm "$(SYSO)"
	go build -tags="$(TAGS)" ./...

$(SYSO): 
	$(MAKE) -C "$(LIB)" syso

build: $(SYSO)
	go build -tags="$(SYSOTAGS)" ./...

test: $(SYSO)
	go test -tags="$(SYSOTAGS)" -i ./...
	go test -tags="$(SYSOTAGS)" ./...

install: $(SYSO)
	go install -tags="$(SYSOTAGS)"

clean:
	$(MAKE) -C "$(LIB)" clean
	-[ ! -e "$(TARGET)" ] || rm "$(TARGET)"

