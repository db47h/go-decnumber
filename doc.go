// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package dec is a Go wrapper around the decNumber library (arbitrary precision decimal arithmetic).

For more information about the decNumber library, and detailed documentation,
see http://speleotrove.com/decimal/dec.html

It implements the General Decimal Arithmetic Specification in ANSI C. This specification defines a
decimal arithmetic which meets the requirements of commercial, financial, and human-oriented
applications. It also matches the decimal arithmetic in the IEEE 754 Standard for Floating Point
Arithmetic.

The library fully implements the specification, and hence supports integer, fixed-point, and
floating- point decimal numbers directly, including infinite, NaN (Not a Number), and subnormal
values. Both arbitrary-precision and fixed-size representations are supported.

Go implementation details:

Contexts are created with a immutable precision (i.e. number of digits). If one needs to change
precision on the fly, discard the existing context and create a new one with the required precision.

From a programming standpoint, any initialized Number is a valid operand in arithmetic operations,
regardless of the settings or existence of its creator Context (not to be confused with having a
valid value in a given arithmetic operation).

Arithmetic functions are Number methods. The value of the receiver of the method will be set to the
result of the operation. For example:

	n.Add(x, y, context) // n = x + y

Arithmetic methods always return the receiver in order to allow chain calling:

	n.Add(x, y, context).Multiply(n, z, context) // n = (x + y) * z

Using the same Number as operand and result, like in n.Multiply(n, n, ctx), is legal and will not
produce unexpected results.

The package provides facilities for managing free-lists of Numbers in order to relieve pressure on
the garbage collector in computation intensive applications. NumberPool is in fact a simple wrapper
around a *Context and a sync.Pool, or the lighter dec.Pool, which will automatically cast the return
value of Get() to the desired type.

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

Note the use of pool.Context on the last statement.

The provided dec.Pool implementation is not thread safe and is only provided as a lightweight
alternative to sync.Pool.

If an application needs to change its arithmetic precision on the fly, any NumberPool built on top
of the affected Context's will need to be discarded and recreated along with the Context. This will
not affect existing numbers that can still be used as valid operands in arithmetic functions.

*/
package dec
