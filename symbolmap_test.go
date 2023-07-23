package symbolassert

import (
	"errors"
	"go/types"
	"testing"
)

func TestSymbolMap_Resolve(t *testing.T) {
	from := &PackageProvider{Package: remotepkgLocalImport}
	if err := from.Load(remotepkgLocalImport); err != nil {
		t.Fatal(err)
	}
	to := &PackageProvider{Package: localpkgLocalImport}
	if err := to.Load(localpkgLocalImport); err != nil {
		t.Fatal(err)
	}

	symbols := SymbolMap{"Uint": "Uint"}
	objects, err := symbols.Resolve(from, to)
	if objects != nil {
		t.Errorf("got: %v, want: %v", objects, (SymbolMap)(nil))
	}
	if err == nil {
		t.Fatal("expect error")
	}

	var errs *Errors
	if !errors.As(err, &errs) {
		t.Fatalf("got %T, want: %T", err, errs)
	}
	if len(errs.Errs) != 1 {
		t.Errorf("got %d errors, want: 1", len(errs.Errs))
	}

	var got *UnresolvedError
	want := &UnresolvedError{from, "Uint"}
	if !errors.As(errs.Errs[0], &got) {
		t.Fatalf("got %T, want: %T", got, want)
	}
	if *got != *want {
		t.Errorf("got %v, want: %v", got, want)
	}
}

func TestCompare(t *testing.T) {
	from := &PackageProvider{
		GOOS:    "linux",
		GOARCH:  "amd64",
		Package: remotepkgLocalImport,
	}
	if err := from.Load(remotepkgLocalImport); err != nil {
		t.Fatal(err)
	}

	t.Run("Match", func(t *testing.T) {
		to := &PackageProvider{Package: localpkgLocalImport}
		if err := to.Load(localpkgLocalImport); err != nil {
			t.Fatal(err)
		}

		symbols := SymbolMap{
			"ConstUntypedBool":    "ConstUntypedBool",
			"ConstBool":           "ConstBool",
			"ConstUntypedRune":    "ConstUntypedRune",
			"ConstRune":           "ConstRune",
			"ConstUntypedInt":     "ConstUntypedInt",
			"ConstInt":            "ConstInt",
			"ConstInt64":          "ConstInt64",
			"ConstUint":           "ConstUint",
			"ConstUint64":         "ConstUint64",
			"ConstUntypedFloat":   "ConstUntypedFloat",
			"ConstFloat32":        "ConstFloat32",
			"ConstFloat64":        "ConstFloat64",
			"ConstUntypedComplex": "ConstUntypedComplex",
			"ConstComplex64":      "ConstComplex64",
			"ConstComplex128":     "ConstComplex128",
			"ConstUntypedString":  "ConstUntypedString",
			"ConstString":         "ConstString",
			"Bool":                "Bool",
			"Rune":                "Rune",
			"Int":                 "Int",
			"Int64":               "Int64",
			"Uint":                "Uint",
			"Uint64":              "Uint64",
			"Float32":             "Float32",
			"Float64":             "Float64",
			"Complex64":           "Complex64",
			"Complex128":          "Complex128",
			"String":              "String",
			"Struct":              "Struct",
			"Interface":           "Interface",
			"Func":                "Func",
			"Method":              "Method",
		}
		in, err := symbols.Resolve(from, to)
		if err != nil {
			t.Fatal(err)
		}
		if err := Compare(in, nil); err != nil {
			t.Error("unexpected error:", err)
		}
	})

	t.Run("Mismatch", func(t *testing.T) {
		to := &PackageProvider{
			Package:   localpkgLocalImport,
			BuildTags: []string{"mismatch"},
		}
		if err := to.Load(localpkgLocalImport); err != nil {
			t.Fatal(err)
		}

		cases := []struct {
			from string
			to   string
			err  bool
		}{
			{from: "ConstBool", to: "MismatchBool", err: true},
			{from: "ConstUntypedBool", to: "MismatchUntypedBool", err: true},
			{from: "ConstUntypedRune", to: "MismatchUntypedRune", err: true},
			{from: "ConstRune", to: "MismatchRune", err: true},
			{from: "ConstUntypedInt", to: "MismatchUntypedInt", err: true},
			{from: "ConstInt", to: "MismatchInt", err: true},
			{from: "ConstInt64", to: "ConstInt64", err: false},
			{from: "ConstUint", to: "MismatchUint", err: true},
			{from: "ConstUint64", to: "ConstUint64", err: false},
			{from: "ConstUntypedFloat", to: "MismatchUntypedFloat", err: true},
			{from: "ConstFloat32", to: "ConstFloat32", err: false},
			{from: "ConstFloat64", to: "MismatchFloat64", err: true},
			{from: "ConstUntypedComplex", to: "MismatchUntypedComplex", err: true},
			{from: "ConstComplex64", to: "ConstComplex64", err: false},
			{from: "ConstComplex128", to: "MismatchComplex128", err: true},
			{from: "ConstUntypedString", to: "ConstUntypedString", err: false},
			{from: "ConstString", to: "ConstString", err: false},
			{from: "Bool", to: "Bool", err: false},
			{from: "Rune", to: "Rune", err: false},
			{from: "Int", to: "Int", err: false},
			{from: "Int64", to: "Int64", err: false},
			{from: "Uint", to: "Uint", err: false},
			{from: "Uint64", to: "Uint64", err: false},
			{from: "Float32", to: "Float32", err: false},
			{from: "Float64", to: "Float64", err: false},
			{from: "Complex64", to: "Complex64", err: false},
			{from: "Complex128", to: "Complex128", err: false},
			{from: "String", to: "String", err: false},
			{from: "Struct", to: "Struct", err: false},
			{from: "Interface", to: "Interface", err: false},
			{from: "Func", to: "Func", err: false},
			{from: "Method", to: "Method", err: true},
		}

		// manually resolve symbols, allocate zeroed slice
		resolved := make([]struct{ from, to types.Object }, len(cases))
		objMap := make(ObjectMap, len(cases))
		mismatch := make(map[types.Object]bool)
		for i, c := range cases {
			resolved[i].from = from.Lookup(c.from)
			if resolved[i].from == nil {
				t.Fatal(&UnresolvedError{from, c.from})
			}
			resolved[i].to = to.Lookup(c.to)
			if resolved[i].to == nil {
				t.Fatal(&UnresolvedError{to, c.to})
			}
			objMap[resolved[i].from] = resolved[i].to
			mismatch[resolved[i].from] = c.err
		}

		err := Compare(objMap, nil)
		if err == nil {
			t.Error("error expected")
		}
		var errs *Errors
		if !errors.As(err, &errs) {
			t.Fatalf("error is a %T, expected %T", err, errs)
		}

		for _, e := range errs.Errs {
			var me *mismatchError
			if !errors.As(e, &me) {
				t.Errorf("error is a %T, expected %T", e, me)
				continue
			}

			delete(mismatch, me.from)
		}

		for from, unhandledErr := range mismatch {
			if unhandledErr {
				t.Errorf("got error for symbol resolved from %v", from)
			}
		}
	})
}
