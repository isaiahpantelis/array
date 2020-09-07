package array

// import "fmt"
// var _ = fmt.Printf

/* -------------------------------------------------------------------------------- */

// Counter keeps track, in `digits`, of a counter/multi-index.
type Counter struct {
	digits []int
	dims   []int
	micf   []int
	scalar int
}

// Set sets the value of a counter using a linear index.
func (c *Counter) Set(k int) {
	c.scalar = k
	copy(c.digits, Indices(c.dims, k))
}

// Init initialises a counter using a pointer to an array.
func (c *Counter) Init(dims, micf []int) {
	c.digits = make([]int, len(dims))
	c.dims = make([]int, len(dims))
	c.micf = make([]int, len(micf))
	copy(c.dims, dims)
	copy(c.micf, micf)
}

// func (c *Counter) Init(meta *Metadata) {
// 	c.digits = make([]int, len(meta.Dims()))
// 	c.dims = make([]int, len(meta.Dims()))
// 	copy(c.dims, meta.Dims())
// 	c.micf = make([]int, len(meta.Micf()))
// 	copy(c.micf, meta.Micf())
// }

// Next returns the next value of the counter.
func (c *Counter) Next() []int {
	return Indices(c.dims, c.scalar+1)
}

// Previous returns the previous value of the counter.
func (c *Counter) Previous() []int {
	return Indices(c.dims, c.scalar-1)
}

// Calculates the difference between two values of the counter.
func diff(x []int, y []int, dims []int) []int {
	result := []int{}
	var tmp int = 0
	for k := range x {
		tmp = (y[k] - x[k]) % (dims[k])
		if tmp < 0 {
			tmp += dims[k]
		}
		result = append(result, tmp)
	}
	return result
}
