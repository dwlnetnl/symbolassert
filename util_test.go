package symbolassert

import (
	"flag"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/dwlnetnl/symbolassert/internal/packageloader"
)

func init() {
	// make testing.Short work
	testing.Init()
	flag.Parse()

	if testing.Short() {
		// use package load cache to speed tests up
		cache := &packageloader.Cache{}
		loadPackageAfter = cache.Put
		loadPackageBefore = func(cfg *packages.Config, path string) *packages.Package {
			if pkg := cache.Get(cfg, path); pkg != nil {
				return pkg
			}

			// always request name and module as well
			cfg.Mode |= packages.NeedName
			cfg.Mode |= packages.NeedModule
			return nil
		}
	}
}

const (
	localpkgLocalImport  = "./internal/localpkg"
	remotepkgLocalImport = "./internal/remotepkg"
)

var remotepkgFullPkgPath = func() string {
	cfg := &packages.Config{Mode: packages.NeedName}
	pkg, err := loadPackage(cfg, remotepkgLocalImport)
	if err != nil {
		panic(err)
	}
	return pkg.PkgPath
}()

func Test_splitAtLastDot(t *testing.T) {
	cases := []struct {
		in     string
		before string
		after  string
	}{
		{"Bool", "", "Bool"},
		{"./internal/localpkg.Bool", "./internal/localpkg", "Bool"},
		{"code.example.com/localpkg.Bool", "code.example.com/localpkg", "Bool"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			before, after := splitAtLastDot(c.in)
			if before != c.before {
				t.Errorf("got before %q, want: %q", before, c.before)
			}
			if after != c.after {
				t.Errorf("got after %q, want: %q", after, c.after)
			}
		})
	}
}
