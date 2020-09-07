package array

import "fmt"
import "unicode/utf8"

/* -------------------------------------------------------------------------------- */
// -- This file was automatically generated at 2020-08-22 15:58:56.6453 +0000 UTC
// -- The generating template is `../../templates/view.go`
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewInt8 struct {
	Array *Int8
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewInt8) Get(indices ...int) int8 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewInt8.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewInt8.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewInt8.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewInt8.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewInt8) Get(indices ...int) int8 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewInt8) Set(val int8, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewInt8) At(indices ...int) *int8 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewInt8) Iterator(done <-chan struct{}) <-chan *int8 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *int8)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Int8
func (A *ViewInt8) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewInt8) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewUint8 struct {
	Array *Uint8
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewUint8) Get(indices ...int) uint8 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewUint8.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewUint8.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewUint8.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewUint8.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewUint8) Get(indices ...int) uint8 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewUint8) Set(val uint8, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewUint8) At(indices ...int) *uint8 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewUint8) Iterator(done <-chan struct{}) <-chan *uint8 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *uint8)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Uint8
func (A *ViewUint8) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewUint8) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewInt16 struct {
	Array *Int16
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewInt16) Get(indices ...int) int16 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewInt16.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewInt16.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewInt16.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewInt16.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewInt16) Get(indices ...int) int16 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewInt16) Set(val int16, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewInt16) At(indices ...int) *int16 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewInt16) Iterator(done <-chan struct{}) <-chan *int16 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *int16)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Int16
func (A *ViewInt16) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewInt16) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewUint16 struct {
	Array *Uint16
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewUint16) Get(indices ...int) uint16 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewUint16.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewUint16.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewUint16.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewUint16.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewUint16) Get(indices ...int) uint16 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewUint16) Set(val uint16, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewUint16) At(indices ...int) *uint16 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewUint16) Iterator(done <-chan struct{}) <-chan *uint16 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *uint16)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Uint16
func (A *ViewUint16) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewUint16) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewInt32 struct {
	Array *Int32
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewInt32) Get(indices ...int) int32 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewInt32.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewInt32.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewInt32.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewInt32.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewInt32) Get(indices ...int) int32 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewInt32) Set(val int32, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewInt32) At(indices ...int) *int32 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewInt32) Iterator(done <-chan struct{}) <-chan *int32 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *int32)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Int32
func (A *ViewInt32) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewInt32) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewUint32 struct {
	Array *Uint32
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewUint32) Get(indices ...int) uint32 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewUint32.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewUint32.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewUint32.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewUint32.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewUint32) Get(indices ...int) uint32 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewUint32) Set(val uint32, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewUint32) At(indices ...int) *uint32 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewUint32) Iterator(done <-chan struct{}) <-chan *uint32 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *uint32)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Uint32
func (A *ViewUint32) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewUint32) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewInt64 struct {
	Array *Int64
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewInt64) Get(indices ...int) int64 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewInt64.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewInt64.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewInt64.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewInt64.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewInt64) Get(indices ...int) int64 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewInt64) Set(val int64, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewInt64) At(indices ...int) *int64 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewInt64) Iterator(done <-chan struct{}) <-chan *int64 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *int64)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Int64
func (A *ViewInt64) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewInt64) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewUint64 struct {
	Array *Uint64
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewUint64) Get(indices ...int) uint64 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewUint64.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewUint64.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewUint64.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewUint64.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewUint64) Get(indices ...int) uint64 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewUint64) Set(val uint64, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewUint64) At(indices ...int) *uint64 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewUint64) Iterator(done <-chan struct{}) <-chan *uint64 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *uint64)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Uint64
func (A *ViewUint64) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewUint64) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewInt struct {
	Array *Int
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewInt) Get(indices ...int) int {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewInt.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewInt.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewInt.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewInt.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewInt) Get(indices ...int) int {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewInt) Set(val int, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewInt) At(indices ...int) *int {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewInt) Iterator(done <-chan struct{}) <-chan *int {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *int)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Int
func (A *ViewInt) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewInt) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewUint struct {
	Array *Uint
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewUint) Get(indices ...int) uint {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewUint.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewUint.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewUint.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewUint.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewUint) Get(indices ...int) uint {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewUint) Set(val uint, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewUint) At(indices ...int) *uint {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewUint) Iterator(done <-chan struct{}) <-chan *uint {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *uint)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Uint
func (A *ViewUint) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewUint) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewUintptr struct {
	Array *Uintptr
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewUintptr) Get(indices ...int) uintptr {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewUintptr.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewUintptr.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewUintptr.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewUintptr.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewUintptr) Get(indices ...int) uintptr {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewUintptr) Set(val uintptr, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewUintptr) At(indices ...int) *uintptr {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewUintptr) Iterator(done <-chan struct{}) <-chan *uintptr {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *uintptr)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Uintptr
func (A *ViewUintptr) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, A.Data[k])
		//
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Array.Data[k])
		//
		//
		startl = ""
		dimChange = ""

		// //
		// result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
		// //
		// //

	}

	return result

}

