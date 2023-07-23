//go:build mismatch

package localpkg

const (
	MismatchUntypedBool      = false
	MismatchBool        bool = false
)

const (
	MismatchUntypedRune      = 'b'
	MismatchRune        rune = 'b'
)

const (
	MismatchUntypedInt      = -1
	MismatchInt        int  = -1
	MismatchUint       uint = 2
)

const (
	MismatchUntypedFloat         = 2.
	MismatchFloat64      float64 = 2.
)

const (
	MismatchUntypedComplex            = (2 + 2i)
	MismatchComplex128     complex128 = (2 + 2i)
)

func (Method) MismatchMethod() {}
