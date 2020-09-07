package main

import "fmt"
import "array"

func main() {

	/* -------------------------------------------------------------------------------- */
	// -- Constructing arrays
	/* -------------------------------------------------------------------------------- */
	// -- Make a 5x7 array A
	A, err := array.Factory().Dims(7, 5).Float64()
	if err != nil {
		fmt.Printf(err.Error())
	}
	// -- Print the array
	fmt.Printf("\n-- A --\n")
	fmt.Printf("%v\n", A)

}
