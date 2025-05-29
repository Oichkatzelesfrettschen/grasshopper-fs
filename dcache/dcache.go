package dcache

import (
	"github.com/mit-pdos/go-journal/common"
)

// Dentry holds a cached directory entry mapping a name to an inode and offset.
type Dentry struct {
	Inum common.Inum
	Off  uint64
}

// Dcache caches directory lookups for a single directory.
type Dcache struct {
	cache   map[string]Dentry
	Lastoff uint64
}

// MkDcache creates an empty directory cache.
func MkDcache() *Dcache {
	return &Dcache{
		cache:   make(map[string]Dentry),
		Lastoff: uint64(0),
	}
}

// Add inserts a name with its inode and offset into the cache.
func (dc *Dcache) Add(name string, inum common.Inum, off uint64) {
	dc.cache[name] = Dentry{Inum: inum, Off: off}
}

// Lookup retrieves the cached entry for name.
func (dc *Dcache) Lookup(name string) (Dentry, bool) {
	d, ok := dc.cache[name]
	return d, ok
}

// Del removes a name from the cache and reports whether it was present.
func (dc *Dcache) Del(name string) bool {
	_, ok := dc.cache[name]
	if ok {
		delete(dc.cache, name)
	}
	return ok
}
