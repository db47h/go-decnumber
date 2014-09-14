OBJS := obj/decQuad.o obj/decNumber.o obj/decContext.o
HEADER_FILES := decNumberLocal.h
STATICLIB := lib/libdecnumber.a
# default to little endian. Run make DECLITEND=0 on big endian architectures
DECLITEND ?= 1
CFLAGS += -Wall -Werror -std=c99 -DDECPRINT=0 -DDECEXTGLAG=1 -DDECLITEND=$(DECLITEND)
CC ?= gcc
DIRS := obj lib
# TODO: dynamically check the arch, crosscompile
SYSO := ../libdecnumber.syso

all: $(DIRS) $(STATICLIB)

.PHONY: all syso clean

syso: $(SYSO)

$(STATICLIB): $(OBJS)
	$(AR) rcs $@ $^

$(SYSO): $(OBJS)
	$(LD) -r $^ -o $@

obj/%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

$(DIRS):
	mkdir $@

$(OBJS) : $(HEADER_FILES)

clean:
	-rm -f obj/*.o
	-rm -f lib/*.a
	-[ ! -e $(SYSO) ] || rm $(SYSO)
