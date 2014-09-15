# Overview

go-decnumber is a go wrapper package around the [libDecnumber library][lib].

This is a work in progress. The API is in a more or less final state, all the decContext functions
have been implemented, but most decNumber functions are still missing, and there is no decQuad implementation yet.

[lib]: http://speleotrove.com/decimal/decnumber.html

# Implementation details

The decNumber package is split into modules: the decContext module (required), the decNumber module (for arbitrary precision arithmetic), *float* modules, namely decSingle, decDouble and decQuad, and a few other modules for conversion between various formats. The *float* modules are based on the 32-bit, 64-bit, and 128-bit decimal types in the IEEE 754 Standard for Floating Point Arithmetic. In contrast to the arbitrary-precision decNumber module, these modules work directly from the decimal-encoded formats designed by the IEEE 754 committee. Their implementation is also faster: an Add() with 34 digits numbers takes 433 cycles with de decQuad module versus 1180 for the decNumber module.

Note that there is no *standard* libdecnumber.so. The decNumber package is provided as-is, with no makefile to make a shared library out of it (the Makefile in this repository is not part of the original decNumber archive and is only a test).

From a C programming perspective, all one has to do to compile code for decNumber is:

	gcc -DSOME_TUNING_DEFINE=1234 mysource.c decContext.c decQuad.c decNumber.c -o myprogram

or if only decNumber is needed:

	gcc mysource.c decContext.c decNumber.c -o myprogram

My initial intent was to split the wrapper into modules, in the same way as the C code. However, this lead to unsolvable compilation issues, like missing references in the Number module to the Context module. This is a no-Go (sorry) without shared libraries. However, I did not want to force the end-user of the package to build and install a custom shared library, and considering the design decisions discussed in the next section, I ended up making a monolithic package around decNumber and decContext.

The usual way to link Go code against a static library is to use a `#cgo LDFLAGS: static/patch/to/lib/lib.a` directive. Since this package is meant to be imported into other projects, using LDFLAGS was not possible without some standardized/portable/flexible way to specify the path to the static library (whatever a package puts into LDFLAGS propagates to the project importing it). I also wanted the package to be `go get`able.

The trick to make it work is to use Go source files as wrappers around the relevant C files and `#include` them. This works quite well, except for long build times and a weird behaviour of `go test` if not using a relative import path in test source files. The files decContext.go and decNumber.go just do this: include the corresponding C file. See the topic about building/installing for more options.

Most of the short C functions (accessors) have been reimplemented in Go in order to improve performance. Use

	go build -gcflags=-m . 2>&1 | grep inline

to check which functions can be inlined.

## Numbers, Context and precision

The decNumber module can be built to use fixed precision numbers or arbitrary precision (changeable at runtime), or a mix of both. In order to make things easier and more flexible for the clients of the package, the decNumber module is setup for arbitrary precision numbers.

The precision is held in a decContext structure and numbers are held in a decNumber structure. The caveat is that when dealing with arbitrary precision, the decNumber structures do not keep track of how many digits they can hold. It's up to the programmer to keep track of which decNumber structure was created to be used in a given context.

A concrete example, the function Exp() is defined like this:

	decNumber * decNumberExp(decNumber *res, const decNumber *rhs, decContext *set)

It will set *res* to *e* raised to the power of *rhs*. The *rhs* operand can be in any precision (i.e. context independent). However, *\*res*, the decNumber structure that will hold the result, has to have enough storage space to hold the precision specified in the decContext *set*.

In a top-down functional programming model, this is not a serious issue. However, with goroutines flying all over the place, this can get messy. This lead to a few design choices in the go implementation, and try to make the API as Go-like as I could:

- The configuration of a Context is immutable after creation (i.e. cannot change the number of digits, minimum and maximum exponents). Only rounding and status are alterable. If one needs to change precision on the fly, discard the existing context and create a new one with the required precision. Existing Numbers are still usable and valid Numbers.
- Contexts hold a free list of Numbers and Numbers are created by a method of Context. This gives the following idiomatic code for temporary Number creation:

	num := ctx.NewNumber()
	defer ctx.FreeNumber(num)

- Freeing numbers is not mandatory. The FreeNumber method() only returns it to the free list. When a Number is not returned to the free-list and goes out of scope, it will be handled normally by the garbage collector and any allocated memory freed via an internal SetFinalizer() call. However:
  - A Number must not be used by the caller after calling FreeNumber()
  - FreeNumber() must be called on the Context that created it. If for some reason keeping track of this is not possible, just don't call FreeNumber().