func (v *ViewUintptr) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewFloat32 struct {
	Array *Float32
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewFloat32) Get(indices ...int) float32 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewFloat32.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewFloat32.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewFloat32.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewFloat32.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewFloat32) Get(indices ...int) float32 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewFloat32) Set(val float32, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewFloat32) At(indices ...int) *float32 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewFloat32) Iterator(done <-chan struct{}) <-chan *float32 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *float32)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Float32
func (A *ViewFloat32) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%10.4f\t", dimChange, startl, A.Data[k])
		result += fmt.Sprintf("%s%s" + PrintVerbFloats + "\t", dimChange, startl, A.Get(Indices(A.dims, k)...))
		//
		//
		//
		startl = ""
		dimChange = ""

		// result += fmt.Sprintf("%s%*s" + PrintVerbFloats + "\t", dimChange, startl, maxLabelChars, A.Data[k])
		// //
		// //
		// //

	}

	return result

}

func (v *ViewFloat32) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `View` type
/* -------------------------------------------------------------------------------- */

// View is a view into an array.
type ViewFloat64 struct {
	Array *Float64
	S     Slice
	Metadata
	Err error
}

/* -------------------------------------------------------------------------------- */
// -- `View` methods
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- `Get`
/* -------------------------------------------------------------------------------- */

// -- NOTE: View.Get does not check whether the indices are within the bounds of the 
//  		view. As it is now, the method allows access to elements of the original
//			array that are outside the view.

// Get returns the value of a single element of a `View` of an array.
func (v *ViewFloat64) Get(indices ...int) float64 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	// fmt.Printf("[ViewFloat64.Get] indices = %v\n", indices)
	auxi := make([]int, len(indices))
	// fmt.Printf("[ViewFloat64.Get] auxi = %v\n", auxi)
	for k := range indices {
		// if (indices[k] < 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		// if (indices[k] >= v.dims[k]) && (v.S[k][2] > 0) {
		// 	if v.Err == nil {
		// 		v.Err = fmt.Errorf("out of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	} else {
		// 		v.Err = fmt.Errorf(v.Err.Error() + "\nout of bounds access : view with dimensions %v subscripted with %v : the subscript %d used in dimension #%d is not in the range [0, %d[", v.dims, indices, indices[k], k+1, v.dims[k])
		// 	}
		// }
		auxi[k] = v.S[k].Get(indices[k])
	}
	// fmt.Printf("[ViewFloat64.Get] auxi = %v\n", auxi)
	// fmt.Printf("[ViewFloat64.Get] DotProdInt(v.Array.Micf(), auxi) = %v\n", DotProdInt(v.Array.Micf(), auxi))
	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
}
// // Get returns the value of a single element of a `View` of an array.
// func (v *ViewFloat64) Get(indices ...int) float64 {
// 	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
// 	// -- of the array that v.Array points to) to get the linear index (into the
// 	// -- array) of the element specified by the input `indices`.
// 	auxi := make([]int, len(indices))
// 	for k := range indices {
// 		auxi[k] = v.S[k].Get(indices[k])
// 	}
// 	return v.Array.Data[DotProdInt(v.Array.Micf(), auxi)]
// }

/* -------------------------------------------------------------------------------- */
// -- `Set`
/* -------------------------------------------------------------------------------- */

