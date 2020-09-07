package array

import "fmt"
import "strings"
import "strconv"

// import "errors"

var _ = fmt.Printf
var _ = strconv.Atoi
var _ = strings.Contains

/* -------------------------------------------------------------------------------- */
// -- The `Slice` type models array slices.
/* -------------------------------------------------------------------------------- */

// Slice is an array slice used for, well, array slicing.
type Slice []Range

// -- NOTE: The standard convention is followed that start:stop represents the
//  		interval [start, stop[

/* -------------------------------------------------------------------------------- */
// -- METHODS
/* -------------------------------------------------------------------------------- */

// Start returns the starting index of each sub-slice in the slice `s`
func (s *Slice) Start() []int {
	result := make([]int, len(*s))
	for k := range *s {
		result[k] = (*s)[k][0]
	}
	return result
}

// Stop returns the last index of each sub-slice in the slice `s`
func (s *Slice) Stop() []int {
	result := make([]int, len(*s))
	for k := range *s {
		result[k] = (*s)[k][1]
	}
	return result
}

// Step returns the step size of each sub-slice in the slice `s`
func (s *Slice) Step() []int {
	result := make([]int, len(*s))
	for k := range *s {
		result[k] = (*s)[k][2]
	}
	return result
}

/* -------------------------------------------------------------------------------- */
// -- CONSTRUCTOR
/* -------------------------------------------------------------------------------- */

// MakeSlice creates a slice that can be used to slice an array. The input is a
// string representation of the slice to be created. To keep the code simple for now,
// an "optimistic" approach is taken where the return value is completely constructed
// and only then validated. Obviously, the construction of the results could terminate
// early if checking for logical conrrectness is interleaved with the construction
// steps. `MakeSlice` takes the output of `MakeSliceCore` and checks the indices
// against the dimensions of the array. No negative indices allowed.
func MakeSlice(s string, dims []int) (Slice, error) {

	// fmt.Printf("-- Inside MakeSlice\n")
	// fmt.Printf("-- Calling MakeSliceCore\n")

	slice, err := MakeSliceCore(s, dims)
	if err != nil {
		return slice, err
	}

	// fmt.Printf("-- Called MakeSliceCore\n")
	// fmt.Printf("-- Validating the returned slice\n")

	// -- Validation of the slice returned by MakeSliceCore
	for k := range slice {

		// fmt.Printf("-- Validation of %v\n", slice.SubSlices[k])

		// -- if start < 0 or start >= dim or
		// -- stop < 0 or stop > dim
		if (slice[k][0] < 0) || (slice[k][0] >= dims[k]) {
			return slice, fmt.Errorf("[func MakeSlice] The starting index of the range that corresponds to dimension %d is out of bounds", k)
		}

		if slice[k][1] > dims[k] {
			return slice, fmt.Errorf("[func MakeSlice] The stopping index of the range %v in dimension %d of the slice %s is larger than the extent of the dimension", slice[k], k, s)
		}

		// -- The stopping index of the range, in dimension k of the slice, (`k` is the loop counter) is negative but the step is positive.
		if (slice[k][1] < 0) && (slice[k][2] > 0) {
			return slice, fmt.Errorf("[func MakeSlice] The stopping index of the range %v in dimension %d of the slice %s is negative but the step is positive", slice[k], k, s)
		}

		if (slice[k][0] > slice[k][1]) && (slice[k][2] > 0) {
			return slice, fmt.Errorf("[func MakeSlice] The starting index of the range that corresponds to dimension %d is greater than the stopping index and the step is positive", k)
		}

		if (slice[k][0] < slice[k][1]) && (slice[k][2] < 0) {
			return slice, fmt.Errorf("[func MakeSlice] The starting index of the range that corresponds to dimension %d is less than the stopping index and the step is negative", k)
		}

		if slice[k][2] == 0 {
			return slice, fmt.Errorf("[func MakeSlice] The step of the range that corresponds to dimension %d is zero", k)
		}

	}

	return slice, err

}

