// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber

/*
#include "go-decnumber.h"
#include "decContext.h"
*/
import "C"

// Status represents the status flags (exceptional conditions), and their names.
// The top byte is reserved for internal use
type Status uint32

const (
	ZeroStatus          Status = 0
	ConversionSyntax    Status = C.DEC_Conversion_syntax
	DivisionByZero      Status = C.DEC_Division_by_zero
	DivisionImpossible  Status = C.DEC_Division_impossible
	DivisionUndefined   Status = C.DEC_Division_undefined
	InsufficientStorage Status = C.DEC_Insufficient_storage // when malloc fails
	Inexact             Status = C.DEC_Inexact
	InvalidContext      Status = C.DEC_Invalid_context
	InvalidOperation    Status = C.DEC_Invalid_operation
	Overflow            Status = C.DEC_Overflow
	Clamped             Status = C.DEC_Clamped
	Rounded             Status = C.DEC_Rounded
	Subnormal           Status = C.DEC_Subnormal
	Underflow           Status = C.DEC_Underflow

	Errors      Status = C.DEC_Errors      // flags which are normally errors (result is qNaN, infinite, or 0)
	NaNs        Status = C.DEC_NaNs        // flags which cause a result to become qNaN
	Information Status = C.DEC_Information // flags which are normally for information only (finite results)
)

var statusString = map[Status]string{
	ZeroStatus:          "No status",
	ConversionSyntax:    "Conversion syntax",
	DivisionByZero:      "Division by zero",
	DivisionImpossible:  "Division impossible",
	DivisionUndefined:   "Division undefined",
	InsufficientStorage: "Insufficient storage",
	Inexact:             "Inexact",
	InvalidContext:      "Invalid context",
	InvalidOperation:    "Invalid operation",
	Overflow:            "Overflow",
	Clamped:             "Clamped",
	Rounded:             "Rounded",
	Subnormal:           "Subnormal",
	Underflow:           "Underflow",
}

// String returns a human-readable description of a status bit as a string..
// The bits set in the status field must comprise only bits defined.
// If no bits are set in the status field, the string “No status” is returned. If more than one
// bit is set, the string “Multiple status” is returned.
func (s *Status) String() string {
	if str, ok := statusString[*s]; ok {
		return str
	}
	return "Multiple status"
}

// SetFromString sets the status from a string. str is a string exactly equal to one that might be
// returned by Status.String().
// The status bit corresponding to the string is set.
//
// Returns a non-nil ContextError if str was equal to "Multiple status" or was not recognized.
func (s *Status) SetFromString(str string) error {
	for k, v := range statusString {
		if v == str {
			s.Set(k)
			return nil
		}
	}
	err := ConversionSyntax
	return err.ToError()
}

// Set sets one or more status bits in the status field. Since traps are
// not supported in the Go implementation, it acts like decContextSetStatusQuiet.
//
// Normally, only library modules use this function. Applications may clear status bits with
// Clear() or Zero() but should not set them (except, perhaps, for testing).
//
// Returns s.
func (s *Status) Set(newStatus Status) *Status {
	*s |= newStatus
	return s
}

// Clear clears (sets to zero) one or more bits in the status.
//
// Any 1 (set) bit in the status argument will cause the corresponding bit to be cleared.
//
// Returns s.
func (s *Status) Clear(mask Status) *Status {
	*s &^= mask
	return s
}

// Zero is used to clear (set to zero) all the status bits.
//
// Returns s.
func (s *Status) Zero() *Status {
	*s = 0
	return s
}

// Test tests bits in the status and returns true if any of the tested bits are 1.
func (s *Status) Test(mask Status) bool {
	return *s&mask != 0
}

// Save saves bits in current status. mask indicates the bits to be saved (the status bits that
// correspond to each 1 bit in the mask are saved). See the implementation of
// Context.NewNumberFromString() for a typical use of Save().
//
// Returns a *Status that represents the AND of the mask and the current status.
func (s *Status) Save(mask Status) *Status {
	res := *s & mask
	return &res
}

// Restore restores bits in the current status.
// newStatus is the source for the bits to be restored.
// mask indicates the bits to be restored (the status bit that corresponds to each 1 bit in the
// mask is set to the value of the correspnding bit in newstatus).
//
// Returns s.
func (s *Status) Restore(newStatus Status, mask Status) *Status {
	*s &^= mask            // clear bits
	*s |= mask & newStatus // or in the new bits
	return s
}

// Func ToError() checks the status for any error condition and returns, as an error,
// a ContextError if any, nil otherwise.
// Convert the return value with err.(decnumber.ContextError) to compare it
// against any of the Status values.
//
// Status bits considered errors are:
//
//	DivisionByZero
//	ConversionSyntax
//	DivisionImpossible
//	DivisionUndefined
//	InsufficientStorage
//	InvalidContext
//	InvalidOperation
//	Overflow
//	Underflow
func (s *Status) ToError() error {
	if e := *s & Errors; e != 0 {
		return ContextError(e)
	}
	return nil
}

// ContextError represents an error condition for a Context. One can check if the last operation
// in a Context generated an error either with Context.ErrorStatus() (returns a ContextError cast as
// an error) or Context.TestStatus(Context.Errors) (returns true if an error occured).
type ContextError Status

// Error returns a string representation of the error status
func (e ContextError) Error() string {
	return (*Status)(&e).String()
}
