package symbolassert

import (
	"fmt"
	"go/build"
	"go/types"
	"os"

	"golang.org/x/tools/go/packages"
)

type PackageProvider struct {
	GOOS      string   // target operating system
	GOARCH    string   // target architecture
	BuildTags []string // build tags

	// Package is used to resolve an unqualified identifier,
	// an identifier without a package name. The caller is
	// responsible to Load this package.
	Package string

	cfg    *packages.Config
	names  map[string]string // package name resolved to package path
	local  map[string]string // local import resolved to package path
	scopes map[string]*types.Scope
}

var _ Provider = (*PackageProvider)(nil)

// Load implements the Provider interface.
func (p *PackageProvider) Load(path string) error {
	if p.scopes[path] != nil {
		// already loaded
		return nil
	}

	if p.cfg == nil {
		p.cfg = &packages.Config{
			Mode: packages.NeedName | packages.NeedTypes,
		}

		buildTags := make([]string, len(p.BuildTags))
		copy(buildTags, p.BuildTags)

		if p.GOOS != "" {
			p.cfg.Env = append(p.cfg.Env, "GOOS="+p.GOOS)
			buildTags = append(buildTags, p.GOOS)
		}
		if p.GOARCH != "" {
			p.cfg.Env = append(p.cfg.Env, "GOARCH="+p.GOARCH)
			buildTags = append(buildTags, p.GOARCH)
		}
		if GOCACHE, ok := os.LookupEnv("GOCACHE"); ok {
			p.cfg.Env = append(p.cfg.Env,
				"GOCACHE="+GOCACHE,
			)
		} else {
			HOME, _ := os.UserHomeDir()
			p.cfg.Env = append(p.cfg.Env,
				"GOCACHE=",
				"HOME="+HOME,
			)
		}

		if len(buildTags) > 0 {
			p.cfg.BuildFlags = buildFlags(buildTags)
		}
	}
	pkg, err := loadPackage(p.cfg, path)
	if err != nil {
		return err
	}

	if path := p.names[pkg.Name]; path != "" {
		return fmt.Errorf("package name conflict, already loaded: %s", path)
	}
	mapassign(&p.names, pkg.Name, pkg.PkgPath)
	if build.IsLocalImport(path) {
		mapassign(&p.local, path, pkg.PkgPath)
	}

	// resolve package for unqualified identifiers
	if p.Package != "" {
		_, resolved := p.names[p.Package]
		if !resolved && build.IsLocalImport(p.Package) {
			cfg := &packages.Config{Mode: packages.NeedName}
			resolved, err := loadPackage(cfg, p.Package)
			if err != nil {
				return err
			}
			mapassign(&p.names, p.Package, resolved.PkgPath)
			mapassign(&p.local, p.Package, resolved.PkgPath)
		}
	}

	if p.scopes == nil {
		p.scopes = make(map[string]*types.Scope)
	}

	p.scopes[pkg.PkgPath] = pkg.Types.Scope()
	return nil
}

// Lookup implements the Provider interface.
func (p *PackageProvider) Lookup(symbol string) types.Object {
	pkg, name := splitAtLastDot(symbol)
	if pkg == "" && p.Package != "" {
		pkg = p.Package
	}
	if path, ok := p.names[pkg]; ok {
		pkg = path
	} else if path, ok := p.local[pkg]; ok {
		pkg = path
	}
	if s, ok := p.scopes[pkg]; ok {
		return s.Lookup(name)
	}
	return nil
}
