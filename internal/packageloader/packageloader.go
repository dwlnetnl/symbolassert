// Package packageloader provides helpers for loading packages.
package packageloader

import (
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

// A Cache caches loaded packages for a given package path.
type Cache struct {
	m  map[string][]*cacheEntry
	mu sync.Mutex
}

type cacheEntry struct {
	mode packages.LoadMode
	tags []string
	pkg  *packages.Package
}

// Put stores a loaded package pkg for a given configuration
// and package path.
func (c *Cache) Put(cfg *packages.Config, path string, pkg *packages.Package) {
	e := &cacheEntry{
		mode: cfg.Mode,
		tags: buildTags(cfg),
		pkg:  pkg,
	}

	// fmt.Printf("Put: mode=%s tags=%v path=%s -> %s",
	// 	e.mode, e.tags, path, spew.Sdump(pkg))

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(map[string][]*cacheEntry)
	}

	// reuse existing package if equal
	reused := false
lookup:
	for cachepath, entries := range c.m {
		for _, cached := range entries {
			if e.isEqual(cached) && path != cachepath {
				e.pkg = cached.pkg
				break lookup
			}
			if e.isEquivalent(cached) {
				if e.mode&^cached.mode == 0 {
					cached = e
				}
				reused = true
			}
		}
	}
	if reused {
		return
	}

	c.m[path] = append(c.m[path], e)
}

func (e *cacheEntry) isEqual(other *cacheEntry) bool {
	return e.pkg.ID == other.pkg.ID &&
		e.mode&other.mode == e.mode &&
		equalStrings(e.tags, other.tags)
}

func (e *cacheEntry) isEquivalent(other *cacheEntry) bool {
	return e.pkg.ID == other.pkg.ID &&
		e.mode|other.mode == e.mode &&
		equalStrings(e.tags, other.tags)
}

// Get loads a package from the cache for a given config and package.
// It returns nil if package is not found.
func (c *Cache) Get(cfg *packages.Config, path string) (pkg *packages.Package) {
	tags := buildTags(cfg)
	mode := cfg.Mode

	// defer func(p **packages.Package) {
	// 	spew.Printf("Get: mode=%s tags=%v path=%s -> %s",
	// 		mode, tags, path, spew.Sdump(*p))
	// }(&pkg)

	c.mu.Lock()
	defer c.mu.Unlock()

	entries := c.m[path]
	if entries == nil {
		return nil
	}
	for _, e := range entries {
		if e.mode&mode != mode {
			continue
		}
		if !equalStrings(e.tags, tags) {
			continue
		}
		return e.pkg
	}

	return nil
}

func buildTags(cfg *packages.Config) (tags []string) {
	for _, flag := range cfg.BuildFlags {
		const prefix = "-tags="
		if !strings.HasPrefix(flag, prefix) {
			continue
		}
		tags = strings.Split(flag[len(prefix):], ",")
		sort.Strings(tags)
		break
	}
	return tags
}

func equalStrings(lhs, rhs []string) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	for i := 0; i < len(lhs); i++ {
		if lhs[i] != rhs[i] {
			return false
		}
	}
	return true
}
