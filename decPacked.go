// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !syso

/* This file is a wrapper around decPacked.c */

package dec

/*
// #cgo flags are specified in context.go
#include "go-decnumber.h"
#include "decPacked.c"
*/
import "C"
