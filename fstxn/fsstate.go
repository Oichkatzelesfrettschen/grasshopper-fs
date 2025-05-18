package fstxn

import (
	"github.com/mit-pdos/go-journal/alloc"
	"github.com/mit-pdos/go-journal/common"
	"github.com/mit-pdos/go-journal/lockmap"
	"github.com/mit-pdos/go-journal/obj"
	"github.com/mit-pdos/go-nfsd/cache"
	"github.com/mit-pdos/go-nfsd/inode"
	"github.com/mit-pdos/go-nfsd/super"
)

const ICACHESZ uint64 = 100

type FsState struct {
	Super   *super.FsSuper
	Txn     *obj.Log
	Icache  *cache.Cache[*inode.Inode]
	Lockmap *lockmap.LockMap
	Balloc  *alloc.Alloc
	Ialloc  *alloc.Alloc
}

func readBitmap(super *super.FsSuper, start common.Bnum, len uint64) []byte {
	var bitmap []byte
	for i := uint64(0); i < len; i++ {
		blk := super.Disk.Read(uint64(start) + i)
		bitmap = append(bitmap, blk...)
	}
	return bitmap
}

func MkFsState(super *super.FsSuper, log *obj.Log) *FsState {
	balloc := alloc.MkAlloc(readBitmap(super, super.BitmapBlockStart(),
		super.NBlockBitmap))
	ialloc := alloc.MkAlloc(readBitmap(super, super.BitmapInodeStart(),
		super.NInodeBitmap))
	icache := cache.MkCache[*inode.Inode](ICACHESZ)
	st := &FsState{
		Super:   super,
		Txn:     log,
		Icache:  icache,
		Lockmap: lockmap.MkLockMap(),
		Balloc:  balloc,
		Ialloc:  ialloc,
	}
	return st
}
