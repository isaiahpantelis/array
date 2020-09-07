package array

import "unicode/utf8"
import "strings"
import "strconv"
import "fmt"

var _ = fmt.Printf

/* -------------------------------------------------------------------------------- */
// -- Free functions
/* -------------------------------------------------------------------------------- */

// -- Functions like `Index` and `Indices` should not require the creation of a
// -- matrix. It is wasteful to allocate a matrix simply to access its attributes.
// -- Obviously, if a matrix has already been allocated, its attributes can be used
// -- as inputs to the functions below.

// Index returns the linear index (into an array) that corresponds to the multi-index
// `indices` passed as input. The conversion requires the multi-index conversion
// factors as additional input.
func Index(micf []int, indices ...int) int {
	return DotProdInt(indices, micf)
}

/* -------------------------------------------------------------------------------- */

// Indices returns the multi-index (into an array) that corresponds to the linear
// index `index` passed as input. The conversion requires the dimensions of the
// array as additional input.
func Indices(dims []int, index int) []int {

	// dbgprfx := "[func Indices]>>"

	// multi-index result
	lenDims := len(dims)
	mindex := make([]int, lenDims, lenDims)

	num := index
	den := CummProdInt(dims[1:])
	// fmt.Printf("---- %s num = %v\n", dbgprfx, num)
	// fmt.Printf("---- %s den = %v\n", dbgprfx, den)

	// fmt.Printf("-- %s dims = %v\n", dbgprfx, dims)
	// fmt.Printf("---- %s index = %v\n", dbgprfx, index)

	var k, q, r int

	for k = 1; k < lenDims; k++ {
		q = num / den
		r = num % den
		// fmt.Printf("---- %s q = %v\n", dbgprfx, q)
		// fmt.Printf("---- %s r = %v\n", dbgprfx, r)
		// -- TODO: Eventually remove this verification step.
		if correctDivision := (q*den+r == num); !correctDivision {
			panic("** The conversion from a single index to a multi-index is incorrect.\n")
		}
		mindex[k-1] = q
		num = r
		den = CummProdInt(dims[(k + 1):])
	}

	mindex[k-1] = r
	// fmt.Printf("---- %s mindex = %v\n", dbgprfx, mindex)
	return mindex

}

/* -------------------------------------------------------------------------------- */

// NumelsFromDims returns the total numbers of elements contained in an `Array` whose dimensions are described by `dims`.
func NumelsFromDims(dims []int) int {
	var numels int = 1
	for k := range dims {
		numels *= dims[k]
	}
	return numels
}

/* -------------------------------------------------------------------------------- */

// MultiIndexConversionFactors returns a slice `micf` of integers such that the
// inner product <micf,(i_1, i_2,..., i_p)>,  where (i_1, i_2,..., i_p) is a
// multi-index with p == len(A.dims), satisfies
// A[<micf,(i_1, i_2,..., i_p)>] == A[i_1, i_2,..., i_p].
// In other words, the result is used to convert multi-indices to linear indices
// for arrays.
func MultiIndexConversionFactors(dims []int, ndims int) (micf []int) {

	for k := range dims {
		if k == ndims-1 {
			// -- The last conversion factor (the one corresponding to the dimension that changes the fastest) is 1.
			micf = append(micf, 1)
		} else {
			// -- Otherwise, the k-th conversion factor is the cummulative product, starting at k+1, of A.dims.
			micf = append(micf, CummProdInt(dims, k+1))
		}
	}

	return

}

/* -------------------------------------------------------------------------------- */

// Indices2Str converts a slice of type []int to a customised string representation.
// Example input arguments:
// indices: []int{1, 2, 3}
// bracketType: "()" or "[]"
// separator: ","
func Indices2Str(indices []int, bracketType string, separator string) (result string) {
	var bracketRune rune
	// -- The opening bracket
	if len(bracketType) > 0 {
		bracketRune, _ := utf8.DecodeRuneInString(bracketType)
		result = strings.TrimSpace(string(bracketRune))
	}
	for k := range indices {
		result += strconv.Itoa(indices[k])
		if k < len(indices)-1 {
			result += separator
		}
	}
	// fmt.Printf("-- [Indices2Str] result = %#v\n", result)
	// -- Add the closing bracket.
	if len(bracketType) > 1 {
		bracketRune, _ = utf8.DecodeRuneInString(bracketType[1:])
		result += strings.TrimSpace(string(bracketRune))
	}
	return result
}

/* -------------------------------------------------------------------------------- */

func lastRunes(s string, n int) (result string) {
	N := len(s)
	// -- If the string is empty or more runes than bytes are requested, return the string itself.
	// -- Always holds: #runes <= #bytes
	if (N == 0) || (n > N) {
		return s
	}
	for i := 0; i < n && N > 0; i++ {
		_, size := utf8.DecodeLastRuneInString(s[:N])
		N -= size
	}
	result = s[N:]
	return
}
