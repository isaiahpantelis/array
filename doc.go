/*
Package array implements multi-dimensional numerical arrays and convenient methods to work with them.

File Summaries

Short summaries for some of the main files of the package.

	`array\cat.go`: a Go source file generated from the text template `array\templates\cat.go`. It implements concrete array types and their methods.

	`array\cmd\codegen\codegen`: the Go script the parses the code templates in `array\templates`

	`array\cmd\codegen\config.json`: the settings used by `array\cmd\codegen\codegen` to parse the code templates in `array\templates`

	`array\cmd\codegen\configLoader.go`: part of the `codegen` package; implements a function for loading `array\cmd\codegen\config.json`

	`array\cmd\codegen\main.go`: implements the `main` function of the package `codegen`

	`array\templates\gslice.go`: a text template processed by `array\cmd\codegen\codegen` to generate the Go file `array\gslice.go`.

	`array\templates\cat.go`: a text template processed by `array\cmd\codegen\codegen` to generate the Go file `array\cat.go`.

File name key

Explanation of some file names to partially alleviate the mental burden of remembering what each file contains.

	`cat.go`: (c)oncrete (a)rray (t)ypes

	`gslice.go`: (g)eneric slice
*/
package array

/*
NOTE: Anything after the package clause won't be part of the documentation served by `godoc`.
*/
