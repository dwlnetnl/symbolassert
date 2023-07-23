package symbolassert

import (
	"testing"
)

func newTestFileProvider(t *testing.T, importPath string) Provider {
	t.Helper()
	p, err := FileProvider(importPath, []string{
		"./internal/remotepkg/alias.go",
		"./internal/remotepkg/consts.go",
		"./internal/remotepkg/funcs_linux_amd64.go",
		"./internal/remotepkg/types_linux.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFileProvider(t *testing.T) {
	for _, c := range []struct {
		name, importPath, path string
	}{
		{"FullPkgPath/Unqualified", remotepkgFullPkgPath, ""},
		{"FullPkgPath/Qualified", remotepkgFullPkgPath, "remotepkg"},
		{"FullPkgPath/FullPkgPath", remotepkgFullPkgPath, remotepkgFullPkgPath},
		{"LocalImport/Unqualified", remotepkgLocalImport, ""},
		{"LocalImport/Qualified", remotepkgLocalImport, "remotepkg"},
		{"LocalImport/FullPkgPath", remotepkgLocalImport, remotepkgFullPkgPath},
		{"LocalImport/LocalImport", remotepkgLocalImport, remotepkgLocalImport},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := newTestFileProvider(t, c.importPath)

			if err := p.Load(""); err == nil {
				t.Error("expect error on empty path")
			}
			if err := p.Load("remotepkg"); err != nil {
				t.Error(err)
			}
			if err := p.Load(c.importPath); err != nil {
				t.Error(err)
			}

			testLookup(t, p, c.path, "linux", "amd64")
		})
	}

	t.Run("FullPkgPath/LocalImport", func(t *testing.T) {
		p := newTestFileProvider(t, remotepkgFullPkgPath)

		if err := p.Load(""); err == nil {
			t.Error("expect error on empty path")
		}
		if err := p.Load("remotepkg"); err != nil {
			t.Error(err)
		}
		if err := p.Load(remotepkgFullPkgPath); err != nil {
			t.Error(err)
		}
		if obj := p.Lookup(remotepkgLocalImport + ".Bool"); obj != nil {
			t.Errorf("expect failed lookup, got: %v", obj)
		}

		if err := p.Load(remotepkgLocalImport); err != nil {
			t.Error(err)
		}
		if obj := p.Lookup(remotepkgLocalImport + ".Bool"); obj == nil {
			t.Error(remotepkgLocalImport + ".Bool not found")
		}
	})

	t.Run("Unloaded", func(t *testing.T) {
		p := newTestFileProvider(t, remotepkgFullPkgPath)
		testLookup(t, p, remotepkgFullPkgPath, "linux", "amd64")
	})

	t.Run("InvalidPkg", func(t *testing.T) {
		p, err := FileProvider("invalid/package/path", []string{
			"./internal/remotepkg/alias.go",
			"./internal/remotepkg/consts.go",
			"./internal/remotepkg/funcs_linux_amd64.go",
			"./internal/remotepkg/types_linux.go",
		})
		if p != nil {
			t.Errorf("expect no provider, got: %#v", p)
		}
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("DiffPkg", func(t *testing.T) {
		p := newTestFileProvider(t, remotepkgFullPkgPath)

		if err := p.Load("./internal/localpkg"); err == nil {
			t.Error("expect error loading different package")
		}
	})
}
