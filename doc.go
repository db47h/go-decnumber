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
floating-point decimal numbers directly, including infinite, NaN (Not a Number), and subnormal
values. Both arbitrary-precision and fixed-size representations are supported.

General usage notes:

The Number format is optimized for efficient processing of relatively short numbers; It does,
however, support arbitrary precision (up to 999,999,999 digits) and arbitrary exponent range (Emax
in the range 0 through 999,999,999 and Emin in the range -999,999,999 through 0).  Mathematical
functions (for example Exp()) as identified below are restricted more tightly: digits, emax, and
-emin in the context must be <= MaxMath (999999), and their operand(s) must be within these bounds.

Logical functions are further restricted; their operands must be finite, positive, have an exponent
of zero, and all digits must be either 0 or 1.  The result will only contain digits which are 0 or 1
(and will have exponent=0 and a sign of 0).

Operands to operator functions are never modified unless they are also specified to be the result
number (which is always permitted).  Other than that case, operands must not overlap.

Go implementation details:

The decimal32, decimal64 and decimal128 types are merged into Single, Double and Quad, respectively.

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

Error handling: although most arithmetic functions can cause errors, the standard Go error handling
is not used in its idiomatic form. That is, arithmetic functions do not return errors. Instead, the
type of the error is ORed into the status flags in the current context (Context type). It is the
responsibility of the caller to clear the status flags as required. The result of any routine which
returns a number will always be a valid number (which may be a special value, such as an Infinity or
NaN). This permits the use of much fewer error checks; a single check for a whole computation is
often enough.

To check for errors, get the Context's status with the Status() function (see the Status
type), or use the Context's ErrorStatus() function.

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

If a particular implementation needs always-initialized numbers (like any new Go variable being initialize
to the 0 value of its type), the pool's New function can be set for example to:

	func() interface{} { return dec.NewNumber(ctx.Digits()).Zero() }

The provided dec.Pool implementation is not thread safe and is only provided as a lightweight
alternative to sync.Pool.

If an application needs to change its arithmetic precision on the fly, any NumberPool built on top
of the affected Context's will need to be discarded and recreated along with the Context. This will
not affect existing numbers that can still be used as valid operands in arithmetic functions.

*/
package dec
