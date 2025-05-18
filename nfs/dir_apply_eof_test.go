package nfs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mit-pdos/go-journal/common"
	"github.com/mit-pdos/go-nfsd/dir"
	"github.com/mit-pdos/go-nfsd/fstxn"
	"github.com/mit-pdos/go-nfsd/inode"
)

func TestDirApplyEOF(t *testing.T) {
	ts := newTest(t)
	defer ts.Close()

	ts.Create("a")
	ts.Create("b")
	ts.Create("c")

	op := fstxn.Begin(ts.clnt.srv.fsstate)
	dip := op.GetInodeInum(common.ROOTINUM)

	var last uint64
	eof := dir.Apply(dip, op, 0, 64, 1<<20, func(ip *inode.Inode, name string, inum common.Inum, off uint64) {
		last = off
	})
	op.Commit()

	assert.False(t, eof, "expected more entries after first call")

	op = fstxn.Begin(ts.clnt.srv.fsstate)
	eof2 := dir.Apply(dip, op, last, 1<<20, 1<<20, func(ip *inode.Inode, name string, inum common.Inum, off uint64) {
		last = off
	})
	op.Commit()

	assert.True(t, eof2, "expected EOF after second call")
}

func TestDirApplyEntsEOF(t *testing.T) {
	ts := newTest(t)
	defer ts.Close()

	ts.Create("a")
	ts.Create("b")
	ts.Create("c")

	op := fstxn.Begin(ts.clnt.srv.fsstate)
	dip := op.GetInodeInum(common.ROOTINUM)

	var last uint64
	eof := dir.ApplyEnts(dip, op, 0, 64, func(name string, inum common.Inum, off uint64) {
		last = off
	})
	op.Commit()

	assert.False(t, eof, "expected more entries after first call")

	op = fstxn.Begin(ts.clnt.srv.fsstate)
	eof2 := dir.ApplyEnts(dip, op, last, 1<<20, func(name string, inum common.Inum, off uint64) {
		last = off
	})
	op.Commit()

	assert.True(t, eof2, "expected EOF after second call")
}
