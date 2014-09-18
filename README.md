# Overview

Package dec is a go wrapper package around the [libDecnumber library][lib].

This is a work in progress. The API is in a more or less final state, all the decContext functions
have been implemented, but most decNumber functions are still missing, and there is no decQuad
implementation yet.

[lib]: http://speleotrove.com/decimal/decnumber.html


# C decNumber to Go

The decNumber package is split into modules: the decContext module (required), the decNumber module
(for arbitrary precision arithmetic), *float* modules, namely decSingle, decDouble and decQuad, and
a few other modules for conversion between various formats. The *float* modules are based on the
32-bit, 64-bit, and 128-bit decimal types in the IEEE 754 Standard for Floating Point Arithmetic. In
contrast to the arbitrary-precision decNumber module, these modules work directly from the
decimal-encoded formats designed by the IEEE 754 committee. Their implementation is also faster: an
Add() with 34 digits numbers takes 433 cycles with de decQuad module versus 1180 for the decNumber
module.

Note that there is no *standard* libdecnumber.so. The decNumber package is provided as-is, with no
makefile to make a shared library out of it (the Makefile in this repository is not part of the
original decNumber archive and is only a test).

From a C programming perspective, all one has to do to compile code for decNumber is:

	gcc -DSOME_TUNING_DEFINE=1234 mysource.c decContext.c decQuad.c decNumber.c -o myprogram

or if only decNumber is needed:

	gcc mysource.c decContext.c decNumber.c -o myprogram

My initial intent was to split the wrapper into modules, in the same way as the C version. However,
this lead to unsolvable compilation issues, like missing references in the Number (Go) module to the
decContext (C) module. This is currently a no-Go (sorry) without shared libraries. However, I did
not want to force the end-user of the package to build and install a custom shared library, and I
also wanted the package to be `go get`'able. Also considering the design decisions discussed in the
next section, I ended up making a monolithic package around decNumber and decContext.

The usual way to link Go code against a static library is to use a `#cgo LDFLAGS:
static/patch/to/lib/lib.a` directive. Since this package is meant to be imported into other
projects, using LDFLAGS was not possible without some standardized/portable/flexible way to specify
the path to the static library (whatever a package puts into LDFLAGS propagates to the project
importing it).

The trick to make it work is to use Go source files as wrappers around the relevant C files and
`#include` them. This works quite well, except for long build times and a weird behaviour of `go
test` if not using a relative import path in test source files. The files decContext.go and
decNumber.go just do this: include the corresponding C file. See the topic about building/installing
for more options.

Most of the short C functions (accessors) have been reimplemented in Go in order to improve
performance. Use

	go build -gcflags=-m . 2>&1 | grep inline

to check which functions can be inlined.


# Numbers, Context and precision

The decNumber module can be built to use fixed precision numbers or arbitrary precision (changeable
at runtime), or a mix of both. In order to make things easier and more flexible for the clients of
this package, the decNumber module is setup for arbitrary precision numbers.

Most functions take a *context* parameter which provides the context for operations (precision,
rounding mode, etc.) and also controls the handling of exceptional conditions. For example, the
Exp() function is defined like this (C version):

	decNumber * decNumberExp(decNumber *res, const decNumber *rhs, decContext *set)

It will set *res* to *e* raised to the power of *rhs*. The *rhs* operand can be in any precision
(i.e. context independent). However, \**res*, the decNumber structure that will hold the result, has
to have enough storage space to hold the precision specified in the decContext *set*.

One of the major problems a programmer may have when working with the decNumber library is keeping
track of which decNumber was created to work with which decContext (i.e. with enough storage space
for that context). We tried several approaches to get rid of this context parameter in the Go
implementation, but with no success. Suggestions welcome!

Note that this is a non-issue for applications using a fixed precision (global Context), while
applications that require dynamic precision can leverage the dec.NumberPool facility to keep track
of their working context and free Number list with a single variable.


# Go implementation details

- Contexts are created with a immutable precision (i.e. number of digits). If one needs to change
  precision on the fly, discard the existing context and create a new one with the required precision.
- From a programming standpoint, any initialized Number is a valid operand in arithmetic operations,
  regardless of the settings or existence of its creator Context (not to be confused with having a
valid value in a given arithmetic operation).
- Arithmetic functions are Number methods. The value of the receiver of the method will be set to
  the result of the operation. For example:

	n.Add(x, y, context) // n = x + y

- Arithmetic methods always return the receiver in order to allow chain calling:

	n.Add(x, y, context).Multiply(n, z, context) // n = (x + y) * z

- Using the same Number as operand and result, like in `n.Multiply(n, n, ctx)`, is legal and will not
  produce unexpected results.

- A few functions like the context status manipulation functions have been moved to their own type
(Status). The C call decContextTestStatus(ctx, mask) is therefore replaced by ctx.Status().Set(mask)
in the Go implementation. The same goes for decNumberClassToString(number) which is repleaced by
number.Class().String() in go.

In the C implementation, decQuads are defined in a supporting module for the decimal128 format and
provide a set of functions that work directly in this format. The decimal128 and decQuad structures
are identical (except in name) so pointers to the structures can safely be cast from one to the
other.  The separation between decQuad and decimal128 in the source code allowed to use the decQuad
module stand-alone (that is, it has no dependency on the decNumber module).

The same goes for decSingle and decimal32, decDouble and decimal64.

In the Go implementation, and even if we could split the wrapper into sub-packages, this separation
does not make much sense since we want to provide access to everything that decNumber has to offer;
the linker will take care of including only the used bits and pieces into the final application
executable. As such, the decimal32/64/128 are merged into Single, Double and Quad.

### Error handling

