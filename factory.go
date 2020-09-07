package array

import "fmt"

/* -------------------------------------------------------------------------------- */
// -- ARRAY FACTORY
/* -------------------------------------------------------------------------------- */

// ArrayFactory is to be thought of as a struct that contains the settings or
// blueprint needed to construct a new array. Together with its methods, it is the
// factory that makes new arrays.
//
// Note: Normally, `ArrayFactory` instances will not be used directly. Their main
// use is in the "builder pattern"; the latter is used because there is no
// "constructor overloading" available in Go.
type ArrayFactory struct {
	// fillVal: Value used to fill the newly constructed array; f(ill)val(ue)
	// fillVal interface{}
	Metadata
}

// Factory returns an array factory on which `Make` can be called to construct an
// instance of the core `Array` type. The `Make<ElementType>` methods are
// implemented in `concrete.go`
func Factory() *ArrayFactory {
	return &ArrayFactory{}
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY-FACTORY METHODS
/* -------------------------------------------------------------------------------- */

// Dims sets the value of `dims` (in the array factory) representing the dimensions
// of the newly constructed array
func (af *ArrayFactory) Dims(dims ...int) *ArrayFactory {
	// fmt.Printf("[ArrayFactory.Dims]>> dims = %v\n", dims)
	af.dims = dims
	return af
}

// Fill sets the value `fillVal` (in the array factory) used to fill in the newly
// constructed array
// func (af *ArrayFactory) Fill(fillVal interface{}) *ArrayFactory {
// 	af.fillVal = fillVal
// 	return af
// }

// SetMetadata completes the initialisation of `Metadata` inside the receiver `af`.
// SetMetadata checks that `af.Dims` has meaningful values and calculates `numels`
// and `micf`. The meta-data include all the information to define the array,
// except for the elements of the array themselves.
func (af *ArrayFactory) SetMetadata() (*ArrayFactory, error) {

	if err := checkPreconditions(af); err != nil {
		return af, err
	}

	// -- Construct the meta-data using a composite literal according to the
	// -- "specs" in `ArrayFactory`.
	af.ndims = len(af.dims)
	af.numels = NumelsFromDims(af.dims)
	af.micf = MultiIndexConversionFactors(af.dims, len(af.dims))

	return af, nil

}

/* -------------------------------------------------------------------------------- */
// -- Helper functions
/* -------------------------------------------------------------------------------- */

// checkPreconditions checks whether the properties of a given `ArrayFactory`
// satisfy certain constraints. `checkPreconditions` is used inside `Make` before
// attempting to construct a new `Array`.
func checkPreconditions(af *ArrayFactory) error {

	if len(af.dims) == 1 {
		return fmt.Errorf("all arrays must have at least two dimensions, except for the empty array whose array of dimensions is empty")
	}

	for k := range af.dims {
		if af.dims[k] <= 0 {
			return fmt.Errorf("one of the requested dimensions is negative or zero: dims = %v", af.dims)
		}
	}

	return nil

}

/* -------------------------------------------------------------------------------- */

// emptyArray checks whether the properties of the `ArrayFactory` af imply that the
// array to be constructed should be empty.
func emptyArray(af *ArrayFactory) bool {

	// -- If the input slice that contains the dimensions was of zero length, return an empty array.
	if len(af.dims) == 0 {
		return true
	}

	// -- If one of the dimensions is zero, return an empty array.
	// if NumelsFromDims(af.dims) == 0 {
	// 	return true
	// }

	return false

}
