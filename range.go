package array

import "fmt"
import "math"

// import "errors"

/*
	NOTE: The standard convention is followed that start:stop represents the interval [start, stop[
*/

/* -------------------------------------------------------------------------------- */
// -- `Range` type
/* -------------------------------------------------------------------------------- */

// Range represents a range of integers with a stride/step in the form [start, stop, step].
type Range [3]int

/* -------------------------------------------------------------------------------- */
// -- `Range` constructor
/* -------------------------------------------------------------------------------- */

// MakeRange constructs a `Range` object from 3 integers that represent the start,
// stop, and step of a range. Of course, since a `Range` is simply an array of type
// [3]int, it can be constructed directly (e.g., range := Range{1, 2, 1}); however,
// not every combination of values for start, stop, and step is logically valid and
// the function `MakeRange` checks the validity of the inputs. "Logically valid" has
// a meaning in a given context of rules and definitions. The latter are implicit in
// `MakeRange`'s logic. For example, `start` is not allowed to be larger than `stop`,
// but, alternatively, `start > stop` could have been allowed to signify an empty
// range.
func MakeRange(start, stop, step int) (Range, error) {
	var err error
	if step == 0 {
		err = fmt.Errorf("step == 0")
	}
	// if (start > stop) && (step > 0) {
	// 	err = fmt.Errorf("start > stop and step > 0; a start index grater than a stop index can be used in a range if the step is negative")
	// }
	return Range{start, stop, step}, err
}

/* -------------------------------------------------------------------------------- */
// -- `Range` methods
/* -------------------------------------------------------------------------------- */

// Start returns the start/beginning of the range.
func (r *Range) Start() int { return r[0] }

// Stop returns the stop/end of the range.
func (r *Range) Stop() int { return r[1] }

// Step returns the step/stride of the range.
func (r *Range) Step() int { return r[2] }

// Get returns the k-th element of a range.
func (r *Range) Get(k int) int {
	return r[0] + k*r[2]
}

// Numels returns the number of elements in a Range.
// A range with zero step is considered ill-formed, however `Numels` will still
// attempt to calculate the number of elements in the range and return an error
// with the result.
func (r *Range) Numels() (int, error) {
	var result int
	var err error
	switch {
	default: // Without this default case, the compiler will complain about missing return statement.
		fallthrough
	/* -------------------------------------------------------------------------------- */
	// -- start == stop
	/* -------------------------------------------------------------------------------- */
	case r[0] == r[1]:
		// -- step != 0
		if r[2] != 0 {
			result, err = 1, nil
		}
		// -- step == 0
		// -- In this case, we can calculate the number of elements to be zero
		// -- since start == stop, but we also signify to the caller that the
		// -- range has an invalid step.
		if r[2] == 0 {
			result, err = 1, fmt.Errorf("[Range.Numels()] the step of the range %v is equal to 0", r)
		}
		return result, err
	/* -------------------------------------------------------------------------------- */
	// -- start < stop
	/* -------------------------------------------------------------------------------- */
	case r[0] < r[1]:
		// -- step > 0
		if r[2] > 0 {
			result, err = int(math.Ceil(float64(r[1]-r[0])/float64(r[2]))), nil
		}
		// -- step < 0
		// -- The caller can ignore the error and view the result as a sensible
		// -- value (python returns an empty list when slicing with start < stop and step < 0).
		if r[2] < 0 {
			result, err = 0, fmt.Errorf("[Range.Numels()] start < stop but step < 0 in the range %v", r)
		}
		// -- step == 0
		// -- Zero is not a valid value for the step of a range (which is the third element r[2]).
		if r[2] == 0 {
			result, err = 0, fmt.Errorf("[Range.Numels()] the step of the range %v is equal to 0", r)
		}
		return result, err
	/* -------------------------------------------------------------------------------- */
	// -- start > stop
	/* -------------------------------------------------------------------------------- */
	case r[0] > r[1]:
		// -- step > 0
		if r[2] < 0 {
			result, err = int(math.Ceil(float64(r[0]-r[1])/float64(-r[2]))), nil
		}
		// -- step < 0
		// -- The caller can ignore the error and view the result as a sensible
		// -- value (python returns an empty list when slicing with start > stop and step > 0).
		if r[2] > 0 {
			result, err = 0, fmt.Errorf("[Range.Numels()] start > stop but step > 0 in the range %v", r)
		}
		// -- step == 0
		// -- Zero is not a valid value for the step of a range (which is the third element r[2] of a range).
		if r[2] == 0 {
			result, err = 0, fmt.Errorf("[Range.Numels()] the step of the range %v is equal to 0", r)
		}
		return result, err
	}
}
