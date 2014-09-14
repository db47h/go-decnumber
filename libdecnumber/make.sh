#!/bin/sh

# Use this if you don't have a working make
#
# in the CFLAGS below, set DECLITEND to 0 if your architecture is big endian
#
# should be easily convertible to a Windows .bat file.

[ -d obj  ] || mkdir obj # create obj folder if it does not exist
[ -d lib  ] || mkdir lib

CC=gcc # full path works as well, like /usr/local/bin/gcc-4.9
AR=ar  # gnu ar
LD=ld
CFLAGS="-Wall -Werror -std=c99 -DDECPRINT=0 -DDECEXTGLAG=1 -DDECLITEND=1"

set -x # display commands as they are being run

# compile needed C files
"$CC" $CFLAGS -c -o obj/decQuad.o decQuad.c
"$CC" $CFLAGS -c -o obj/decNumber.o decNumber.c
"$CC" $CFLAGS -c -o obj/decContext.o decContext.c

# build static library
"$AR" rcs lib/libdecnumber.a obj/decQuad.o obj/decNumber.o obj/decContext.o

# And syso
# "$LD" -r lib/*.o -o ../libdecnumber.syso
