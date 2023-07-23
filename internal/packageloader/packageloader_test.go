package packageloader

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestCache(t *testing.T) {
	type key struct {
		mode packages.LoadMode
		tags string
	}
	pkgs := make(map[key]*packages.Package)
	wantPkg := func(mode packages.LoadMode, tags string) *packages.Package {
		return pkgs[key{mode, tags}]
	}

	c := &Cache{}
	put := func(path string, mode packages.LoadMode, tags string) {
		pkg := &packages.Package{ID: "module.path/internal/remotepkg"}
		cfg := &packages.Config{Mode: mode}
		if tags != "" {
			cfg.BuildFlags = []string{"-tags=" + tags}
		}
		c.Put(cfg, path, pkg)

		if pkgs[key{mode, tags}] == nil {
			pkgs[key{mode, tags}] = pkg
		}
	}
	get := func(path string, mode packages.LoadMode, tags string) *packages.Package {
		cfg := &packages.Config{Mode: mode}
		if tags != "" {
			cfg.BuildFlags = []string{"-tags=" + tags}
		}
		return c.Get(cfg, path)
	}

	const (
		NeedName   = packages.NeedName
		NeedTypes  = packages.NeedTypes
		NeedModule = packages.NeedModule
	)

	check := func(t *testing.T, path string) {
		t.Run("Types", func(t *testing.T) {
			const mode = NeedName | NeedTypes
			t.Run("NONE", func(t *testing.T) {
				const tags = ""
				var want *packages.Package
				if got := get(path, mode, tags); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
				put(path, mode, tags)
				want = wantPkg(mode, tags)
				if got := get(path, NeedTypes, ""); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
				if got := get(path, NeedName, ""); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
			})
			t.Run("LinuxAMD64", func(t *testing.T) {
				const tags = "amd64,linux"
				var want *packages.Package
				if got := get(path, NeedTypes, tags); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
				put(path, mode, tags)
				want = wantPkg(mode, tags)
				if got := get(path, NeedTypes, tags); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
				// note: build tags are unsorted
				if got := get(path, NeedTypes, "linux,amd64"); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
				if got := get(path, NeedName, "amd64,linux"); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
				// note: build tags are unsorted
				if got := get(path, NeedName, "linux,amd64"); got != want {
					t.Errorf("got %p, want: %p", got, want)
				}
			})
		})
		t.Run("Module/NONE", func(t *testing.T) {
			const mode = NeedName | NeedModule
			const tags = ""
			var want *packages.Package
			if got := get(path, mode, tags); got != want {
				t.Errorf("got %p, want: %p", got, want)
			}
			put(path, mode, tags)
			want = wantPkg(mode, tags)
			if got := get(path, NeedModule, tags); got != want {
				t.Errorf("got %p, want: %p", got, want)
			}

			if get(path, NeedName|NeedTypes, tags) == nil {
				t.Fatal("expect NeedName|NeedTypes entry in cache")
			}
			want = wantPkg(NeedName|NeedTypes, tags)
			if got := get(path, NeedName, tags); got != want {
				t.Errorf("got %p, want: %p", got, want)
			}
		})
	}

	t.Run("LocalImport", func(t *testing.T) {
		check(t, "./internal/remotepkg")
	})
	t.Run("FullPkgPath", func(t *testing.T) {
		check(t, "module.path/internal/remotepkg")
	})

	t.Run("Update", func(t *testing.T) {
		c := &Cache{}
		put := func(path string, mode packages.LoadMode, tags string) {
			pkg := &packages.Package{ID: "module.path/internal/remotepkg"}
			cfg := &packages.Config{Mode: mode}
			if tags != "" {
				cfg.BuildFlags = []string{"-tags=" + tags}
			}
			c.Put(cfg, path, pkg)

			if pkgs[key{mode, tags}] == nil {
				pkgs[key{mode, tags}] = pkg
			}
		}

		const path = "module.path/internal/remotepkg"
		const tags = "amd64,linux"
		n := len(c.m[path])
		put(path, NeedTypes, tags)
		put(path, NeedName|NeedTypes, tags)
		put(path, NeedName|NeedTypes, tags)
		put(path, NeedTypes, tags)
		if n+1 != len(c.m[path]) {
			t.Error("cache entry is not replaced")
		}
	})
}