Active eror handling via traps is not supported in the Go implementation. The os/signal package does
not seem to be able to handle signals raised from C code (this always causes a panic), while
external signals can be handled just fine.

Although most arithmetic functions can cause errors, the standard Go error handling is not used in
its idiomatic form. That is, arithmetic functions do not return errors. Instead, the type of the
error is ORed into the status flags in the current context (Context type). It is the responsibility
of the caller to clear the status flags as required. The result of any routine which returns a
number will always be a valid number (which may be a special value, such as an Infinity or NaN).
This permits the use of much fewer error checks; a single check for a whole computation is often
enough.

To check for errors, get the Context's status with the Status() function (see the Status type), or
use the Context's ErrorStatus() function.


## Free-list of Numbers

The package provides facilities for managing free-lists of Numbers in order to relieve pressure on
the garbage collector in computation intensive applications. NumberPool is in fact a simple wrapper
around a \*Context and a `sync.Pool`, or the lighter `dec.Pool`, which will automatically cast the
return value of `Get()` to the desired type.

For example:

	ctx := dec.NewContext(dec.InitDecimal128, 0)
	// idomatic code for NumberPool creation
	pool := &dec.NumberPool{
		&sync.Pool{
			New: func() interface{} { return dec.NewNumber(ctx.Digits()) },
		},
		ctx,                             // same context as the one used in New()
	}
	number := pool.Get()                 // with no need to type cast to *Number
	defer pool.Put(number)               // idiomatic code for short lived numbers
	number.FromString("1243", pool.Context)

Note the use of `pool.Context` on the last statement.

The provided `dec.Pool` implementation is not thread safe and is only provided as a
lightweight alternative to sync.Pool.

If an application needs to change its arithmetic precision on the fly, any NumberPool built on top
of the affected Context's will need to be discarded and recreated along with the Context. This will
not affect existing numbers that can still be used as valid operands in arithmetic functions.

## Example scenario

A concrete usage example of the fixed context precision and free-list management could be a
calculator application where we have:

- a global NumberPool managing a free-list of Number's and only reference to a global pool.
- a global stack of numbers (implemented as a slice)

For all arithmetic computations, temporary numbers, etc., we use the idiomatic `number := pool.Get()`
followed by a deferred call to `pool.Put(number)`. To compute the addition of the top two
numbers on the stack, we do the following:

	X, Y = := globalStack.Pop2()          // pop top 2 numbers off the stack
	result := globalPool.Get()            // Get a new Number
	result.Add(X, Y, globalPool.Context)  // and set it to X+Y
	globalPool.Put(X)                     // Put X and Y back in the pool
	globalPool.Put(Y)
	globalStack.Push(result)              // push the result on top of the stack

When the user requests a change in precision, we create a new Context setup for the requested
precision and replace the global NumberPool with a new one referencing this new Context.  Numbers
present on the stack are kept as-is since they are still valid Numbers when used as operands in
arithmetic functions. New operations will be performed using the new Context's precision since we
make sure that every operation is done with a freshly created Number for its result.

## Threading, goroutines

The decNumber library is thread safe as long as threads do not share decContext or decNumber
structures. The same rule applies to the Go wrapper package. The provided Pool is not thread safe
either.

A thread safe application could use an immutable global context with a sync.Pool to manage Number
allocation, and share Number's between goroutines by communicating.

## What about decSingle, decDouble, decQuad ?

Right now, the main focus of the dec package is on decNumber. Other modules are only partially
implemented with just enough functionality to be able to run the C decNumber examples. decQuad will
be next.


# Building / Installing

## go get

If you only intend to include this package in your own project, just run

	go get -u github.com/wildservices/go-decnumber

and you're all set.

## .syso technique

Another option for package maintainers is to use the [.syso mechanism][syso] which greatly speeds up
the build process. The idea is to bundle together all the .o files into a single .syso file. When
such a file is present in a package folder, it will automatically be linked with the other object
files.

To make the .syso file:

	cd libdecnumber
	make syso
	cd ..

And benefit:

	go test -tags="syso" -i ./...
	go test -tags="syso "./...

When the `syso` tag is specified in a Go build, the wrapper files (decContext.go and decNumber.go)
are ignored and the build uses the precompiled object file libdecnumber_${GOOS}_${GOARCH}.syso. The
difference with a static library is that we do not have to use a `#cgo LDFLAGS` directive that would
bring in the static library in client projects as well.

The top level Makefile (at the root of the package folder) has the following targets:

	make            # builds using the "normal" wrapper mechanism, removing any syso file beforehand
	make build      # build using the syso file, compiling it if necessary
	make test       # test using the syso file, compiling it if necessary
	make clean      # the usual + removes the syso file.

Running tests using the Go->C wrappers takes 6.7 seconds on my workstation, versus only 1.6 second
when using the syso mechanism.

[syso]: https://code.google.com/p/go-wiki/wiki/GcToolchainTricks#Use_syso_file_to_embed_arbitrary_self-contained_C_code


# TODO

- Implement basic math functions.
- Thoroughly test free-list management and proper resource clean-up.
- merge decimal32/64/128 into Single, Double and Quad.

# Licensing

## go-decnumber (dec package)

The go-decnumber wrapper is:

Copyright 2014 Denis Bernard (wldsvc at gmail.com)

All rights reserved.

Use of this package is governed by a BSD-style license that can be found in the LICENSE file.

## decNumber C library

The decNumber C library is:

Copyright (c) 1995-2010 International Business Machines Corporation and others

All rights reserved.

The decNumber C library is made available under the terms of the ICU License -- ICU 1.8.1 and later,
which can be found in the LICENSE-ICU file.