// Set sets the value of a single element of a `View` of an array.
func (v *ViewFloat64) Set(val float64, indices ...int) {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	v.Array.Data[DotProdInt(v.Array.Micf(), auxi)] = val
}

/* -------------------------------------------------------------------------------- */
// -- `At`
/* -------------------------------------------------------------------------------- */

// At returns the address of a single element of a `View` of an array.
func (v *ViewFloat64) At(indices ...int) *float64 {
	// -- Auxiliary indices that will be combined with the v.Array.micf (the micf
	// -- of the array that  v.Array points to) to get the linear index (into the
	// -- array) of the element specified by the input `indices`.
	auxi := make([]int, len(indices))
	for k := range indices {
		auxi[k] = v.S[k].Get(indices[k])
	}
	return &(v.Array.Data[DotProdInt(v.Array.Micf(), auxi)])
}

/* -------------------------------------------------------------------------------- */
// -- Iterator over a view
/* -------------------------------------------------------------------------------- */
// Iterator returns an iterator over the view of an array
func (v *ViewFloat64) Iterator(done <-chan struct{}) <-chan *float64 {
	// fmt.Printf("-- [Iterator] numels = %v\n", v.Numels())
	iter := make(chan *float64)
	if v.Numels() == 0 {
		close(iter)
		return iter
	}
	var k int = 0
	// var arrayElem float64 = v.Get(Indices(v.Dims(), k)...)
	// fmt.Printf("-- arrayElem = %v\n", arrayElem)
	go func() {
		// defer fmt.Printf("-- Closing the view channel\n")
		defer close(iter)
		for {
			select {
			case iter <- v.At(Indices(v.Dims(), k)...):
				if k < v.Numels()-1 {
					k++
				} else {
					return
				}
			case <-done:
				return
			}
		}
	}()
	return iter
}

/* -------------------------------------------------------------------------------- */
// -- `String`
/* -------------------------------------------------------------------------------- */
// String prints an Float64
func (A *ViewFloat64) String() string {

	if A.ndims == 0 {
		return "[]"
	}

	var result string = ""
	var dimChange string = ""
	var ndims int = A.ndims

	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var prev []int

	var row int = 0
	newRow := row

	// -- The row label (a string) with the maximum number of characters. We want this information to align the labels properly.
	// -- Note that the last label is the label of the first element in the last row,
	// -- not the label of the last element in the view. This matters because the 
	// -- last element in the view may have, say, label (20,10) whereas the last 
	// -- label is (20,0); the latter has one rune less.

	// fmt.Printf("-- A.dims = %v\n", A.dims)
	// fmt.Printf("-- A.ndims = %v\n", A.ndims)
	// fmt.Printf("-- A.dims[A.ndims-1] = %v\n", A.dims[A.ndims-1])
	// fmt.Printf("-- A.numels = %v\n", A.numels)
	
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := 0; k < A.numels; k++ {
		counter.Set(k)
		prev = counter.Previous()
		diff := diff(prev, counter.digits, counter.dims)
		if EndsIn1Int(diff) {
			// dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
			dimChange = fmt.Sprintf("\n")
		}
		row = newRow
		newRow = counter.digits[ndims-2]
		if newRow != row {
			// startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
			startl = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, k), "()", ","), maxLabelChars-2)
		}
		// result += fmt.Sprintf("%s%s%10.4f\t", dimChange, startl, A.Data[k])
		result += fmt.Sprintf("%s%s" + PrintVerbFloats + "\t", dimChange, startl, A.Get(Indices(A.dims, k)...))
		//
		//
		//
		startl = ""
		dimChange = ""

		// result += fmt.Sprintf("%s%*s" + PrintVerbFloats + "\t", dimChange, startl, maxLabelChars, A.Data[k])
		// //
		// //
		// //

	}

	return result

}

func (v *ViewFloat64) Dir() string {
	s := fmt.Sprintf("Array: %T (%p)\nSlice: %v\nMetadata:\n\tdims: %v\n\tndims: %v\n\tnumels: %v\n\tmicf: %v\n\tErr: %v", v.Array, v.Array, v.S, v.Metadata.Dims(), v.Metadata.Ndims(), v.Metadata.Numels(), v.Micf(), v.Err)
	return s
}


/* -------------------------------------------------------------------------------- */

