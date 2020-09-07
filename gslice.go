package array


/* -------------------------------------------------------------------------------- */
// -- This file was automatically generated at 2020-08-22 15:58:56.637988 +0000 UTC
// -- The generating template is `../../templates/gslice.go`
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- NOTE: There is no error checking because these functions have to be fast.
/* -------------------------------------------------------------------------------- */

// DotProdInt returns the inner product of two []int].
func DotProdInt(u, v []int) (dotProd int) {
	// -- NOTE: If len(u) > len(v), the following code will cause an access-out-of-bounds error.
	// var dotProd int = 0
	for k := range u {
		dotProd += u[k] * v[k]
	}
	// return dotProd
	return
}

// CummProdInt returns the cummulative product of a slice []int,
// starting the calculation at position k. If `k` is not supplied, then the
// calculation starts at v[0].
// The argument k is variadic so that it's optional; we don't need more than one
// value, but we want to have at most one value.
func CummProdInt(v []int, k ...int) int {

	var prod int = 1
	var j0 int = 0

	if len(k) != 0 {
		j0 = k[0]
	}

	for j := j0; j < len(v); j++ {
		prod *= v[j]
	}

	return prod

}

// EndsIn1Int checks whether the input slice `x` ends in at least 3 
// consecutive 1s (ones).
func EndsIn1Int(x []int) bool {

	// -- Count the consecutive 1s at the end of the input slice.
	var numOf1 int = 0
	for k := len(x) - 1; k > -1; k-- {
		// fmt.Printf("-- x[%v] = %v\n", k, x[k])
		if x[k] == 1 {
			numOf1++
			continue
		}
		break
	}

	if numOf1 < 3 {
		return false
	}

	return true

}


/* -------------------------------------------------------------------------------- */

