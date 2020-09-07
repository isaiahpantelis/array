package array

import "runtime"

var ncpu int = runtime.NumCPU()

/* -------------------------------------------------------------------------------- */
// -- Print verbs that control the printing of arrays.
/* -------------------------------------------------------------------------------- */

// PrintVerbFloats is for floats
var PrintVerbFloats string = "%12.4f"

// PrintVerbInts is for ints
var PrintVerbInts string = "%-10d"

// PrintVerbStrings is for strings
var PrintVerbStrings string = "%s"

// PrintVerbBools is for bools
var PrintVerbBools string = "%v"

/* -------------------------------------------------------------------------------- */
// -- Control the printing of error messages.
/* -------------------------------------------------------------------------------- */

// ErrPrefix is the prefix added to error messages.
var ErrPrefix string = "[error:ArrPkgError]"

// -- Below this number of array elements, sequential filling is used.
// var fillSeqThreshold int = 1001
