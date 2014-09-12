# Overview

go-decnumber is a go wrapper package around the [libDecnumber library](http://speleotrove.com/decimal/decnumber.html).

# Implementation details

The decNumber package is meant to be modular. There are the decContext module (required), decNumber module (for arbitrary presision arithmetic) and *float* modules, namely decSingle, decDouble and decQuad. The *float* modules are based on the 32-bit, 64-bit, and 128-bit decimal types in the IEEE 754 Standard for Floating Point Arithmetic. In contrast to the arbitrary-precision decNumber module, these modules work directly from the decimal-encoded formats designed by the IEEE 754 committee. Thir implementation is also faster: an Add() with 34 digits numbers takes 433 cycles with de decQuad module versus 1180 for the decNumber module.

Note that there is no *standard* libdecnumber.so. The decNumber package is provided as-is, with no makefile to make a shared library out of it (the Makefile in this reporitory is not part of the original decNumber archive and is only a test).

From a C programming perspective, all one has to do to compile code for decNumber is:

	gcc mysource.c decContext.c decQuad.c decNumber.c -o myprogram

or if only decNumber is needed:

	gcc mysource.c decContext.c decNumber.c -o myprogram

My initial intent was to split the wrapper into modules, in the same way as the C code. However, this lead to unsolvable compilation errors, like missing references in the Number module to the Context module. I also did not want to force the end-user of the package to build and install a custom shared library. Also, taking into consideration the issues discussed in the following section, I ended up making a monolithic package around decNumber and decContext.

## Numbers, Context and precision

Ont of the nice features of decNumber, and why I use it, is the ability to change the arithmetic precision
(i.e. number of significant digits) at runtime. The precision is held in a decContext structure and numbers are held in a decNumber structure. The caveat is that when dealing with arbitrary precision, the decNumber structures do not keep track of how many digits they can hold. It's up to the programmer to keep track of which decNumber structure was created to work with a given context.

A concrete example, the function Exp() is defined like this:

	decNumber * decNumberExp(decNumber *res, const decNumber *rhs, decContext *set)

It will set *res* to *e* raised to the power of *rhs*. The *rhs* operand can be in any precision (i.e. context independant). However, *\*res*, the decNumber structure that will hold the result, has to have enough storage space to hold the precision specified in the decContext *set*.

In a top-down functionnal programming model, this is not a serious issue. However, with goroutines flying all over the place, this can get messy. This lead to a few design choices in the go implementation:

- The configuration of a Context is immutable after creation (i.e. cannot change the number of digits, minimum and maximum exponents). Only rounding and status are alterable. If one needs to change precision on the fly, discard the existing context and create a new one with the required precision. Existing Numbers are still usable and valid Numbers.
- Contexts hold a free list of Numbers and Numbers are created by a Context method. This gives the following idiomatic code for temp Number creation:

	num := ctx.NewNumber()
	defer ctx.FreeNumber(num)

- Arithmetic functions are Context methods and always return a new Number taken from the free list. This leads to the same idomatic code than NewNumber:

	num := ctx.NumberAdd(x, y)
	defer ctx.FreeNumber(num)

- Freeing numbers is not mandatory. The FreeNumber method() only returns it to the free list. Actual resource cleanup is handled by the garbage collector and an internal call to SetFinalizer(). however:
  - A Number must not be used by the caller after calling FreeNumber()
  - FreeNumber() must be called on the Context that created it. If for some reason keeping track of this is not possible, just don't call FreeNumber().

As a more concrete usage example, in a calculator application we have:

- a global context
- a global stack of numbers (implemented as a slice)

For all arithmetic computations, temporary numbers, etc., we use the idiomatic deferred call to FreeNumber(). When numbers get pushed off the stack, they are just discarded, without a call to FreeNumber(). When the user requests a change in precision, the global context is just recreated with the requested precision. Numbers present on the stack are kept as is since they are still valid Numbers when used as operands in arithmetic functions.

## Threading, goroutines

The decNumber package is thread safe as long as threads do not share decContext or decNumber structures. The same goes for the Go wrapper. Goroutines should have their own context. For Numbers, if you need to share them, share by communicating.

## What about decDouble, decQuad ?

Right now, go-decnumber only supports decNumber. Adding support for any of the *float* modules would require:

- Adding the relevant type in the decnumber module
- Adding the relevant methods to the Context type
- figure out a way to have only one free list of Numbers, Quads, etc. Although a free list is less necessary for the float types since they are of a static size (up to 128 bits), unlike Number which are of variable size (debending on the Context's precision) and require malloc/feee calls.

Another thing to consider is that for the float types, the Context is used only for error checking and rounding mode.

# TODO

- Context: create accessor functions for clamping, emin and emax (for easier context duplication)

# Licensing

	go-decnumber
	Copyright 2014 Denis Bernard. All rights reserved.

Use of this package is governed by a BSD-style license that can be found in the LICENSE file.

The decNumber library is made available under the terms of the ICU License -- ICU 1.8.1 and later,
which can be found in the LICENSE-ICU file.
