package kvs

import (
	"fmt"

	"github.com/goose-lang/primitive/disk"
	"github.com/mit-pdos/go-journal/addr"
	"github.com/mit-pdos/go-journal/common"
	"github.com/mit-pdos/go-journal/jrnl"
	"github.com/mit-pdos/go-journal/obj"
	"github.com/mit-pdos/go-journal/util"
)

// Package kvs implements a very small key-value store using journaled transactions.

const DISKNAME string = "goose_kvs.img"

// KVS is a transactional key-value store backed by a disk log.
type KVS struct {
	sz  uint64
	log *obj.Log
}

// KVPair represents a key-value pair.
type KVPair struct {
	Key uint64
	Val []byte
}

// MkKVS initializes a new store on the provided disk.
func MkKVS(d disk.Disk, sz uint64) *KVS {
	/*if sz > d.Size() {
		panic("kvs larger than disk")
	}*/
	// XXX just need to assume that the kvs is less than the disk size?
	log := obj.MkLog(d)
	kvs := &KVS{
		sz:  sz,
		log: log,
	}
	return kvs
}

// MultiPut atomically writes multiple key/value pairs.
func (kvs *KVS) MultiPut(pairs []KVPair) bool {
	op := jrnl.Begin(kvs.log)
	for _, p := range pairs {
		if p.Key >= kvs.sz || p.Key < common.LOGSIZE {
			panic(fmt.Errorf("out-of-bounds put at %v", p.Key))
		}
		akey := addr.MkAddr(p.Key, 0)
		op.OverWrite(akey, common.NBITBLOCK, p.Val)
	}
	ok := op.CommitWait(true)
	return ok
}

// Get retrieves the value for a key.
func (kvs *KVS) Get(key uint64) (*KVPair, bool) {
	if key > kvs.sz || key < common.LOGSIZE {
		panic(fmt.Errorf("out-of-bounds get at %v", key))
	}
	op := jrnl.Begin(kvs.log)
	akey := addr.MkAddr(key, 0)
	data := util.CloneByteSlice(op.ReadBuf(akey, common.NBITBLOCK).Data)
	ok := op.CommitWait(true)
	return &KVPair{
		Key: key,
		Val: data,
	}, ok
}

// Delete releases all resources associated with the store.
func (kvs *KVS) Delete() {
	kvs.log.Shutdown()
}
