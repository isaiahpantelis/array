# Contents
1. [Summary](#Summary)
1. [Introduction](#Introduction)
1. [Quickstart](#Quickstart)
    1. [Constructing arrays](##Constructing\ arrays)

# Summary
`array` implements multi-dimensional numeric arrays. 

**As of 2020-09-06:**
- `array` is a package that implements a common data structure; it is *not* a package for linear algebra.
- The implementation is in pure Go (cgo is not used).
- Arrays with [complex](https://golang.org/pkg/math/cmplx/) elements are not supported.
- Arrays of **strings** are not supported.

# Introduction
For each built-in numeric type,<sup>1</sup> there is a corresponding array type. The name of the array type is the same as the name of the type of its elements, but with the first letter capitalised. For example, the type of an array of `float64` is `Float64` and the type of an array of `uint8` is `Uint8`.

All definitions of array types follow the same pattern:

```
type Type struct {
	Metadata
	Data []type
}
```

For example, the definition of `Float64` is

```go
type Float64 struct {
	Metadata
	Data []float64
}
```
`Metadata` is a type used for the bookkeeping that makes it possible to treat a slice as an array. That is, `Metadata` is the bridge between the *mental model* of a multi-dimensional array and the *actual* storage of data in contiguous memory.

From the definition of the array types it becomes evident that, since the underlying storage of an array is a slice, arrays are "homogeneous containers"; that is, an array cannot contain elements of different types.

### Footnotes
<sup>1</sup> `int8`, `uint8`, `int16`, `uint16`, `int32`, `uint32`, `int64`, `uint64`, `int`, `uint`, `uintptr`, `float32`, `float64`

# Quickstart
## Constructing arrays