/* -------------------------------------------------------------------------------- */
// TODO
/* -------------------------------------------------------------------------------- */

// MakeSliceWrap takes the result of `MakeSliceCore` and wraps around indices that
// exceed the dimensions of the array.
func MakeSliceWrap(s string, dims []int) (Slice, error) {
	slice, err := MakeSliceCore(s, dims)
	return slice, err
}

/* -------------------------------------------------------------------------------- */
// TODO
/* -------------------------------------------------------------------------------- */

// MakeSliceClip takes the result of `MakeSliceCore` and truncates the indices that
// exceed the dimensions of the array.
func MakeSliceClip(s string, dims []int) (Slice, error) {
	slice, err := MakeSliceCore(s, dims)
	return slice, err
}

/* -------------------------------------------------------------------------------- */

// MakeSliceCore constructs a `Slice` without checking the results against array
// dimensions or doing any other post-processing.
func MakeSliceCore(slice string, dims []int) (Slice, error) {

	// -- Remove the square brackets from both ends of the string representation of the input slice `s`.
	// -- Split the string representation at the commas.
	strSplitSliceRepr := strings.Split(strings.Trim(slice, "[]"), ",")

	// -- Check consistency of the input data. The string representation `s` of
	// -- the slice to be created should describe as many dimensions as there
	// -- are in `dims`.
	if len(strSplitSliceRepr) != len(dims) {
		return Slice{}, fmt.Errorf("inconsistent arguments passed to `MakeSliceCore`: mismatch between the dimensions implied by the string-slice and the dimensions passed as input to MakeCoreSlice")
	}

	// -- Make an empty slice that will be populated and returned by this function.
	// result := Slice{SubSlices: make([][3]int, len(strSplitSliceRepr))}
	result := make([]Range, len(strSplitSliceRepr))

	var err error

	// -- Iterate over the splitted string representation of the input string `s`
	// -- to handle each sub-slice. Each sub-slice corresponds to a dimension of
	// -- the matrix for which the slice will be used.
	for k := range strSplitSliceRepr {
		trimmedRange := strings.TrimSpace(strSplitSliceRepr[k])
		result[k], err = RangeStr2Arr(trimmedRange, dims[k])
		if err != nil {
			// -- Combine the error returned by `RangeStr2Arr` with additional info.
			// return result, fmt.Errorf(fmt.Sprintf("%s\nthe problem is likely to be in dimension %d\ninvalid/illegal/ill-formed input to array.MakeSlice", err.Error(), k))
			return result, MakeError("func MakeSliceCore", fmt.Sprintf("ill-formed array slice string representation; the problem is likely to be in dimension %d (zero-based counting) of slice %s", k, slice))
		}
	}

	// // -- Validation of the result
	// for k := range result.SubSlices {
	// 	fmt.Printf("-- %v\n", result.SubSlices[k])
	// }

	return result, nil

}

// /* -------------------------------------------------------------------------------- */
// // -- HELPERS
// /* -------------------------------------------------------------------------------- */

// /* -------------------------------------------------------------------------------- */
// // -- `RangeStr2Arr` is used by `MakeSlice`.
// /* -------------------------------------------------------------------------------- */

