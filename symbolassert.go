// Package symbolassert asserts that symbols are defined in a package.
package symbolassert

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
)

// Provider looks up symbol names.
type Provider interface {
	// Load loads the package at the given import path
	// and makes the symbols available for lookup.
	Load(importPath string) error

	// Lookup looks up a symbol.
	// It must return nil if not found.
	Lookup(symbol string) types.Object
}

// A SymbolMap maps from a locally defined identifier to
// an identifier that is authoritative. The package name
// may be omitted.
type SymbolMap map[string]string

// Resolve resolves the symbol map to an ObjectMap.
func (m SymbolMap) Resolve(from, to Provider) (ObjectMap, error) {
	var (
		o    ObjectMap
		errb errorsBuilder
	)
	for remote, local := range m {
		objFrom := from.Lookup(remote)
		objTo := to.Lookup(local)
		switch {
		case objFrom == nil:
			errb = append(errb, &UnresolvedError{from, remote})
		case objTo == nil:
			errb = append(errb, &UnresolvedError{to, local})
		default:
			if o == nil {
				o = make(ObjectMap)
			}
			o[objFrom] = objTo
		}
	}
	return o, errb.Build()
}

// UnresolvedError is returned if Resolve can't lookup a symbol.
type UnresolvedError struct {
	Provider Provider
	Symbol   string
}

func (e *UnresolvedError) Error() string {
	return fmt.Sprintf("unresolved symbol: %s", e.Symbol)
}

// A ObjectMap maps from the locally defined symbol value to
// a value that is authoritative.
type ObjectMap map[types.Object]types.Object

// Config configures the Compare function.
type Config struct {
	SortByKey bool
}

// Compare asserts that locally defined symbols are
// defined the same as the package that authoritatively
// defines them.
// Currently the configuration parameter is ignored and
// may be ni.
func Compare(m ObjectMap, cfg *Config) error {
	if cfg != nil {
		// ignore configuration for now
		cfg = nil
	}

	var errb errorsBuilder
	for from, to := range m {
		if err := compare(from, to, cfg); err != nil {
			errb = append(errb, err)
		}
	}
	return errb.Build()
}

func compare(lhs, rhs types.Object, _ *Config) error {
	switch lhs := lhs.(type) {
	case *types.Const:
		if rhs, ok := rhs.(*types.Const); ok {
			return compareConst(lhs, rhs)
		}

	case *types.Func:
		rhs, ok := rhs.(*types.Func)
		if ok && equalFunc(lhs, rhs, false) {
			return nil
		}

	case *types.TypeName:
		if equalType(lhs.Type(), rhs.Type()) {
			return nil
		}

	default:
		panic(fmt.Sprintf("unhandled type object: %T", lhs))
	}

	return &mismatchError{lhs, rhs, "type mismatch"}
}

func compareConst(lhs, rhs *types.Const) error {
	ltyp, ok := lhs.Type().Underlying().(*types.Basic)
	if !ok {
		panic("constant is not basic type (left)")
	}
	rtyp, ok := rhs.Type().Underlying().(*types.Basic)
	if !ok {
		panic("constant is not basic type (right)")
	}

	if ltyp.Kind() != rtyp.Kind() {
		return &mismatchError{lhs, rhs, "constant type mismatch"}
	}
	if constant.Compare(lhs.Val(), token.NEQ, rhs.Val()) {
		return &mismatchError{lhs, rhs, "constant value mismatch"}
	}

	return nil
}

func equalFunc(lfn, rfn *types.Func, hasRecv bool) bool {
	lsig, ok := lfn.Type().Underlying().(*types.Signature)
	if !ok {
		panic("function has no signature (left)")
	}
	rsig, ok := rfn.Type().Underlying().(*types.Signature)
	if !ok {
		panic("function has no signature (right)")
	}

	if lsig.Variadic() != rsig.Variadic() ||
		!equalTuple(lsig.Params(), rsig.Params()) ||
		!equalTuple(lsig.Results(), rsig.Results()) {
		return false
	}

	if hasRecv {
		if lsig.Recv() == nil || rsig.Recv() == nil {
			return false
		}
	}

	return true
}

func equalType(lhs, rhs types.Type) bool {
	switch ltyp := lhs.(type) {
	case *types.Named:
		if !equalType(lhs.Underlying(), rhs.Underlying()) {
			return false
		}
		rtyp, ok := rhs.(*types.Named)
		return ok && equalMethods(ltyp, rtyp)

	case *types.Interface:
		rtyp, ok := rhs.(*types.Interface)
		return ok && equalMethods(ltyp, rtyp)

	case *types.Basic:
		rtyp, ok := rhs.(*types.Basic)
		return ok && ltyp.Kind() == rtyp.Kind()

	case *types.Struct:
		fields := ltyp.NumFields()
		rtyp, ok := rhs.(*types.Struct)
		if !ok || fields != rtyp.NumFields() {
			return false
		}
		for i := 0; i < fields; i++ {
			if ltyp.Field(i).Type().Underlying() != rtyp.Field(i).Type().Underlying() {
				return false
			}
		}

	default:
		panic(fmt.Sprintf("unhandled type: %T", lhs))
	}

	return true
}

func equalTuple(lhs, rhs *types.Tuple) bool {
	n := lhs.Len()
	if n != rhs.Len() {
		return false
	}
	for i := 0; i < n; i++ {
		if lhs.At(i).Type().Underlying() != rhs.At(i).Type().Underlying() {
			return false
		}
	}
	return true
}

type methods interface {
	NumMethods() int
	Method(i int) *types.Func
}

func equalMethods(lhs, rhs methods) bool {
	methods := lhs.NumMethods()
	if methods != rhs.NumMethods() {
		return false
	}
	for i := 0; i < methods; i++ {
		if !equalFunc(lhs.Method(i), rhs.Method(i), true) {
			return false
		}
	}
	return true
}

type mismatchError struct {
	from types.Object
	to   types.Object
	msg  string
}

func (e *mismatchError) Error() string {
	return fmt.Sprintf("%s (%v -> %v)", e.msg, e.from, e.to)
}

type errorsBuilder []error

func (b errorsBuilder) Build() error {
	if len(b) > 0 {
		return &Errors{Errs: b}
	}
	return nil
}

// Errors is returned from SymbolMap.Resolve and Compare.
type Errors struct {
	Errs []error
}

func (e *Errors) Error() string {
	n := len(e.Errs)
	if n == 1 {
		return e.Errs[0].Error()
	}
	return fmt.Sprintf("%v (and %d more)", e.Errs[0], n)
}
