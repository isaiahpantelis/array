package array

/* -------------------------------------------------------------------------------- */
// -- Metadata: attributes that every array must carry.
/* -------------------------------------------------------------------------------- */

// Metadata contains information that must accompany an array, but not the actual
// data stored in the array. Each concrete array type (in `cat.go`) has a `Metadata`
// member.
type Metadata struct {
	dims   []int
	ndims  int
	numels int
	micf   []int // (m)ulti-(i)ndex (c)onversion (f)actors
}

/* -------------------------------------------------------------------------------- */
// -- Metadata: methods
/* -------------------------------------------------------------------------------- */

/*
NOTE: 	The getters, below, are not defined because of some Pavlovian OO reflex.
		Rather, at this point, it seems like a bad idea to give access to a state
		that has to remain consistent while operations are applied to arrays.

EDIT:	Benchmarks show that there is a noticeable performance hit by calling
		methods instead of directly accessing the properties of a `Metadata`
		struct. Since performance is important of an array library, the properties
		of `Metadata` are now exported.

EDIT:   Things got messy after switching to direct access to the properties: there
		are name clashes between the setters of the array factory and the getters
		of the Metadata (an `ArrayFactory` has `Metadata` so if, for example, the
		field `Dims` is exported by `Metadata`, then `Dims()` cannot be used to
		construct an array as in `A := array.Factory().Dims([2, 2])`). Back to
		access via methods. Anoher way to resolve the name clashes is, of course,
		to use less concise function / method names.
*/

// Dims returns the dimensions of an array.
func (A *Metadata) Dims() []int {
	return A.dims
}

// Ndims returns the number of dimensions of an array.
func (A *Metadata) Ndims() int {
	return A.ndims
}

// Numels returns the number of elements of an array.
func (A *Metadata) Numels() int {
	return A.numels
}

// Micf returns the multi-index conversion factors of an array.
func (A *Metadata) Micf() []int {
	return A.micf
}