- Arithmetic functions are Context methods. Aside from the Context parameter, we use the same calling convention than in decNumber. That is:

	result = ctx.NumberAdd(result, x, y) // result = x + y

Regarding this last point, my initial intent was to do something like `res := ctx.NumberAdd(x,y)` which would always return a new Number taken from the free-list. However, this complicated clean reuse of Numbers. For example, example 2 would have been written like this for the rate computation:

	rate = ctx.NumberDivide(rate, hundred)        // rate=rate/100
	rate = ctx.NumberAdd(rate, one)               // rate=rate+1
	rate = ctx.NumberPower(rate, years)           // rate=rate**years

The previous *rate* is dereferenced at each called and free for garbage collection with no chance to send it back to the free-list. Even if the free-list did not Zero() the numbers, we would have to to something like this:

	rate = ctx.NumberDivide(ctx.FreeNumber(rate), hundred)

which is not good either from a usability and performance standpoint.

As a more concrete usage example, in a calculator application we have:

- a global context
- a global stack of numbers (implemented as a slice)

For all arithmetic computations, temporary numbers, etc., we use the idiomatic deferred call to FreeNumber(). When numbers get pushed off the stack, they are just discarded, without a call to FreeNumber(). When the user requests a change in precision, the global context is replaced by a newly created one with the requested precision. Numbers present on the stack are kept as is since they are still valid Numbers when used as operands in arithmetic functions.

## Threading, goroutines

The decNumber package is thread safe as long as threads do not share decContext or decNumber structures. The same goes for the Go wrapper. Goroutines should have their own context. For Numbers, if you need to share them, share by communicating.

## What about decDouble, decQuad ?

Right now, go-decnumber only supports decNumber. Adding support for any of the *float* modules would require:

- Adding the relevant type in the decnumber module
- Adding the relevant methods to the Context type
- A free list is less necessary for the float types since their size is static (128 bits), unlike Number which has a variable size (depending on the Context's precision) and require malloc/feee calls. Quads that are not used outside of a function body should be allocated on the stack (to be tested).

Another thing to consider is that for the float types, the Context is used only for error checking and rounding mode. Defining a Context interface with a Number and Quad implementation could also be a solution.

# Building / Installing

If you only intend to include this package in your own project, just run

	go get -u github.com/wildservices/go-decnumber

and you're all set.

Another option for package maintainers is to use the [.syso mechanism][syso] which greatly speeds up the build process. The idea is to bundle together all the .o files into a single .syso file. When such a file is present in a package folder, it will automatically be linked with the other object files.

To make the .syso file:

	cd libdecnumber
	make syso
	cd ..

And benefit:

	go test -tags="syso" -i ./...
	go test -tags="syso "./...

When the `syso` tag is specified in a Go build, the wrapper files (decContext.go and decNumber.go) are ignored and the build uses the precompiled object file libdecnumber_${GOOS}_${GOARCH}.syso. The difference with a static library is that we do not have to use a ``#cgo LDFLAGS` directive that would bring in the static library in client projects as well.

The top level Makefile (at the root of the package folder) has the following targets:

	make            # builds using the "normal" wrapper mechanism, removing any syso file beforehand
	make build      # build using the syso file, compiling it if necessary
	make test       # test using the syso file, compiling it if necessary
	make clean      # the usual + removes the syso file.

In the early days of the package, running tests using the Go->C wrappers (for only 2 C files) took 3 seconds on my workstation, versus only 1 second when using the syso mechanism. 

[syso]: https://code.google.com/p/go-wiki/wiki/GcToolchainTricks#Use_syso_file_to_embed_arbitrary_self-contained_C_code

# TODO

- CreateNewCustomContext() is just plain ugly. Implement SetDigits() and re-create the free-list in it. But Need to add a pointer to the freelist as a member of Number (pointer needs to be cleared when puting it back, avoid circular deps).
- Implement basic math functions.

# Licensing

	go-decnumber
	Copyright 2014 Denis Bernard. All rights reserved.

Use of this package is governed by a BSD-style license that can be found in the LICENSE file.

The decNumber library is made available under the terms of the ICU License -- ICU 1.8.1 and later,
which can be found in the LICENSE-ICU file.
