// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !syso

/* This file is a wrapper around decContext.c */

package decnumber

/*
#cgo CFLAGS: -Ilibdecnumber

#include "go-decnumber.h"
#include "decContext.h"
#include <stdlib.h>

#include "decContext.c"
*/
import "C"
