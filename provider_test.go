package symbolassert

import (
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

var remotepkgConsts = []struct {
	name       string
	underlying string
}{
	{"ConstUntypedBool", "untyped bool"},
	{"ConstBool", "bool"},
	{"ConstUntypedRune", "untyped rune"},
	{"ConstRune", "rune"},
	{"ConstUntypedInt", "untyped int"},
	{"ConstInt", "int"},
	{"ConstInt64", "int64"},
	{"ConstUint", "uint"},
	{"ConstUint64", "uint64"},
	{"ConstUntypedFloat", "untyped float"},
	{"ConstFloat32", "float32"},
	{"ConstFloat64", "float64"},
	{"ConstUntypedComplex", "untyped complex"},
	{"ConstComplex64", "complex64"},
	{"ConstComplex128", "complex128"},
	{"ConstUntypedString", "untyped string"},
	{"ConstString", "string"},
}

var remotepkgTypes = []struct {
	name       string
	underlying string
}{
	{"Bool", "bool"},
	{"Rune", "rune"},
	{"Int", "int"},
	{"Int64", "int64"},
	{"Uint", "uint"},
	{"Uint64", "uint64"},
	{"Float32", "float32"},
	{"Float64", "float64"},
	{"Complex64", "complex64"},
	{"Complex128", "complex128"},
	{"String", "string"},
	{"Struct", "struct{" +
		"Bool bool; " +
		"Rune rune; " +
		"Int int; " +
		"Int64 int64; " +
		"Uint uint; " +
		"Uint64 uint64; " +
		"Float32 float32; " +
		"Float64 float64; " +
		"Complex64 complex64; " +
		"Complex128 complex128; " +
		"String string" +
		"}"},
	{"Interface", "interface{Method()}"},
}

func testLookup(t *testing.T, p Provider, path string, goos, goarch string) {
	t.Parallel()

	for _, tt := range remotepkgConsts {
		t.Run(tt.name, func(t *testing.T) {
			testLookupSymbol(t, p, path, tt.name, tt.underlying)
		})
	}

	testLookupAliases(t, p, path)

	if goos == "linux" {
		for _, tt := range remotepkgTypes {
			t.Run(tt.name, func(t *testing.T) {
				testLookupSymbol(t, p, path, tt.name, tt.underlying)
			})
		}
	}
	if goos == "linux" && goarch == "amd64" {
		testLookupFuncs(t, p, path)
	}
}

func testLookupSymbol(t *testing.T, p Provider, path, name, underlying string) types.Object {
	t.Parallel()

	symbol := path
	if symbol != "" {
		symbol += "."
	}
	symbol += name

	obj := p.Lookup(symbol)
	if obj == nil {
		t.Fatalf("%s is undefined", symbol)
	}
	if obj.Name() != name {
		t.Error("unexpected type name:", obj.Name())
	}
	if !strings.HasSuffix(obj.Pkg().Name(), "remotepkg") {
		t.Error("unexpected package name:", obj.Pkg().Name())
	}
	if !strings.HasSuffix(obj.Pkg().Path(), "internal/remotepkg") {
		t.Error("unexpected package path:", obj.Pkg().Path())
	}
	if obj.Type().Underlying().String() != underlying {
		t.Error("unexpected underlying type:", obj.Type().Underlying())
	}

	return obj
}

func testLookupFuncs(t *testing.T, p Provider, path string) {
	const underlying = "func(" +
		"b github.com/dwlnetnl/symbolassert/internal/remotepkg.Bool, " +
		"r github.com/dwlnetnl/symbolassert/internal/remotepkg.Rune, " +
		"i github.com/dwlnetnl/symbolassert/internal/remotepkg.Int" +
		") " +
		"error"

	t.Run("Func", func(t *testing.T) {
		testLookupSymbol(t, p, path, "Func", underlying)
	})

	t.Run("Method", func(t *testing.T) {
		obj := testLookupSymbol(t, p, path, "Method", "struct{}")

		n := obj.Type().(*types.Named)
		if n.NumMethods() != 1 {
			t.Fatal("expect method on struct Method")
		}
		m := n.Method(0)
		if !strings.HasSuffix(m.Pkg().Name(), "remotepkg") {
			t.Error("unexpected package name:", m.Pkg().Name())
		}
		if !strings.HasSuffix(m.Pkg().Path(), "internal/remotepkg") {
			t.Error("unexpected package path:", m.Pkg().Path())
		}
		if m.Type().Underlying().String() != underlying {
			t.Error("unexpected underlying type:", m.Type().Underlying())
		}
	})
}

func testLookupAliases(t *testing.T, p Provider, path string) {
	cfg := &packages.Config{Mode: packages.NeedTypes}
	binaryPkg, err := loadPackage(cfg, "encoding/binary")
	if err != nil {
		t.Fatal(err)
	}
	lookupBinaryPkg := binaryPkg.Types.Scope().Lookup

	t.Run("AliasType", func(t *testing.T) {
		obj := testLookupSymbol(t, p, path, "AliasType", "interface{"+
			"PutUint16([]byte, uint16); "+
			"PutUint32([]byte, uint32); "+
			"PutUint64([]byte, uint64); "+
			"String() string; "+
			"Uint16([]byte) uint16; "+
			"Uint32([]byte) uint32; "+
			"Uint64([]byte) uint64"+
			"}")

		if obj.Type().String() != "encoding/binary.ByteOrder" {
			t.Error("unexpected type:", obj.Type())
		}
	})
	t.Run("AliasConst", func(t *testing.T) {
		obj := testLookupSymbol(t, p, path, "AliasConst", "untyped int")

		c := lookupBinaryPkg("MaxVarintLen64")
		if c == nil {
			t.Fatal("package encoding/binary doesn't declare MaxVarintLen64")
		}

		cobj := obj.(*types.Const)
		cdef := c.(*types.Const)
		if cobj.Type() != cdef.Type() {
			t.Errorf("type mismatch; got %v, want: %v", cobj.Type(), cdef.Type())
		}
		if cobj.Val() != cdef.Val() {
			t.Errorf("val mismatch; got %v, want: %v", cobj.Val(), cdef.Val())
		}
	})
	t.Run("AliasVar", func(t *testing.T) {
		obj := testLookupSymbol(t, p, path, "AliasVar", "struct{}")

		v := lookupBinaryPkg("LittleEndian")
		if v == nil {
			t.Fatal("package encoding/binary doesn't declare LittleEndian")
		}

		vobj := obj.(*types.Var)
		vdef := v.(*types.Var)
		tnobj := vobj.Type().(*types.Named).Obj()
		tndef := vdef.Type().(*types.Named).Obj()
		if tnobj.Id() != tndef.Id() {
			t.Errorf("got %v, want: %v", tnobj.Id(), tndef.Id())
		}
	})
}
