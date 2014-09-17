# customize this
LIB          := libdecnumber
SOURCES      := decQuad.c decimal64.c decimal128.c decPacked.c decNumber.c decContext.c
HEADER_FILES := decNumberLocal.h
# default to little endian. Run make DECLITEND=0 on big endian architectures
DECLITEND    ?= 1
# project specific CFLAGS
LOCALCFLAGS  := -std=c99 -DDECPRINT=0 -DDECEXTGLAG=1 -DDECLITEND=$(DECLITEND)


# shouldn't need to customize anything from this point
CFLAGS ?= -Wall -Werror
GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)

CC := $(shell go env CC)

CFLAGS += $(shell go env GOGCCFLAGS)
CFLAGS += $(LOCALCFLAGS)

OBJDIR  := obj_$(GOOS)_$(GOARCH)
LIBDIR = lib
DIRS := $(OBJDIR) $(LIBDIR)

OBJECTS = $(SOURCES:%.c=$(OBJDIR)/%.o)

SYSOBJECT := ../$(LIB)_$(GOOS)_$(GOARCH).syso
STATICLIB := lib/$(LIB)_$(GOOS)_$(GOARCH).a

all: staticlib syso

.PHONY: all syso staticlib clean

syso: $(SYSOBJECT)

staticlib: $(STATICLIB)

$(STATICLIB): $(OBJECTS) $(LIBDIR)
	$(AR) rcs $@ $^

# The following also works with: $(LD) -r input1.o ... inputN.o -o output.syso
# here we use the GCC frontend instead of ld for easier cross-compiling
$(SYSOBJECT): $(OBJECTS)
	$(CC) $(CFLAGS) -nostdlib -Wl,-r $^ -o $@

$(OBJDIR)/%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

$(DIRS):
	mkdir $@

$(OBJECTS): $(HEADER_FILES) $(OBJDIR)

clean:
	-rm -f obj_*/*.o
	-rm -f lib/*.a
	rm ../$(LIB)_*.syso 2>/dev/null || true
