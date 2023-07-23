package symbolassert

import (
	"errors"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func mapassign(m *map[string]string, key, val string) {
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[key] = val
}

type stringSet map[string]struct{}

func (s *stringSet) Add(str string) {
	if *s == nil {
		*s = make(stringSet)
	}
	(*s)[str] = struct{}{}
}

func (s stringSet) Contains(str string) bool {
	if s == nil {
		return false
	}
	_, ok := s[str]
	return ok
}

func (s stringSet) entries() []string {
	if s == nil {
		return nil
	}
	entries := make([]string, 0, len(s))
	for entry := range s {
		entries = append(entries, entry)
	}
	return entries
}

func (s stringSet) String() string {
	entries := s.entries()
	sort.Strings(entries)
	return "{" + strings.Join(entries, " ") + "}"
}

func splitAtLastDot(s string) (before, after string) {
	if i := strings.LastIndexByte(s, '.'); i != -1 {
		return s[:i], s[i+1:]
	}
	return "", s
}

func buildFlags(tags []string) []string {
	return []string{
		"-tags=" + strings.Join(tags, ","),
	}
}

// for testing
var (
	loadPackageBefore func(cfg *packages.Config, path string) *packages.Package
	loadPackageAfter  func(cfg *packages.Config, path string, pkg *packages.Package)
)

func loadPackage(cfg *packages.Config, pattern string) (*packages.Package, error) {
	if strings.HasPrefix(pattern, "file=") || strings.HasSuffix(pattern, "...") {
		return nil, errors.New("invalid package path")
	}

	path := pattern
	if loadPackageBefore != nil {
		if pkg := loadPackageBefore(cfg, path); pkg != nil {
			return pkg, nil
		}
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, err
	}
	var visitErr error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			if visitErr == nil {
				visitErr = err
				break
			}
		}
	})
	if visitErr != nil {
		return nil, visitErr
	}

	if n := len(pkgs); n == 0 {
		return nil, errors.New("no packages loaded")
	} else if n > 1 {
		return nil, errors.New("loaded multiple packages")
	}

	pkg := pkgs[0]
	if loadPackageAfter != nil {
		loadPackageAfter(cfg, path, pkg)
	}

	return pkg, nil
}
