package remotepkg

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

type Interface interface {
	Method()
}
