package localpkg

const (
	ConstUntypedBool      = true
	ConstBool        Bool = true
)

const (
	ConstUntypedRune      = 'a'
	ConstRune        rune = 'a'
)

const (
	ConstUntypedInt        = 1
	ConstInt        int    = 1
	ConstInt64      int64  = 1
	ConstUint       uint   = 1
	ConstUint64     uint64 = 1
)

const (
	ConstUntypedFloat         = 1.
	ConstFloat32      float32 = 1.
	ConstFloat64      float64 = 1.
)

const (
	ConstUntypedComplex            = (1 + 1i)
	ConstComplex64      complex64  = (1 + 1i)
	ConstComplex128     complex128 = (1 + 1i)
)

const (
	ConstUntypedString        = "string"
	ConstString        string = "string"
)

type Struct struct {
	Bool       bool
	Rune       rune
	Int        int
	Int64      int64
	Uint       uint
	Uint64     uint64
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
	String     string
}

type (
	Bool       bool
	Rune       rune
	Int        int
	Int64      int64
	Uint       uint
	Uint64     uint64
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
	String     string
)

type Interface interface {
	Method()
}

func Func(b Bool, r Rune, i Int) error {
	return nil
}

type Method struct{}

func (Method) Method(b Bool, r Rune, i Int) error {
	return nil
}
