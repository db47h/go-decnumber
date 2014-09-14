LIB := libdecnumber

GOARCH ?= $(shell go env | awk -F'["]' '/GOARCH=/{print $$2}')
GOOS ?= $(shell go env | awk -F'["]' '/GOOS=/{print $$2}')

GOTAGS += syso
SYSOBJECT := $(LIB)_$(GOOS)_$(GOARCH).syso

all: build-nosyso

.PHONY: all build-nosyso build test install clean syso

build-nosyso:
	-@[ ! -e "$(SYSOBJECT)" ] || rm "$(SYSOBJECT)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags="$(TAGS)" ./...

syso:	$(SYSOBJECT)

$(SYSOBJECT): 
	$(MAKE) -C "$(LIB)" syso

build: $(SYSOBJECT)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags="$(GOTAGS)" ./...

# why does test need CGO_ENABLED set to work when cross compiling?
test: $(SYSOBJECT)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 go test -tags="$(GOTAGS)" -i ./...
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 go test -tags="$(GOTAGS)" ./...

install: $(SYSOBJECT)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go install -tags="$(GOTAGS)"

clean:
	$(MAKE) -C "$(LIB)" clean

