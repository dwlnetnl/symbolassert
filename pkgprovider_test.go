package symbolassert

import (
	"testing"
)

func TestPackageProvider(t *testing.T) {
	for _, c := range []struct {
		name, importPath, path string
	}{
		{"FullPkgPath/Qualified", remotepkgFullPkgPath, "remotepkg"},
		{"FullPkgPath/FullPkgPath", remotepkgFullPkgPath, remotepkgFullPkgPath},
		{"LocalImport/Qualified", remotepkgLocalImport, "remotepkg"},
		{"LocalImport/LocalImport", remotepkgLocalImport, remotepkgLocalImport},
		{"LocalImport/FullPkgPath", remotepkgLocalImport, remotepkgFullPkgPath},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := &PackageProvider{}
			if err := p.Load(c.importPath); err != nil {
				t.Error(err)
			}
			testLookup(t, p, c.path, "", "")
		})
	}

	t.Run("Platform", func(t *testing.T) {
		for _, c := range []struct {
			name string
			make func() *PackageProvider
		}{
			{"GOOS_GOARCH", func() *PackageProvider {
				return &PackageProvider{GOOS: "linux", GOARCH: "amd64"}
			}},
			{"BuildTags", func() *PackageProvider {
				return &PackageProvider{BuildTags: []string{"linux", "amd64"}}
			}},
		} {
			t.Run(c.name, func(t *testing.T) {
				p := c.make()
				if err := p.Load(remotepkgFullPkgPath); err != nil {
					t.Fatal(err)
				}
				testLookup(t, p, remotepkgFullPkgPath, "linux", "amd64")
			})
		}
	})

	t.Run("ResolvePackage", func(t *testing.T) {
		for _, c := range []struct {
			name string
			make func() *PackageProvider
		}{
			{"PkgMame", func() *PackageProvider {
				return &PackageProvider{Package: "remotepkg"}
			}},
			{"LocalImport", func() *PackageProvider {
				return &PackageProvider{Package: remotepkgLocalImport}
			}},
		} {
			t.Run(c.name, func(t *testing.T) {
				p := c.make()
				if err := p.Load(remotepkgFullPkgPath); err != nil {
					t.Fatal(err)
				}
				if p.names[p.Package] != remotepkgFullPkgPath {
					t.Errorf("resolved to %q, should be %q", p.names[p.Package], remotepkgFullPkgPath)
				}
				testLookup(t, p, "", "", "")
			})
		}
	})
}
