package array

import "fmt"
import "io"
import "sync"
import "unicode/utf8"
import "os"
import "encoding/csv"
import "encoding/json"
import "strings"
import "io/ioutil"
import "strconv"

/* -------------------------------------------------------------------------------- */
// -- This file was automatically generated at 2020-08-22 15:58:56.626638 +0000 UTC
// -- The generating template is `../../templates/cat.go`
/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Int8 is a concrete array type whose elements are of type `int8`. 
// Int8 is defined by composition of `Metadata` and the slice `Data`.
type Int8 struct {
	Metadata
	Data []int8
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Int8) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Int8.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Int8) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseInt(record[j], 10, 8); A.Data[k] = int8(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Int8) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Int8.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Int8.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Int8.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Int8) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Int8.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Int8) View(slice Slice) (*ViewInt8, error) {

	view := &ViewInt8{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Int8.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Int8) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Int8) MapSeq(f func(int8) int8) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Int8) MapIndexSeq(f func(int) int8) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Int8) MapIndexValSeq(f func(int, int8) int8) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Int8) Map(f func(int8) int8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkInt8(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualInt8(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int8.Fill helper
func mapChunkInt8(f func(int8) int8, wg *sync.WaitGroup, s *[]int8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int8.Fill helper
func mapResidualInt8(f func(int8) int8, wg *sync.WaitGroup, s *[]int8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Int8) MapIndex(f func(int) int8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkInt8(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualInt8(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int8.Fill helper
func mapIndexChunkInt8(f func(int) int8, wg *sync.WaitGroup, s *[]int8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int8.Fill helper
func mapIndexResidualInt8(f func(int) int8, wg *sync.WaitGroup, s *[]int8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Int8) MapIndexVal(f func(int, int8) int8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkInt8(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualInt8(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int8.Fill helper
func mapIndexValChunkInt8(f func(int, int8) int8, wg *sync.WaitGroup, s *[]int8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int8.Fill helper
func mapIndexValResidualInt8(f func(int, int8) int8, wg *sync.WaitGroup, s *[]int8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Int8) FillSeq(val int8) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Int8) Fill(val int8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkInt8(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualInt8(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int8.Fill helper
func fillChunkInt8(val int8, wg *sync.WaitGroup, s *[]int8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int8.Fill helper
func fillResidualInt8(val int8, wg *sync.WaitGroup, s *[]int8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Int8) At(indices ...int) *int8 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int8) Get(indices ...int) int8 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Int8) Get(k int) int8 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int8) Set(val int8, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Int8) Set(val int8, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Int8) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Int8 makes a new array with element type `int8`.
func (af *ArrayFactory) Int8() (*Int8, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Int8), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Int8), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Int8{
		Metadata: af.Metadata,
		Data:     make([]int8, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Int8
func (A *Int8) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Int8
// func (A *Int8) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Uint8 is a concrete array type whose elements are of type `uint8`. 
// Uint8 is defined by composition of `Metadata` and the slice `Data`.
type Uint8 struct {
	Metadata
	Data []uint8
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Uint8) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Uint8.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Uint8) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseUint(record[j], 10, 8); A.Data[k] = uint8(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Uint8) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Uint8.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Uint8.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Uint8.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Uint8) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Uint8.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Uint8) View(slice Slice) (*ViewUint8, error) {

	view := &ViewUint8{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Uint8.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Uint8) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Uint8) MapSeq(f func(uint8) uint8) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Uint8) MapIndexSeq(f func(int) uint8) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint8) MapIndexValSeq(f func(int, uint8) uint8) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Uint8) Map(f func(uint8) uint8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkUint8(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualUint8(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint8.Fill helper
func mapChunkUint8(f func(uint8) uint8, wg *sync.WaitGroup, s *[]uint8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint8.Fill helper
func mapResidualUint8(f func(uint8) uint8, wg *sync.WaitGroup, s *[]uint8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Uint8) MapIndex(f func(int) uint8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkUint8(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualUint8(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint8.Fill helper
func mapIndexChunkUint8(f func(int) uint8, wg *sync.WaitGroup, s *[]uint8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint8.Fill helper
func mapIndexResidualUint8(f func(int) uint8, wg *sync.WaitGroup, s *[]uint8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint8) MapIndexVal(f func(int, uint8) uint8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkUint8(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualUint8(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint8.Fill helper
func mapIndexValChunkUint8(f func(int, uint8) uint8, wg *sync.WaitGroup, s *[]uint8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint8.Fill helper
func mapIndexValResidualUint8(f func(int, uint8) uint8, wg *sync.WaitGroup, s *[]uint8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Uint8) FillSeq(val uint8) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Uint8) Fill(val uint8, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkUint8(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualUint8(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint8.Fill helper
func fillChunkUint8(val uint8, wg *sync.WaitGroup, s *[]uint8, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint8.Fill helper
func fillResidualUint8(val uint8, wg *sync.WaitGroup, s *[]uint8, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Uint8) At(indices ...int) *uint8 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint8) Get(indices ...int) uint8 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Uint8) Get(k int) uint8 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint8) Set(val uint8, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Uint8) Set(val uint8, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Uint8) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Uint8 makes a new array with element type `uint8`.
func (af *ArrayFactory) Uint8() (*Uint8, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Uint8), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Uint8), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Uint8{
		Metadata: af.Metadata,
		Data:     make([]uint8, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Uint8
func (A *Uint8) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Uint8
// func (A *Uint8) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Int16 is a concrete array type whose elements are of type `int16`. 
// Int16 is defined by composition of `Metadata` and the slice `Data`.
type Int16 struct {
	Metadata
	Data []int16
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Int16) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Int16.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Int16) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseInt(record[j], 10, 16); A.Data[k] = int16(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Int16) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Int16.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Int16.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Int16.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Int16) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Int16.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Int16) View(slice Slice) (*ViewInt16, error) {

	view := &ViewInt16{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Int16.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Int16) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Int16) MapSeq(f func(int16) int16) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Int16) MapIndexSeq(f func(int) int16) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Int16) MapIndexValSeq(f func(int, int16) int16) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Int16) Map(f func(int16) int16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkInt16(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualInt16(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int16.Fill helper
func mapChunkInt16(f func(int16) int16, wg *sync.WaitGroup, s *[]int16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int16.Fill helper
func mapResidualInt16(f func(int16) int16, wg *sync.WaitGroup, s *[]int16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Int16) MapIndex(f func(int) int16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkInt16(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualInt16(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int16.Fill helper
func mapIndexChunkInt16(f func(int) int16, wg *sync.WaitGroup, s *[]int16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int16.Fill helper
func mapIndexResidualInt16(f func(int) int16, wg *sync.WaitGroup, s *[]int16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Int16) MapIndexVal(f func(int, int16) int16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkInt16(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualInt16(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int16.Fill helper
func mapIndexValChunkInt16(f func(int, int16) int16, wg *sync.WaitGroup, s *[]int16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int16.Fill helper
func mapIndexValResidualInt16(f func(int, int16) int16, wg *sync.WaitGroup, s *[]int16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Int16) FillSeq(val int16) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Int16) Fill(val int16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkInt16(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualInt16(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int16.Fill helper
func fillChunkInt16(val int16, wg *sync.WaitGroup, s *[]int16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int16.Fill helper
func fillResidualInt16(val int16, wg *sync.WaitGroup, s *[]int16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Int16) At(indices ...int) *int16 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int16) Get(indices ...int) int16 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Int16) Get(k int) int16 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int16) Set(val int16, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Int16) Set(val int16, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Int16) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Int16 makes a new array with element type `int16`.
func (af *ArrayFactory) Int16() (*Int16, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Int16), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Int16), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Int16{
		Metadata: af.Metadata,
		Data:     make([]int16, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Int16
func (A *Int16) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Int16
// func (A *Int16) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Uint16 is a concrete array type whose elements are of type `uint16`. 
// Uint16 is defined by composition of `Metadata` and the slice `Data`.
type Uint16 struct {
	Metadata
	Data []uint16
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Uint16) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Uint16.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Uint16) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseUint(record[j], 10, 16); A.Data[k] = uint16(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Uint16) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Uint16.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Uint16.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Uint16.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Uint16) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Uint16.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Uint16) View(slice Slice) (*ViewUint16, error) {

	view := &ViewUint16{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Uint16.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Uint16) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Uint16) MapSeq(f func(uint16) uint16) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Uint16) MapIndexSeq(f func(int) uint16) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint16) MapIndexValSeq(f func(int, uint16) uint16) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Uint16) Map(f func(uint16) uint16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkUint16(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualUint16(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint16.Fill helper
func mapChunkUint16(f func(uint16) uint16, wg *sync.WaitGroup, s *[]uint16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint16.Fill helper
func mapResidualUint16(f func(uint16) uint16, wg *sync.WaitGroup, s *[]uint16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Uint16) MapIndex(f func(int) uint16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkUint16(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualUint16(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint16.Fill helper
func mapIndexChunkUint16(f func(int) uint16, wg *sync.WaitGroup, s *[]uint16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint16.Fill helper
func mapIndexResidualUint16(f func(int) uint16, wg *sync.WaitGroup, s *[]uint16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint16) MapIndexVal(f func(int, uint16) uint16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkUint16(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualUint16(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint16.Fill helper
func mapIndexValChunkUint16(f func(int, uint16) uint16, wg *sync.WaitGroup, s *[]uint16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint16.Fill helper
func mapIndexValResidualUint16(f func(int, uint16) uint16, wg *sync.WaitGroup, s *[]uint16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Uint16) FillSeq(val uint16) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Uint16) Fill(val uint16, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkUint16(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualUint16(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint16.Fill helper
func fillChunkUint16(val uint16, wg *sync.WaitGroup, s *[]uint16, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint16.Fill helper
func fillResidualUint16(val uint16, wg *sync.WaitGroup, s *[]uint16, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Uint16) At(indices ...int) *uint16 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint16) Get(indices ...int) uint16 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Uint16) Get(k int) uint16 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint16) Set(val uint16, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Uint16) Set(val uint16, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Uint16) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Uint16 makes a new array with element type `uint16`.
func (af *ArrayFactory) Uint16() (*Uint16, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Uint16), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Uint16), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Uint16{
		Metadata: af.Metadata,
		Data:     make([]uint16, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Uint16
func (A *Uint16) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Uint16
// func (A *Uint16) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Int32 is a concrete array type whose elements are of type `int32`. 
// Int32 is defined by composition of `Metadata` and the slice `Data`.
type Int32 struct {
	Metadata
	Data []int32
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Int32) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Int32.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Int32) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseInt(record[j], 10, 32); A.Data[k] = int32(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Int32) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Int32.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Int32.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Int32.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Int32) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Int32.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Int32) View(slice Slice) (*ViewInt32, error) {

	view := &ViewInt32{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Int32.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Int32) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Int32) MapSeq(f func(int32) int32) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Int32) MapIndexSeq(f func(int) int32) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Int32) MapIndexValSeq(f func(int, int32) int32) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Int32) Map(f func(int32) int32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkInt32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualInt32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int32.Fill helper
func mapChunkInt32(f func(int32) int32, wg *sync.WaitGroup, s *[]int32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int32.Fill helper
func mapResidualInt32(f func(int32) int32, wg *sync.WaitGroup, s *[]int32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Int32) MapIndex(f func(int) int32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkInt32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualInt32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int32.Fill helper
func mapIndexChunkInt32(f func(int) int32, wg *sync.WaitGroup, s *[]int32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int32.Fill helper
func mapIndexResidualInt32(f func(int) int32, wg *sync.WaitGroup, s *[]int32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Int32) MapIndexVal(f func(int, int32) int32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkInt32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualInt32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int32.Fill helper
func mapIndexValChunkInt32(f func(int, int32) int32, wg *sync.WaitGroup, s *[]int32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int32.Fill helper
func mapIndexValResidualInt32(f func(int, int32) int32, wg *sync.WaitGroup, s *[]int32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Int32) FillSeq(val int32) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Int32) Fill(val int32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkInt32(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualInt32(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int32.Fill helper
func fillChunkInt32(val int32, wg *sync.WaitGroup, s *[]int32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int32.Fill helper
func fillResidualInt32(val int32, wg *sync.WaitGroup, s *[]int32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Int32) At(indices ...int) *int32 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int32) Get(indices ...int) int32 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Int32) Get(k int) int32 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int32) Set(val int32, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Int32) Set(val int32, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Int32) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Int32 makes a new array with element type `int32`.
func (af *ArrayFactory) Int32() (*Int32, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Int32), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Int32), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Int32{
		Metadata: af.Metadata,
		Data:     make([]int32, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Int32
func (A *Int32) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Int32
// func (A *Int32) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Uint32 is a concrete array type whose elements are of type `uint32`. 
// Uint32 is defined by composition of `Metadata` and the slice `Data`.
type Uint32 struct {
	Metadata
	Data []uint32
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Uint32) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Uint32.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Uint32) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseUint(record[j], 10, 32); A.Data[k] = uint32(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Uint32) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Uint32.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Uint32.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Uint32.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Uint32) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Uint32.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Uint32) View(slice Slice) (*ViewUint32, error) {

	view := &ViewUint32{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Uint32.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Uint32) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Uint32) MapSeq(f func(uint32) uint32) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Uint32) MapIndexSeq(f func(int) uint32) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint32) MapIndexValSeq(f func(int, uint32) uint32) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Uint32) Map(f func(uint32) uint32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkUint32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualUint32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint32.Fill helper
func mapChunkUint32(f func(uint32) uint32, wg *sync.WaitGroup, s *[]uint32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint32.Fill helper
func mapResidualUint32(f func(uint32) uint32, wg *sync.WaitGroup, s *[]uint32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Uint32) MapIndex(f func(int) uint32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkUint32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualUint32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint32.Fill helper
func mapIndexChunkUint32(f func(int) uint32, wg *sync.WaitGroup, s *[]uint32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint32.Fill helper
func mapIndexResidualUint32(f func(int) uint32, wg *sync.WaitGroup, s *[]uint32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint32) MapIndexVal(f func(int, uint32) uint32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkUint32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualUint32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint32.Fill helper
func mapIndexValChunkUint32(f func(int, uint32) uint32, wg *sync.WaitGroup, s *[]uint32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint32.Fill helper
func mapIndexValResidualUint32(f func(int, uint32) uint32, wg *sync.WaitGroup, s *[]uint32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Uint32) FillSeq(val uint32) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Uint32) Fill(val uint32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkUint32(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualUint32(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint32.Fill helper
func fillChunkUint32(val uint32, wg *sync.WaitGroup, s *[]uint32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint32.Fill helper
func fillResidualUint32(val uint32, wg *sync.WaitGroup, s *[]uint32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Uint32) At(indices ...int) *uint32 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint32) Get(indices ...int) uint32 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Uint32) Get(k int) uint32 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint32) Set(val uint32, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Uint32) Set(val uint32, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Uint32) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Uint32 makes a new array with element type `uint32`.
func (af *ArrayFactory) Uint32() (*Uint32, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Uint32), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Uint32), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Uint32{
		Metadata: af.Metadata,
		Data:     make([]uint32, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Uint32
func (A *Uint32) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Uint32
// func (A *Uint32) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Int64 is a concrete array type whose elements are of type `int64`. 
// Int64 is defined by composition of `Metadata` and the slice `Data`.
type Int64 struct {
	Metadata
	Data []int64
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Int64) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Int64.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Int64) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseInt(record[j], 10, 64); A.Data[k] = tmp
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Int64) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Int64.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Int64.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Int64.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Int64) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Int64.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Int64) View(slice Slice) (*ViewInt64, error) {

	view := &ViewInt64{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Int64.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Int64) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Int64) MapSeq(f func(int64) int64) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Int64) MapIndexSeq(f func(int) int64) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Int64) MapIndexValSeq(f func(int, int64) int64) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Int64) Map(f func(int64) int64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkInt64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualInt64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int64.Fill helper
func mapChunkInt64(f func(int64) int64, wg *sync.WaitGroup, s *[]int64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int64.Fill helper
func mapResidualInt64(f func(int64) int64, wg *sync.WaitGroup, s *[]int64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Int64) MapIndex(f func(int) int64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkInt64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualInt64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int64.Fill helper
func mapIndexChunkInt64(f func(int) int64, wg *sync.WaitGroup, s *[]int64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int64.Fill helper
func mapIndexResidualInt64(f func(int) int64, wg *sync.WaitGroup, s *[]int64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Int64) MapIndexVal(f func(int, int64) int64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkInt64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualInt64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int64.Fill helper
func mapIndexValChunkInt64(f func(int, int64) int64, wg *sync.WaitGroup, s *[]int64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int64.Fill helper
func mapIndexValResidualInt64(f func(int, int64) int64, wg *sync.WaitGroup, s *[]int64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Int64) FillSeq(val int64) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Int64) Fill(val int64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkInt64(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualInt64(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int64.Fill helper
func fillChunkInt64(val int64, wg *sync.WaitGroup, s *[]int64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int64.Fill helper
func fillResidualInt64(val int64, wg *sync.WaitGroup, s *[]int64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Int64) At(indices ...int) *int64 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int64) Get(indices ...int) int64 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Int64) Get(k int) int64 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int64) Set(val int64, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Int64) Set(val int64, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Int64) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Int64 makes a new array with element type `int64`.
func (af *ArrayFactory) Int64() (*Int64, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Int64), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Int64), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Int64{
		Metadata: af.Metadata,
		Data:     make([]int64, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Int64
func (A *Int64) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Int64
// func (A *Int64) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Uint64 is a concrete array type whose elements are of type `uint64`. 
// Uint64 is defined by composition of `Metadata` and the slice `Data`.
type Uint64 struct {
	Metadata
	Data []uint64
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Uint64) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Uint64.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Uint64) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseUint(record[j], 10, 64); A.Data[k] = uint64(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Uint64) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Uint64.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Uint64.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Uint64.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Uint64) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Uint64.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Uint64) View(slice Slice) (*ViewUint64, error) {

	view := &ViewUint64{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Uint64.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Uint64) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Uint64) MapSeq(f func(uint64) uint64) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Uint64) MapIndexSeq(f func(int) uint64) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint64) MapIndexValSeq(f func(int, uint64) uint64) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Uint64) Map(f func(uint64) uint64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkUint64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualUint64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint64.Fill helper
func mapChunkUint64(f func(uint64) uint64, wg *sync.WaitGroup, s *[]uint64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint64.Fill helper
func mapResidualUint64(f func(uint64) uint64, wg *sync.WaitGroup, s *[]uint64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Uint64) MapIndex(f func(int) uint64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkUint64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualUint64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint64.Fill helper
func mapIndexChunkUint64(f func(int) uint64, wg *sync.WaitGroup, s *[]uint64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint64.Fill helper
func mapIndexResidualUint64(f func(int) uint64, wg *sync.WaitGroup, s *[]uint64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint64) MapIndexVal(f func(int, uint64) uint64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkUint64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualUint64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint64.Fill helper
func mapIndexValChunkUint64(f func(int, uint64) uint64, wg *sync.WaitGroup, s *[]uint64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint64.Fill helper
func mapIndexValResidualUint64(f func(int, uint64) uint64, wg *sync.WaitGroup, s *[]uint64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Uint64) FillSeq(val uint64) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Uint64) Fill(val uint64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkUint64(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualUint64(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint64.Fill helper
func fillChunkUint64(val uint64, wg *sync.WaitGroup, s *[]uint64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint64.Fill helper
func fillResidualUint64(val uint64, wg *sync.WaitGroup, s *[]uint64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Uint64) At(indices ...int) *uint64 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint64) Get(indices ...int) uint64 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Uint64) Get(k int) uint64 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint64) Set(val uint64, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Uint64) Set(val uint64, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Uint64) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Uint64 makes a new array with element type `uint64`.
func (af *ArrayFactory) Uint64() (*Uint64, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Uint64), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Uint64), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Uint64{
		Metadata: af.Metadata,
		Data:     make([]uint64, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Uint64
func (A *Uint64) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Uint64
// func (A *Uint64) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Int is a concrete array type whose elements are of type `int`. 
// Int is defined by composition of `Metadata` and the slice `Data`.
type Int struct {
	Metadata
	Data []int
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Int) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Int.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Int) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseInt(record[j], 10, 0); A.Data[k] = int(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Int) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Int.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Int.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Int.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Int) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Int.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Int) View(slice Slice) (*ViewInt, error) {

	view := &ViewInt{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Int.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Int) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Int) MapSeq(f func(int) int) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Int) MapIndexSeq(f func(int) int) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Int) MapIndexValSeq(f func(int, int) int) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Int) Map(f func(int) int, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkInt(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualInt(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int.Fill helper
func mapChunkInt(f func(int) int, wg *sync.WaitGroup, s *[]int, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int.Fill helper
func mapResidualInt(f func(int) int, wg *sync.WaitGroup, s *[]int, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Int) MapIndex(f func(int) int, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkInt(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualInt(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int.Fill helper
func mapIndexChunkInt(f func(int) int, wg *sync.WaitGroup, s *[]int, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int.Fill helper
func mapIndexResidualInt(f func(int) int, wg *sync.WaitGroup, s *[]int, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Int) MapIndexVal(f func(int, int) int, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkInt(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualInt(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int.Fill helper
func mapIndexValChunkInt(f func(int, int) int, wg *sync.WaitGroup, s *[]int, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int.Fill helper
func mapIndexValResidualInt(f func(int, int) int, wg *sync.WaitGroup, s *[]int, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Int) FillSeq(val int) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Int) Fill(val int, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkInt(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualInt(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Int.Fill helper
func fillChunkInt(val int, wg *sync.WaitGroup, s *[]int, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Int.Fill helper
func fillResidualInt(val int, wg *sync.WaitGroup, s *[]int, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Int) At(indices ...int) *int {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int) Get(indices ...int) int {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Int) Get(k int) int {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Int) Set(val int, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Int) Set(val int, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Int) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Int makes a new array with element type `int`.
func (af *ArrayFactory) Int() (*Int, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Int), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Int), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Int{
		Metadata: af.Metadata,
		Data:     make([]int, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Int
func (A *Int) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Int
// func (A *Int) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Uint is a concrete array type whose elements are of type `uint`. 
// Uint is defined by composition of `Metadata` and the slice `Data`.
type Uint struct {
	Metadata
	Data []uint
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Uint) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Uint.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Uint) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseUint(record[j], 10, 0); A.Data[k] = uint(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Uint) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Uint.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Uint.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Uint.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Uint) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Uint.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Uint) View(slice Slice) (*ViewUint, error) {

	view := &ViewUint{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Uint.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Uint) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Uint) MapSeq(f func(uint) uint) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Uint) MapIndexSeq(f func(int) uint) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint) MapIndexValSeq(f func(int, uint) uint) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Uint) Map(f func(uint) uint, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkUint(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualUint(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint.Fill helper
func mapChunkUint(f func(uint) uint, wg *sync.WaitGroup, s *[]uint, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint.Fill helper
func mapResidualUint(f func(uint) uint, wg *sync.WaitGroup, s *[]uint, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Uint) MapIndex(f func(int) uint, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkUint(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualUint(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint.Fill helper
func mapIndexChunkUint(f func(int) uint, wg *sync.WaitGroup, s *[]uint, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint.Fill helper
func mapIndexResidualUint(f func(int) uint, wg *sync.WaitGroup, s *[]uint, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Uint) MapIndexVal(f func(int, uint) uint, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkUint(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualUint(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint.Fill helper
func mapIndexValChunkUint(f func(int, uint) uint, wg *sync.WaitGroup, s *[]uint, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint.Fill helper
func mapIndexValResidualUint(f func(int, uint) uint, wg *sync.WaitGroup, s *[]uint, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Uint) FillSeq(val uint) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Uint) Fill(val uint, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkUint(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualUint(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uint.Fill helper
func fillChunkUint(val uint, wg *sync.WaitGroup, s *[]uint, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uint.Fill helper
func fillResidualUint(val uint, wg *sync.WaitGroup, s *[]uint, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Uint) At(indices ...int) *uint {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint) Get(indices ...int) uint {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Uint) Get(k int) uint {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uint) Set(val uint, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Uint) Set(val uint, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Uint) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Uint makes a new array with element type `uint`.
func (af *ArrayFactory) Uint() (*Uint, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Uint), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Uint), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Uint{
		Metadata: af.Metadata,
		Data:     make([]uint, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Uint
func (A *Uint) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Uint
// func (A *Uint) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Uintptr is a concrete array type whose elements are of type `uintptr`. 
// Uintptr is defined by composition of `Metadata` and the slice `Data`.
type Uintptr struct {
	Metadata
	Data []uintptr
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Uintptr) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Uintptr.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Uintptr) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseUint(record[j], 10, 64); A.Data[k] = uintptr(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Uintptr) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Uintptr.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Uintptr.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Uintptr.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Uintptr) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Uintptr.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Uintptr) View(slice Slice) (*ViewUintptr, error) {

	view := &ViewUintptr{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Uintptr.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Uintptr) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Uintptr) MapSeq(f func(uintptr) uintptr) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Uintptr) MapIndexSeq(f func(int) uintptr) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Uintptr) MapIndexValSeq(f func(int, uintptr) uintptr) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Uintptr) Map(f func(uintptr) uintptr, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkUintptr(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualUintptr(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uintptr.Fill helper
func mapChunkUintptr(f func(uintptr) uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uintptr.Fill helper
func mapResidualUintptr(f func(uintptr) uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Uintptr) MapIndex(f func(int) uintptr, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkUintptr(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualUintptr(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uintptr.Fill helper
func mapIndexChunkUintptr(f func(int) uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uintptr.Fill helper
func mapIndexResidualUintptr(f func(int) uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Uintptr) MapIndexVal(f func(int, uintptr) uintptr, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkUintptr(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualUintptr(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uintptr.Fill helper
func mapIndexValChunkUintptr(f func(int, uintptr) uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uintptr.Fill helper
func mapIndexValResidualUintptr(f func(int, uintptr) uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Uintptr) FillSeq(val uintptr) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Uintptr) Fill(val uintptr, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkUintptr(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualUintptr(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Uintptr.Fill helper
func fillChunkUintptr(val uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Uintptr.Fill helper
func fillResidualUintptr(val uintptr, wg *sync.WaitGroup, s *[]uintptr, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Uintptr) At(indices ...int) *uintptr {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uintptr) Get(indices ...int) uintptr {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Uintptr) Get(k int) uintptr {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Uintptr) Set(val uintptr, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Uintptr) Set(val uintptr, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Uintptr) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Uintptr makes a new array with element type `uintptr`.
func (af *ArrayFactory) Uintptr() (*Uintptr, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Uintptr), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Uintptr), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Uintptr{
		Metadata: af.Metadata,
		Data:     make([]uintptr, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Uintptr
func (A *Uintptr) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbInts + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Uintptr
// func (A *Uintptr) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%-10d\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Float32 is a concrete array type whose elements are of type `float32`. 
// Float32 is defined by composition of `Metadata` and the slice `Data`.
type Float32 struct {
	Metadata
	Data []float32
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Float32) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Float32.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Float32) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseFloat(record[j], 32); A.Data[k] = float32(tmp)
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Float32) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Float32.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Float32.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Float32.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Float32) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Float32.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Float32) View(slice Slice) (*ViewFloat32, error) {

	view := &ViewFloat32{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Float32.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Float32) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Float32) MapSeq(f func(float32) float32) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Float32) MapIndexSeq(f func(int) float32) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Float32) MapIndexValSeq(f func(int, float32) float32) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Float32) Map(f func(float32) float32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkFloat32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualFloat32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float32.Fill helper
func mapChunkFloat32(f func(float32) float32, wg *sync.WaitGroup, s *[]float32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float32.Fill helper
func mapResidualFloat32(f func(float32) float32, wg *sync.WaitGroup, s *[]float32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Float32) MapIndex(f func(int) float32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkFloat32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualFloat32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float32.Fill helper
func mapIndexChunkFloat32(f func(int) float32, wg *sync.WaitGroup, s *[]float32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float32.Fill helper
func mapIndexResidualFloat32(f func(int) float32, wg *sync.WaitGroup, s *[]float32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Float32) MapIndexVal(f func(int, float32) float32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkFloat32(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualFloat32(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float32.Fill helper
func mapIndexValChunkFloat32(f func(int, float32) float32, wg *sync.WaitGroup, s *[]float32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float32.Fill helper
func mapIndexValResidualFloat32(f func(int, float32) float32, wg *sync.WaitGroup, s *[]float32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Float32) FillSeq(val float32) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Float32) Fill(val float32, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkFloat32(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualFloat32(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float32.Fill helper
func fillChunkFloat32(val float32, wg *sync.WaitGroup, s *[]float32, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float32.Fill helper
func fillResidualFloat32(val float32, wg *sync.WaitGroup, s *[]float32, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Float32) At(indices ...int) *float32 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Float32) Get(indices ...int) float32 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Float32) Get(k int) float32 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Float32) Set(val float32, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Float32) Set(val float32, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Float32) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Float32 makes a new array with element type `float32`.
func (af *ArrayFactory) Float32() (*Float32, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Float32), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Float32), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Float32{
		Metadata: af.Metadata,
		Data:     make([]float32, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Float32
func (A *Float32) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbFloats + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Float32
// func (A *Float32) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%10.4f\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

/* -------------------------------------------------------------------------------- */
// -- ARRAY TYPE
/* -------------------------------------------------------------------------------- */

// Float64 is a concrete array type whose elements are of type `float64`. 
// Float64 is defined by composition of `Metadata` and the slice `Data`.
type Float64 struct {
	Metadata
	Data []float64
}

/* ################################################################################ */
// -- ARRAY METHODS
/* ################################################################################ */

/* ================================================================================ */
// -- I/O
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FromCSVFile
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSVFile reads from a .csv source into an array.
// `FromCSVFile` is a wrapper around `FromCSV`; regarding the return values, see
// the documentation for the latter method. The input is the path to the .csv file
// to be read into the array `A`.
func (A *Float64) FromCSVFile(fpin string) (int, []int, []error, error) {

	// -- Data file (.csv)
	fin, err := os.OpenFile(fpin, os.O_RDONLY, 0666)
	defer fin.Close()

	if err != nil {
		return 0, []int{}, []error{}, MakeError("func Float64.FromCSVFile",
			fmt.Sprintf("Failed to open input file %s", fpin),
			err)
	}

	// -- Read from the .csv file.
	// -- 			  m int: number of fields/elements read
	// -- parseErrInd []int: indices (into the array `A`) of elements for which parsing failed.
	m, parseErrInd, parseErrors, err := A.FromCSV(fin)

	return m, parseErrInd, parseErrors, err

}

/* -------------------------------------------------------------------------------- */
// -- FromCSV
/* -------------------------------------------------------------------------------- */

// -- NOTE / TODO -- //
// -- Better to have the methods `FromCSV` and `FromCSVFile` read from a file
// -- as many elements as can fit into the existing array on which the methods
// -- are called. Methods with the same names that construct from a file an array
// -- that does not exist are better to be bound to `ArrayFactory`.
// -- Copying as many elements as can fit is consisten with the built-in copy for
// -- slices.

// FromCSV reads from a .csv source into an array.
//
//	- Input
// 		-- in: a reader from which the csv-formated data can be read.
//	- Output
//		--     int: number of element read from the input
//		--   []int: indices for which the parsing of an element failed (when reading csv data into an array, each field in the csv is a string that has be parsed into a numerical type; this is the parsing referref to).
//		-- []error: parsing errors that occur during file reading
//		--   error: nil if no errors occured, something hopefully informative otherwise.
func (A *Float64) FromCSV(in io.Reader) (int, []int, []error, error) {
	// -- Will store indices (into A.Data) of elements for which parsing failed.
	parseErrorIndices := make([]int, 0, 10)
	parseErrors := make([]error, 0, 10)
	r := csv.NewReader(in)
	var k, j int
	var record []string
	var err error
	for k < A.numels {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return k, parseErrorIndices, parseErrors, err
		}
		for (k < A.numels) && (j < len(record)) {
			// A.Data[k], err = strconv.ParseFloat(record[j], 64)
			tmp, err := strconv.ParseFloat(record[j], 64); A.Data[k] = tmp
			if err != nil {
				parseErrorIndices = append(parseErrorIndices, k)
				parseErrors = append(parseErrors, err)
			}
			k++
			j++
			// fmt.Printf("------ err = %v\n", err)
		}
		j = 0
		// fmt.Printf("---- k = %d\n", k)
	}

	return k, parseErrorIndices, parseErrors, nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSVFile
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// fpout: the path of the output file
func (A *Float64) ToCSVFile(fpout string, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Open the output file.
	/* -------------------------------------------------------------------------------- */
	// -- Use `Stat` as a means to check whether the output file exists.
	_, err := os.Stat(fpout)

	// -- If `err` is `nil`, the file must exist. Equivalently, `err` cannot
	// -- be `nil` if the file does not exist.
	if err == nil {
		// -- Deleting the existing output file
		os.Remove(fpout)
	}

	// -- Creating a new output file
	fout, err := os.OpenFile(fpout, os.O_CREATE|os.O_WRONLY, 0666)
	defer fout.Close()

	if err != nil {
		return MakeError("func Float64.ToCSVFile",
		 fmt.Sprintf("Failed to create output file %s", fpout),
		 err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Write to file on disk.
	/* -------------------------------------------------------------------------------- */
	err = A.ToCSV(fout)
	if err != nil {
		return MakeError("func Float64.ToCSVFile",
		fmt.Sprintf("Failed to write array to file %s", fpout),
		err)
	}

	/* -------------------------------------------------------------------------------- */
	// -- Anonymous struct containing the array metadata.
	/* -------------------------------------------------------------------------------- */
	// -- Create an anonymous struct that contains the same fields as A.Metdata *but*
	// -- exports them so that `json` can marshall them.
	arrayMetadata := struct {
		Dims   []int
		Ndims  int
		Numels int
		Micf   []int
	}{
		Dims:   A.Metadata.dims,
		Ndims:  A.Metadata.ndims,
		Numels: A.Metadata.numels,
		Micf:   A.Metadata.micf,
	}

	jout, err := json.Marshal(arrayMetadata)

	/* -------------------------------------------------------------------------------- */
	// -- Save the array metadata to a .json file.
	/* -------------------------------------------------------------------------------- */
	// -- First, construct the metadata file name from the array file name. The 
	// -- array filename should be of the form `name.csv`, but this is not 
	// -- guaranteed; hence the following logic that checks if `fpout` ends in 
	// -- ".csv".
	if arrayFileExtension := lastRunes(fpout, 4); arrayFileExtension == ".csv" {
		fpout = strings.ReplaceAll(fpout, ".csv", ".json")
	} else {
		// -- If the file name passed in as input to this function does not have
		// -- extension `.csv` just append the extension `.json` and hope for the
		// -- best.
		fpout = fpout + ".json"
	}

	// -- Save the metadata.
	err = ioutil.WriteFile(fpout, jout, 0644)
	if err != nil {
		return MakeError("func Float64.ToCSVFile",
		 fmt.Sprintf("Failed to write array metadata to file %s", fpout), 
		 err)
	}

	return nil

}

/* -------------------------------------------------------------------------------- */
// -- ToCSV (output to io.Writer)
/* -------------------------------------------------------------------------------- */

// ToCSV saves an array to a .csv file. It also saves a .json file with the same 
// name containing the array metadata.
// out: io.Writer to write to.
func (A *Float64) ToCSV(out io.Writer, printVerb ...string) error {

	/* -------------------------------------------------------------------------------- */
	// -- Return from the function if the array is empty and, hence, there is nothing 
	// -- to write.
	/* -------------------------------------------------------------------------------- */
	if A.numels == 0 {
		return nil
	}

	/* -------------------------------------------------------------------------------- */
	// -- Preparing for writing to the output .csv file.
	/* -------------------------------------------------------------------------------- */
	// -- Make a .csv writer
	w := csv.NewWriter(out)

	// -- Each row of the .csv file will have as many columns (or fields) as the columns of the array.
	ncol := A.dims[A.ndims-1]

	/* -------------------------------------------------------------------------------- */
	// -- Process each array element: each element is added to `record` ([]string) 
	// -- until `record` represents a full row. Then, `record` is written to the output
	// -- .csv file and it is zeroed out (record = make([]string, 0, ncol)) so that the
	// -- next row can be put into it.
	/* -------------------------------------------------------------------------------- */
	var counter Counter
	counter.Init(A.Dims(), A.Micf())
	var nextCounterDigits []int

	var row, nextRow int = 0, 0

	// -- `record` will hold each row to be written. The first field contains the
	// -- coordinates of the first element of the row.
	record := make([]string, 0, ncol)

	for k := range A.Data {

		if len(printVerb) > 0 {
			record = append(record, fmt.Sprintf(printVerb[0], A.Data[k]))
		} else {
			record = append(record, fmt.Sprintf("%v", A.Data[k]))
		}

		counter.Set(k)
		nextCounterDigits = counter.Next()

		// diff := diff(counter.digits, nextCounterDigits, counter.dims)

		row = nextRow
		nextRow = nextCounterDigits[A.ndims-2]

		// -- If the current (as in current loop iteration) of the array completes a row.
		if nextRow != row {
			// -- Write the record to the file.
			w.Write(record)
			// -- Clear the record.
			record = make([]string, 0, ncol)
			// prependRowLabel = true
		}

	}

	// -- Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return MakeError("func Float64.ToCSV",
		 fmt.Sprintf("Failed to write to %v", out), 
		 err)
	}

	return nil

}

/* ================================================================================ */
// -- VIEWS
/* ================================================================================ */

// View returns a view into the array A.
func (A *Float64) View(slice Slice) (*ViewFloat64, error) {

	view := &ViewFloat64{}
	// var err error

	// fmt.Printf("-- INSIDE FLOAT64.VIEW\n")

	// -- Check if the number of dimensions implied by the slice agrees with the number of dimensions of the array.
	if len(slice) != A.ndims {
		return view, fmt.Errorf("the implied number of dimensions of the slice does not much the number of dimensions of the array")
	}

	// fmt.Printf("-- CHECKED THE # OF DIMENIONS\n")

	// -- A view always has the same number of dimensions as the array.
	view.dims = make([]int, A.ndims)
	var err error
	for k := range slice {
		view.dims[k], err = slice[k].Numels()
		if err != nil {
			return view, fmt.Errorf("[Float64.View] error while populating the dims of a view : %s\n", err.Error())
		}
	}

	view.ndims = A.ndims
	view.numels = NumelsFromDims(view.dims)
	view.micf = MultiIndexConversionFactors(view.dims, len(view.dims))
	view.Array = A
	view.S = slice
	view.Err = nil

	return view, nil

}

/* ================================================================================ */
// -- ARRAY SLICES
/* ================================================================================ */

// Slice takes as input a string representation `s` of an array slice and returns 
// a `Slice` (which is of type []Range which is the same as [][3]int) that 
// represents the same information as the input string `s` but in numeric data 
// structure conducive for computation. For example, if s == "[0:5:2]", then we are
// asking the function `Slice` to construct a `Slice` that starts at 0, stops at 4,
// and advances with step 2. The result, in this case, will be the `Slice` { {0, 5, 2} }.
// One advantage of using the function Slice over "manually" instantiating `Slice`s
// is the the former is slightly user-friendlier.The string representation allows for
// the usual liberal specification of a slice where some information can be ommitted. 
// This is possible by accessing the metadata in the receiver of the method.
func (A *Float64) Slice(s string) (Slice, error) {
	return MakeSlice(s, A.dims)
}

/* ================================================================================ */
// -- MAP (AS IN MAP-REDUCE)
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- MAP: Sequential implementations
/* -------------------------------------------------------------------------------- */

// MapSeq applies the function `f` to each element of the array `A`.
func (A *Float64) MapSeq(f func(float64) float64) {
	for k := range A.Data {
		A.Data[k] = f(A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexSeq applies the function `f` to the linear index of each 
// element of the array `A`.
func (A *Float64) MapIndexSeq(f func(int) float64) {
	for k := range A.Data {
		A.Data[k] = f(k)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexValSeq applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexValSeq is a map function that takes two arguments: 
// the array element and its index.
func (A *Float64) MapIndexValSeq(f func(int, float64) float64) {
	for k := range A.Data {
		A.Data[k] = f(k, A.Data[k])
	}
}

/* -------------------------------------------------------------------------------- */
// -- MAP: Concurrent implementations
/* -------------------------------------------------------------------------------- */

// Map applies the function `f` to every element of the array `A`.
func (A *Float64) Map(f func(float64) float64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapChunkFloat64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapResidualFloat64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float64.Fill helper
func mapChunkFloat64(f func(float64) float64, wg *sync.WaitGroup, s *[]float64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float64.Fill helper
func mapResidualFloat64(f func(float64) float64, wg *sync.WaitGroup, s *[]float64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f((*s)[k])
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndex applies the function `f` to the linear index of each element
// of the array `A`.
func (A *Float64) MapIndex(f func(int) float64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexChunkFloat64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexResidualFloat64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float64.Fill helper
func mapIndexChunkFloat64(f func(int) float64, wg *sync.WaitGroup, s *[]float64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float64.Fill helper
func mapIndexResidualFloat64(f func(int) float64, wg *sync.WaitGroup, s *[]float64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */

// MapIndexVal applies the function `f` to all pairs (x, y), where y := linear index of 
// y in `A`. In other words, MapIndexVal is a map function that takes two arguments: 
// the array element and its index.
func (A *Float64) MapIndexVal(f func(int, float64) float64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go mapIndexValChunkFloat64(f, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go mapIndexValResidualFloat64(f, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float64.Fill helper
func mapIndexValChunkFloat64(f func(int, float64) float64, wg *sync.WaitGroup, s *[]float64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float64.Fill helper
func mapIndexValResidualFloat64(f func(int, float64) float64, wg *sync.WaitGroup, s *[]float64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = f(k, (*s)[k])
		// (*s)[k] = f(k)
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* ================================================================================ */
// -- FILL
/* ================================================================================ */

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Sequential implementation
/* -------------------------------------------------------------------------------- */

// FillSeq sets the value of every element of A equal to `val`. This is a 
// sequential implementation, hence the name. There will be a parallel one too.
func (A *Float64) FillSeq(val float64) {
	for k := range A.Data {
		A.Data[k] = val
	}
}

/* -------------------------------------------------------------------------------- */
// -- FILL AN ARRAY WITH A GIVEN VALUE: Concurrent implementation
/* -------------------------------------------------------------------------------- */

// Fill sets the value of every element of A equal to `val`. In cases where 
// concurrency can be beneficial for speed, `Fill` should be faster than `FillSeq`.
// For really small arrays, not only `Fill` will not be faster than `FillSeq`, but
// actually slower; because for tiny arrays the overhead of concurrency (although 
// minimal in Go) dominates the computation.
func (A *Float64) Fill(val float64, numGoRout ...int) { 

	// -- We want to wait for the filling of the array to finish before returning.
	var fillers sync.WaitGroup
	// -- The number of goroutines to be used.
	var ngr int = ncpu
	// -- Has the user passed the desired number of goroutines?
	if len(numGoRout) > 0 {
		ngr = numGoRout[0]
	}

	// fmt.Printf("-- Will use %d goroutines\n", ngr)

	chunkSize := A.numels / ngr
	numResidualItems := A.numels % ngr

	if chunkSize == 0 {
		ngr = A.numels
		chunkSize = 1
		numResidualItems = 0
	}

	// fmt.Printf("---- # GOROUTINES: %v\n", ngr)
	// fmt.Printf("---- chunk size = %v\n", chunkSize)
	// fmt.Printf("---- residual items = %v\n", numResidualItems)

	// -- Fill in the chunks of the slice using goroutines.
	fillers.Add(ngr)
	for k := 0; k < ngr; k++ {
		go fillChunkFloat64(val, &fillers, &(A.Data), k, chunkSize)
	}

	// -- There may be elements in the slice that will not be filled in by
	// -- the previous loop because the size of slice is not divisible by
	// -- the number of goroutines.
	if numResidualItems > 0 {
		fillers.Add(1)
		go fillResidualFloat64(val, &fillers, &(A.Data), ngr, numResidualItems)
	}

	fillers.Wait()

}

// -- Float64.Fill helper
func fillChunkFloat64(val float64, wg *sync.WaitGroup, s *[]float64, gorCount int, chunkSize int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillChunk. Using the value %v\n", val)
	start := gorCount * chunkSize
	stop := (gorCount + 1) * chunkSize
	for k := start; k < stop; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

// -- Float64.Fill helper
func fillResidualFloat64(val float64, wg *sync.WaitGroup, s *[]float64, gorCount, residualItems int) {
	defer (*wg).Done()
	// fmt.Printf("-- Inside fillResidual. Using the value %v\n", val)
	lens := len(*s)
	for k := lens - residualItems; k < lens; k++ {
		(*s)[k] = val
		// (*s)[k] = math.Exp(math.Log(float64(k + 1)))
		// (*s)[k] = float64(k + 1)
		// (*s)[k] = fmt.Sprintf("GOR #%v,", gorCount)
	}
}

/* -------------------------------------------------------------------------------- */
// -- ACCESS TO ARRAY ELEMENTS, GETTERS, SETTERS, ETC
/* -------------------------------------------------------------------------------- */

// At returns a pointer to the k-th linearly indexed element of an `Array`. It is
// the most primitive way to access elements, without checking bounds or returning
// errors.
func (A *Float64) At(indices ...int) *float64 {
	return &(A.Data[DotProdInt(A.Micf(), indices)])
}

// Get returns the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Float64) Get(indices ...int) float64 {
	return A.Data[DotProdInt(A.Micf(), indices)]
}
// func (A *Float64) Get(k int) float64 {
// 	return A.Data[k]
// }
	
// Set sets the value of a single element of an `Array`.
// Note that `A.Data[k]` can also be accessed directly.
func (A *Float64) Set(val float64, indices ...int) {
	A.Data[DotProdInt(A.Micf(), indices)] = val
}
// func (A *Float64) Set(val float64, k int) {
// 	A.Data[k] = val
// }

/* -------------------------------------------------------------------------------- */
// -- ARRAY ATTRIBUTES
/* -------------------------------------------------------------------------------- */

// Dir prints the attributes of an array. The name is inspired by pythons `dir`.
func (A *Float64) Dir(out io.Writer) {
	fmt.Fprintf(out, "%6v: %T\n%6v: %v\n%6v: %v\n%6v: %v\n%6v: %v\n",
		"type", A,
		"dims", A.dims,
		"ndims", A.ndims,
		"numels", A.numels,
		"micf", A.micf)
}

/* -------------------------------------------------------------------------------- */
// -- ARRAY CONSTRUCTORS
/* -------------------------------------------------------------------------------- */

// Float64 makes a new array with element type `float64`.
func (af *ArrayFactory) Float64() (*Float64, error) {

	_, err := af.SetMetadata()
	
	// -- Return an empty array and the error.
	if err != nil {
		return new(Float64), err
	}
	
	// -- Return an empty array and no error.
	if emptyArray(af) {
		return new(Float64), nil
	}

	// -- Make 0-initialised array of the requested type.
	A := &Float64{
		Metadata: af.Metadata,
		Data:     make([]float64, af.numels)}

	return A, nil
}

/* -------------------------------------------------------------------------------- */
// -- PRINTING
/* -------------------------------------------------------------------------------- */

// -- The following implementation of `String()` is a quick and dirty solution that 
// -- works well enough for now. Some other solution that uses a Writer and an 
// iterator is preferable.

// String prints an Float64
func (A *Float64) String() string {

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
	maxRowLabel := fmt.Sprintf("\n%s ", Indices2Str(Indices(A.dims, A.numels - A.dims[A.ndims-1]), "()", ","))
	maxLabelChars := utf8.RuneCountInString(maxRowLabel)

	// fmt.Printf("-- maxRowLabel = %v\n", maxRowLabel)
	// fmt.Printf("-- maxLabelChars = %v\n", maxLabelChars)

	var startl string = fmt.Sprintf("\n%[2]*[1]s: ", Indices2Str(Indices(A.dims, 0), "()", ","), maxLabelChars-2)

	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

	for k := range A.Data {
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
		result += fmt.Sprintf("%s%s" + PrintVerbFloats + "\t", dimChange, startl, A.Data[k])
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

// // String prints an Float64
// func (A *Float64) String() string {

// 	if A.ndims == 0 {
// 		return "[]"
// 	}

// 	var result string = ""
// 	var dimChange string = ""
// 	var ndims int = A.ndims

// 	var counter Counter
// 	counter.Init(A.Dims(), A.Micf())
// 	var prev []int

// 	var row int = 0
// 	newRow := row

// 	maxNumRowDigits := len(strconv.Itoa(A.dims[A.ndims-2]))
// 	var startl string = fmt.Sprintf("%*d: ", maxNumRowDigits, 0)
// 	// fmt.Printf("-- maxNumRowDigits = %v\n", maxNumRowDigits)
// 	// fmt.Printf("-- A.dims[A.ndims-2] = %v\n", A.dims[A.ndims-2])

// 	for k, v := range A.Data {
// 		counter.Set(k)
// 		prev = counter.Previous()
// 		diff := diff(prev, counter.digits, counter.dims)
// 		// fmt.Printf("diff = %v\n", diff)
// 		if EndsIn1Int(diff) {
// 			dimChange = fmt.Sprintf("\n\n%v\n", counter.digits)
// 		}
// 		row = newRow
// 		newRow = counter.digits[ndims-2]
// 		if newRow != row {
// 			startl = fmt.Sprintf("\n%*d: ", maxNumRowDigits, newRow)
// 		}
// 		// result += fmt.Sprintf("%s%s%-15v\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%-15.4f\t", dimChange, startl, v)
// 		// result += fmt.Sprintf("%s%s%d\t", dimChange, startl, v)
// 		result += fmt.Sprintf("%s%s%10.4f\t", dimChange, startl, v)
// 		startl = ""
// 		dimChange = ""
// 	}

// 	return result

// }

/* -------------------------------------------------------------------------------- */

