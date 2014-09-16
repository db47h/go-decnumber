# Overview

go-decnumber is a go wrapper package around the [libDecnumber library][lib].

This is a work in progress. The API is in a more or less final state, all the decContext functions
have been implemented, but most decNumber functions are still missing, and there is no decQuad
implementation yet.

[lib]: http://speleotrove.com/decimal/decnumber.html


# Implementation details

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
also wanted the packege to be `go get`'able. Also considering the design decisions discussed in the
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

## Numbers, Context and precision

The decNumber module can be built to use fixed precision numbers or arbitrary precision (changeable
at runtime), or a mix of both. In order to make things easier and more flexible for the clients of
this package, the decNumber module is setup for arbitrary precision numbers.

The precision is held in a decContext structure and numbers are held in a decNumber structure. The
caveat is that when dealing with arbitrary precision, the decNumber structures do not keep track of
how many digits they can hold. It's up to the programmer to keep track of which decNumber structure
was created to be used in a given context.

A concrete example, the function Exp() is defined like this:

	decNumber * decNumberExp(decNumber *res, const decNumber *rhs, decContext *set)

It will set *res* to *e* raised to the power of *rhs*. The *rhs* operand can be in any precision
(i.e. context independent). However, \**res*, the decNumber structure that will hold the result, has
to have enough storage space to hold the precision specified in the decContext *set*.

In a top-down functional programming model, this is not a serious issue. However, with goroutines
flying all over the place, this can get messy. This lead to a few design choices in the go
implementation. I also tried to make the API as Go-like as I could:

- Contexts are created with a immutable precision (i.e. number of digits). If one needs to change
  precision on the fly, discard the existing context and create a new one with the required precision.
- Numbers are created by a method of Context.
- From a programming standpoint, any Number is a valid operand in arithmetic operations, regardless
  of the settings or existence of its creator Context (not to be confused with having a valid value
  in a given arithmetic operation).
- Numbers hold a pointer to the Context that created them.
- Arithmetic functions are Number methods. The value of the receiver of the method will be set to
  the result of the operation. For example:

	n.Add(x, y) // n = x + y

- Arithmetic methods always return the receiver in order to allow chain calling:

	n.Add(x, y).Multiply(n, z) // n = (x + y) * z

- Using the same Number as operand and result, like in `n.Multiply(n, n)`, is legal and will not
  produce unexpected results.

## free-list of Numbers

The package provides facilities for managing free-lists of numbers in order to relieve pressure on
the garbage collector in computation intensive applications. This is in fact a simple wrapper around
`sync.Pool`, or the lightweight `decnumber.Pool`, which will automatically cast the return value of
`Get()` to the desired type.

For example:

	ctx := decnumber.NewContext(decnumber.InitDecimal128, 0)
	pool := decnumber.NumberPool(&sync.Pool{
		New: func() interface{} { return decnumber.NewNumber(ctx) },
	})
	number := pool.Get()    // with no need to type cast to *Number
	defer pool.Put(number)  // idiomatic code for short lived numbers

The provided `decnumber.Pool` implementation is not thread safe and is only provided as a
lightweight alternative to sync.Pool.

Note that for pooled numbers, and numbers with a pending deferred `Put()`, there is a dependency
Pool -> Number -> Context, in this order. This means that if an application needs to change its
arithmetic precision on the fly, any Pool built on top of the affected Context's will need to be
discarded and recreated along with the Context. This will not affect existing numbers that can still
be used as valid operands in arithmetic functions.

## example use

A concrete usage example of the fixed context precision and free-lists could be a calculator
application where we have:

- a global context
- a global stack of numbers (implemented as a slice)

For all arithmetic computations, temporary numbers, etc., we use the idiomatic deferred call to
Release(). When computing the addition of the top two number, the

	X, Y = := globalStack.Pop2()               // pop top 2 numbers off the stack
	result := globalPool.Get().Add(X, Y)
	globalPool.Put(X)                          // send X and Y back to their creator
	globalPool.Put(Y)
	globalStack.Push(result)                   // push the result on top of the stack

When the user requests a change in precision, the global Context is replaced by a newly created one
with the requested precision and the global Pool is replaced by a new one built on top of the new
Context.  Numbers present on the stack are kept as-is since they are still valid Numbers when used
as operands in arithmetic functions. New operations will be performed using the new context
precision since we make sure that every operation is done with a freshly created Number for its
result.

## Threading, goroutines

The decNumber library is thread safe as long as threads do not share decContext or decNumber
structures. The same rule applies to the Go wrapper package. The provided Pool is not thread safe
either.

A thread safe application could use an immutable global context with a sync.Pool to manage Number
allocation, and share Number's between goroutines by communicating.

## What about decDouble, decQuad ?

Right now, go-decnumber only supports decNumber. Adding support for any of the *float* modules would
require:

- Adding the relevant type in the decnumber module
- Adding the relevant methods to the Context type
- Free list management is less necessary for the float types since their size is static (128 bits
  for Quad), unlike Number which has a variable size (depending on the Context's precision) and
require malloc/feee calls. Given their small size, Quad's that are not used outside of a function's
body should be allocated on the stack (to be tested).


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

In the early days of the package, running tests using the Go->C wrappers (for only 2 C files) took 3
seconds on my workstation, versus only 1 second when using the syso mechanism.

[syso]: https://code.google.com/p/go-wiki/wiki/GcToolchainTricks#Use_syso_file_to_embed_arbitrary_self-contained_C_code


# TODO

- Implement basic math functions.
- Thoroughly test free-list management and resource clean-up.


# Licensing

	go-decnumber
	Copyright 2014 Denis Bernard (wldsvc at gmail.com). All rights reserved.

Use of this package is governed by a BSD-style license that can be found in the LICENSE file.

The decNumber library is made available under the terms of the ICU License -- ICU 1.8.1 and later,
which can be found in the LICENSE-ICU file.
