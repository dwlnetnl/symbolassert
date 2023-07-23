package symbolassert

import (
	"errors"
	"fmt"
	"go/build"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

type fileProvider struct {
	files []string

	importPath string
	modulePath string
	moduleDir  string
	pkgName    string
	pkgPath    string

	pkgPaths stringSet
	scope    *types.Scope
}

var _ Provider = (*fileProvider)(nil)

// FileProvider returns a Provider that resolves symbols
// based on a set of Go source files.
func FileProvider(importPath string, files []string) (Provider, error) {
	p := &fileProvider{
		files:      make([]string, len(files)),
		importPath: importPath,
	}
	copy(p.files, files)
	if err := p.Load(importPath); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *fileProvider) Load(path string) error {
	if path == "" {
		return errors.New("invalid package")
	}

	if p.pkgPaths == nil {
		// populate package info
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedModule,
		}
		pkg, err := loadPackage(cfg, p.importPath)
		if err != nil {
			return err
		}
		p.pkgName = pkg.Name
		p.pkgPath = pkg.PkgPath
		p.modulePath = pkg.Module.Path
		p.moduleDir = pkg.Module.Dir
		p.pkgPaths.Add(p.importPath)
		p.pkgPaths.Add(p.pkgName)
		p.pkgPaths.Add(p.pkgPath)
	}

	if !p.pkgPaths.Contains(path) {
		// update package import path cache
		if build.IsLocalImport(path) {
			cfg := &packages.Config{Mode: packages.NeedName}
			resolved, err := loadPackage(cfg, path)
			if err != nil {
				return err
			}
			if resolved.PkgPath == p.pkgPath {
				p.pkgPaths.Add(path)
				return p.loadScope()
			}
		}
		return errors.New("try to load different package")
	}

	return p.loadScope()
}

func (p *fileProvider) Lookup(symbol string) types.Object {
	pkg, name := splitAtLastDot(symbol)
	if pkg == "" {
		pkg = p.pkgName
	}
	if !p.pkgPaths.Contains(pkg) {
		return nil
	}
	return p.scope.Lookup(name)
}

func (p *fileProvider) loadScope() error {
	if p.scope != nil {
		// scope is already loaded
		return nil
	}

	// get list of build tags
	tags, err := p.buildTags()
	if err != nil {
		return err
	}

	// load package using build tags
	cfg := &packages.Config{
		Mode:       packages.NeedTypes,
		BuildFlags: buildFlags(tags),
	}
	pkg, err := loadPackage(cfg, p.importPath)
	if err != nil {
		return err
	}

	p.scope = pkg.Types.Scope()
	return nil
}

func (p *fileProvider) buildTags() ([]string, error) {
	var files, dirs stringSet
	for _, file := range p.files {
		dir, name := filepath.Split(file)
		dirs.Add(dir)
		files.Add(name)
	}
	if len(dirs) > 1 {
		return nil, errors.New("files are stored in multiple directories")
	}

	ctx := build.Context{
		Compiler:    "gc",
		UseAllFiles: true,
		ReadDir: func(dir string) ([]fs.FileInfo, error) {
			des, err := os.ReadDir(dir)
			if err != nil {
				return nil, err
			}
			fis := make([]fs.FileInfo, 0, len(p.files))
			for _, de := range des {
				if de.IsDir() {
					continue
				}
				if !files.Contains(de.Name()) {
					continue
				}
				fi, err := de.Info()
				if err != nil {
					return nil, err
				}
				fis = append(fis, fi)
			}
			return fis, nil
		},
	}

	var path string
	if build.IsLocalImport(p.importPath) {
		path = p.importPath
	} else {
		path = strings.ReplaceAll(p.importPath, p.modulePath, p.moduleDir)
	}
	pkg, err := ctx.ImportDir(path, 0)
	if err != nil {
		return nil, err
	}

	for _, file := range pkg.GoFiles {
		if !files.Contains(file) {
			return nil, fmt.Errorf("file is ignored: %s", file)
		}
	}

	return pkg.AllTags, nil
}
