// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "."
	"testing"
)

func TestQuad_EngString(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal64, 0)
	q := new(dec.Quad)
	q.FromString("123.4e7", ctx)
	s := q.EngString()
	if s != "1.234E+9" {
		t.Fatalf("Expected 1.234E+9, got %s", s)
	}
}
