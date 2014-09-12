// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Local configuration file for go-decnumber
// this will work at least with gcc
#ifdef __BYTE_ORDER__
	#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
		#define DECLITEND 1
	#elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
		#define DECLITEND 0
	#else
		#error "Unsupported byte order"
	#endif
#else
	#error "Unable to determine byte order. __BYTE_ORDER__ not defined"
#endif

#define DECPRINT 0
#define DECEXTFLAG 1
