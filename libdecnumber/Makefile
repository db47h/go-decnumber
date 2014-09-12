SRC = obj/decQuad.c obj/decNumber.c obj/decContext.c
HEADER_FILES = decNumberLocal.h
STATICLIB = lib/libdecnumber.a
CFLAGS = -Wall
CC = gcc
DIRS = obj lib

all: $(DIRS) $(STATICLIB)

$(STATICLIB): $(SRC:.c=.o)
	ar rcs $@ $^

obj/%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

$(DIRS):
	mkdir $@

$(OBJS) : $(HEADER_FILES)

clean:
	-rm -f obj/*.o
	-rm -f lib/*.a

