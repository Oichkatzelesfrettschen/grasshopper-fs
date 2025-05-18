package shrinker

import (
	"sync"

	"github.com/mit-pdos/go-journal/common"
	"github.com/mit-pdos/go-journal/util"
	"github.com/mit-pdos/go-nfsd/fstxn"
)

// ShrinkerSt manages background inode shrinking threads.
type ShrinkerSt struct {
	mu       *sync.Mutex
	condShut *sync.Cond
	nthread  uint32
	fsstate  *fstxn.FsState
	crash    bool
}

// MkShrinkerSt allocates a new shrinker state.
func MkShrinkerSt(st *fstxn.FsState) *ShrinkerSt {
	mu := new(sync.Mutex)
	shrinkst := &ShrinkerSt{
		mu:       mu,
		condShut: sync.NewCond(mu),
		nthread:  0,
		fsstate:  st,
		crash:    false,
	}
	return shrinkst
}

// crashed reports whether the shrinker has been crashed.
func (shrinkst *ShrinkerSt) crashed() bool {
	shrinkst.mu.Lock()
	crashed := shrinkst.crash
	shrinkst.mu.Unlock()
	return crashed
}

// If caller changes file size and shrinking is in progress (because
// an earlier call truncated the file), then help/wait with/for
// shrinking.  Also, called by shrinker.
// DoShrink performs the actual shrinking of the given inode.
func (shrinkst *ShrinkerSt) DoShrink(inum common.Inum) bool {
	var more = true
	var ok = true
	for more {
		op := fstxn.Begin(shrinkst.fsstate)
		ip := op.GetInodeInumFree(inum)
		if ip == nil {
			panic("shrink")
		}
		util.DPrintf(1, "%p: doShrink %v\n", op.Atxn.Id(), ip.Inum)
		more = ip.Shrink(op.Atxn)
		ok = op.Commit()
		if !ok {
			break
		}
		if shrinkst.crashed() {
			break
		}
	}
	return ok
}

// Shutdown waits for all shrinker threads to finish.
func (shrinker *ShrinkerSt) Shutdown() {
	shrinker.mu.Lock()
	for shrinker.nthread > 0 {
		util.DPrintf(1, "Shutdown: shrinker wait %d\n", shrinker.nthread)
		shrinker.condShut.Wait()
	}
	shrinker.mu.Unlock()
}

// Crash stops all shrinker threads without waiting for completion.
func (shrinker *ShrinkerSt) Crash() {
	shrinker.mu.Lock()
	shrinker.crash = true
	for shrinker.nthread > 0 {
		util.DPrintf(1, "Crash: wait %d\n", shrinker.nthread)
		shrinker.condShut.Wait()
	}
	shrinker.mu.Unlock()
}

// StartShrinker launches a goroutine to shrink a large file.
func (shrinkst *ShrinkerSt) StartShrinker(inum common.Inum) {
	util.DPrintf(1, "start shrink thread\n")
	shrinkst.mu.Lock()
	shrinkst.nthread = shrinkst.nthread + 1
	shrinkst.mu.Unlock()
	go func() { shrinkst.shrinker(inum) }()
}

// shrinker is the goroutine body that frees blocks.
func (shrinkst *ShrinkerSt) shrinker(inum common.Inum) {
	ok := shrinkst.DoShrink(inum)
	if !ok {
		panic("shrink")
	}
	util.DPrintf(1, "Shrinker: done shrinking # %d\n", inum)
	shrinkst.mu.Lock()
	shrinkst.nthread = shrinkst.nthread - 1
	shrinkst.condShut.Signal()
	shrinkst.mu.Unlock()
}