// RangeStr2Arr converts a string representation of 1-dimensional slice to a usable
// numeric representation.
// s: string representation of a `Range` (e.g., [0:3:2])
//
func RangeStr2Arr(s string, dim int) ([3]int, error) {

	// fmt.Printf("---- RangeStr2Arr(): input = (%q, %d)\n", s, dim)

	result := [3]int{}

	// -- Number of occurrences of the character ":" in the input string.
	numCols := strings.Count(s, ":")
	splitAtCol := strings.Split(s, ":")

	// fmt.Printf("---- RangeStr2Arr(): range = %q\n", s)
	// fmt.Printf("---- RangeStr2Arr(): splitAtCol = %#v\n", splitAtCol)

	var num int
	var err error

	switch numCols {
	case 0:
		// -- If there is no `:` in the range `s`, then the only valid case is
		// -- when `s` represents an int. That is, in this dimension, we pick an
		// -- element with a fixed index.
		// -- Example: `s == [0:10:2, 7]`; in the second dimension there is no `:`
		// -- so there has to be a number (int).
		// fmt.Printf("---- Number of colons in the input = 0\n")
		num, err = strconv.Atoi(s)
		if err != nil {
			return result, err
		}
		// -- The representation of a range with only a number in a given dimension.
		// -- For example, [0:10:2, 7] is equivalent to [0:10:2, 7:7:1]
		result[0], result[1], result[2] = num, num, 1
		return result, nil
	case 1:
		// -- If there is only one ":" in the sub-slice string, then the step has
		// -- been ommitted and it is equal to 1 by default.
		result[2] = 1
		// -- Range over the split-at-":" string representation of a range.
		// -- Examples of strings that `splitAtCol` might have come from: ":", "1:2", "1:", ":2"
		for k := range splitAtCol {
			trimmedRangeElem := strings.TrimSpace(splitAtCol[k])
			if trimmedRangeElem == "" {
				// fmt.Printf("-- No value provided for this element of the subslice\n")
				if k == 0 {
					result[0] = 0
				}
				if k == 1 {
					result[1] = dim
				}
			} else {
				num, err = strconv.Atoi(trimmedRangeElem)
				if err != nil {
					return result, err
				}
				// // -- Now that we successfully got an integer out of the string, we need
				// // -- to check that the integer is non-negative and less than the size of
				// // -- the dimension to which the sub-slice corresponds.
				// if (num < 0) || (num >= dim) {
				// 	return result, fmt.Errorf("index out of bounds")
				// }
				if k == 0 {
					result[0] = num
				}
				if k == 1 {
					result[1] = num
				}
			}
			// fmt.Printf("-- num = %v\n", num)
		}
	case 2:
		// -- Range over the silce of strings ([]string) obtained by splitting at
		// -- at ":" the string representation of the range `s` (which is the
		// -- first input argument to this function).
		// -- Examples of strings that `splitAtCol` might have come from: "::", "1:2:3", "1::", "1:2:", "::2"
		positiveStep := true
		for k := range splitAtCol {
			// -- Range over the slice in reverse.
			k = len(splitAtCol) - k - 1
			trimmedRangeElem := strings.TrimSpace(splitAtCol[k])
			// -- Start, stop, or step ommitted.
			if trimmedRangeElem == "" {
				if k == 0 { // -- The start of the range is missing.
					if positiveStep {
						result[0] = 0
					} else {
						result[0] = dim - 1
					}
				}
				if k == 1 { // -- The stop of the range is missing.
					if positiveStep {
						result[1] = dim
					} else {
						result[1] = -1
					}
				}
				if k == 2 { // -- The step of the range is missing.
					result[2] = 1 // -- The default step is 1.
				}
			} else {
				num, err = strconv.Atoi(trimmedRangeElem)
				if err != nil {
					return result, err
				}
				if k == 0 {
					result[0] = num
				}
				if k == 1 {
					result[1] = num
				}
				if k == 2 {
					result[2] = num
					// -- Record the fact that the step is negative. This is needed to
					// -- fill in default values when the start or stop of the range
					// -- is missing. For example, if the input range is "::-1", then
					// -- the default values that go in `result` are those that would
					// -- have gone if the input range were "::1", but swapped.
					if num < 0 {
						positiveStep = false
					}
				}
			}
			// fmt.Printf("-- num = %v\n", num)
		}
	default:
		return result, fmt.Errorf("invalid number of `:` in the string representation of the range %s", s)
	}

	// fmt.Printf("---- RangeStr2Arr(): input = (%q, %d)\n", s, dim)
	// fmt.Printf("---- RangeStr2Arr(): output = %v\n", result)

	return result, nil

}
